package chat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
	"university-api/internal/platform/filegc"
)

type Repository struct{ db *pgxpool.Pool }

func newRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// ---- группы пользователя ----

// myGroups — группы, чаты которых положены пользователю: своя группа у студента и группы
// текущего семестра у преподавателя (если текущий семестр не выбран — все группы с нагрузкой).
const myGroupsCTE = `WITH my_groups AS (
	SELECT group_id FROM student WHERE user_id = $1::int
	UNION
	SELECT c.group_id
	FROM workload w
	JOIN course_offering c ON c.id = w.course_id
	JOIN semester s ON s.id = c.semester_id
	WHERE w.teacher_id = $1::int
	  AND (s.is_current OR NOT EXISTS (SELECT 1 FROM semester WHERE is_current))
)`

// EnsureGroupChats создаёт недостающие чаты групп пользователя и «точку отсчёта» прочитанного:
// история, написанная до первого захода, не считается непрочитанной.
func (r *Repository) EnsureGroupChats(ctx context.Context, userID int64) error {
	if _, err := r.db.Exec(ctx, myGroupsCTE+`
		INSERT INTO chat (kind, group_id)
		SELECT 'group', group_id FROM my_groups
		ON CONFLICT (group_id) WHERE kind = 'group' DO NOTHING`, userID); err != nil {
		return database.Translate(err)
	}
	_, err := r.db.Exec(ctx, myGroupsCTE+`
		INSERT INTO chat_read (chat_id, user_id, last_read_message_id)
		SELECT ch.id, $1::int, COALESCE((SELECT MAX(m.id) FROM chat_message m WHERE m.chat_id = ch.id), 0)
		FROM chat ch
		WHERE ch.kind = 'group' AND ch.group_id IN (SELECT group_id FROM my_groups)
		ON CONFLICT DO NOTHING`, userID)
	return database.Translate(err)
}

// CanAccessGroup — пользователь учится в группе или ведёт у неё курс (в любом семестре).
func (r *Repository) CanAccessGroup(ctx context.Context, userID, groupID int64) (bool, error) {
	var ok bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM student WHERE user_id = $1 AND group_id = $2)
		    OR EXISTS (SELECT 1 FROM workload w JOIN course_offering c ON c.id = w.course_id
		               WHERE w.teacher_id = $1 AND c.group_id = $2)`, userID, groupID).Scan(&ok)
	return ok, database.Translate(err)
}

// TeachesGroup — преподаватель ведёт хотя бы один курс у группы.
func (r *Repository) TeachesGroup(ctx context.Context, teacherID, groupID int64) (bool, error) {
	var ok bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS (SELECT 1 FROM workload w JOIN course_offering c ON c.id = w.course_id
		               WHERE w.teacher_id = $1 AND c.group_id = $2)`, teacherID, groupID).Scan(&ok)
	return ok, database.Translate(err)
}

func (r *Repository) UserInfo(ctx context.Context, id int64) (userInfo, error) {
	var u userInfo
	err := r.db.QueryRow(ctx, `
		SELECT u.id, `+database.FullName("u")+`, u.is_active, u.is_admin, t.user_id IS NOT NULL, s.group_id
		FROM app_user u
		LEFT JOIN teacher t ON t.user_id = u.id
		LEFT JOIN student s ON s.user_id = u.id
		WHERE u.id = $1`, id).Scan(&u.ID, &u.Name, &u.Active, &u.Admin, &u.Teacher, &u.GroupID)
	if errors.Is(err, pgx.ErrNoRows) {
		return userInfo{}, apperr.NewNotFound("user_not_found")
	}
	return u, database.Translate(err)
}

// ---- чаты ----

func (r *Repository) GetChat(ctx context.Context, id int64) (chatRow, error) {
	var c chatRow
	err := r.db.QueryRow(ctx,
		`SELECT id, kind::text, group_id, user_low, user_high FROM chat WHERE id = $1`, id).
		Scan(&c.ID, &c.Kind, &c.GroupID, &c.UserLow, &c.UserHigh)
	if errors.Is(err, pgx.ErrNoRows) {
		return chatRow{}, apperr.NewNotFound("chat_not_found")
	}
	return c, database.Translate(err)
}

// OpenDirect находит или создаёт личный чат пары пользователей; created — чат новый.
func (r *Repository) OpenDirect(ctx context.Context, a, b int64) (id int64, created bool, err error) {
	low, high := pair(a, b)
	tag, err := r.db.Exec(ctx, `
		INSERT INTO chat (kind, user_low, user_high) VALUES ('direct', $1, $2)
		ON CONFLICT (user_low, user_high) WHERE kind = 'direct' DO NOTHING`, low, high)
	if err != nil {
		return 0, false, database.Translate(err)
	}
	err = r.db.QueryRow(ctx,
		`SELECT id FROM chat WHERE kind = 'direct' AND user_low = $1 AND user_high = $2`, low, high).Scan(&id)
	return id, tag.RowsAffected() == 1, database.Translate(err)
}

// OpenGroup находит или создаёт чат группы и запоминает, что userID «в нём».
func (r *Repository) OpenGroup(ctx context.Context, groupID, userID int64) (int64, error) {
	if _, err := r.db.Exec(ctx, `
		INSERT INTO chat (kind, group_id) VALUES ('group', $1)
		ON CONFLICT (group_id) WHERE kind = 'group' DO NOTHING`, groupID); err != nil {
		if database.ConstraintName(err) == "chat_group_id_fkey" {
			return 0, apperr.NewNotFound("group_not_found")
		}
		return 0, database.Translate(err)
	}
	var id int64
	if err := r.db.QueryRow(ctx,
		`SELECT id FROM chat WHERE kind = 'group' AND group_id = $1`, groupID).Scan(&id); err != nil {
		return 0, database.Translate(err)
	}
	_, err := r.db.Exec(ctx, `
		INSERT INTO chat_read (chat_id, user_id, last_read_message_id)
		VALUES ($1, $2, COALESCE((SELECT MAX(m.id) FROM chat_message m WHERE m.chat_id = $1), 0))
		ON CONFLICT DO NOTHING`, id, userID)
	return id, database.Translate(err)
}

// chatColumns/chatFrom — чат глазами зрителя $1. extra — дополнительные колонки в конце
// (count(*) OVER () для страницы списка).
func chatQuery(extra string) string {
	return `
	SELECT ch.id, ch.kind::text, ch.group_id, g.name, ch.created_at,
	       peer.id, ` + database.FullName("peer") + `,
	       lm.id, lm.sender_id, ` + database.FullName("ls") + `,
	       lm.body, lm.created_at,
	       EXISTS (SELECT 1 FROM chat_message_file mf WHERE mf.message_id = lm.id)` + extra + `
	FROM chat ch
	LEFT JOIN study_group g ON g.id = ch.group_id
	LEFT JOIN app_user peer ON ch.kind = 'direct'
	                       AND peer.id = CASE WHEN ch.user_low = $1 THEN ch.user_high ELSE ch.user_low END
	LEFT JOIN chat_read cr ON cr.chat_id = ch.id AND cr.user_id = $1
	LEFT JOIN LATERAL (
		SELECT m.id, m.sender_id, m.body, m.created_at
		FROM chat_message m WHERE m.chat_id = ch.id AND m.deleted_at IS NULL ORDER BY m.id DESC LIMIT 1
	) lm ON true
	LEFT JOIN app_user ls ON ls.id = lm.sender_id`
}

func scanChat(row pgx.Row, extra ...any) (Chat, error) {
	var (
		c        Chat
		groupNm  *string
		peerID   *int64
		peerName *string
		lmID     *int64
		lmSender *int64
		lmName   *string
		lmBody   *string
		lmAt     *time.Time
		lmFiles  bool
	)
	dest := append([]any{&c.ID, &c.Kind, &c.GroupID, &groupNm, &c.CreatedAt, &peerID, &peerName,
		&lmID, &lmSender, &lmName, &lmBody, &lmAt, &lmFiles}, extra...)
	if err := row.Scan(dest...); err != nil {
		return Chat{}, err
	}
	switch {
	case c.Kind == KindGroup && groupNm != nil:
		c.Title = *groupNm
	case peerID != nil && peerName != nil:
		c.Peer = &Person{ID: *peerID, Name: *peerName}
		c.Title = *peerName
	}
	if lmID != nil && lmSender != nil && lmName != nil && lmBody != nil && lmAt != nil {
		c.LastMessage = &Preview{ID: *lmID, SenderID: *lmSender, SenderName: *lmName, Body: *lmBody,
			HasFiles: lmFiles, CreatedAt: *lmAt}
	}
	return c, nil
}

// ListChats — страница чатов пользователя: его личные чаты и чаты его групп; у администратора
// вместо «своих групп» — групповые чаты, которые он открывал (chat_read). Чужие личные чаты
// администратору не видны. q — подстрока названия (группа или имя собеседника).
// Порядок: свежесть последнего сообщения, чаты без сообщений — в конце, затем id DESC.
func (r *Repository) ListChats(ctx context.Context, userID int64, isAdmin bool, q ListQuery) ([]Chat, int, error) {
	sql := myGroupsCTE + chatQuery(", count(*) OVER ()") + `
		WHERE ((ch.kind = 'direct' AND $1::int IN (ch.user_low, ch.user_high))
		    OR (ch.kind = 'group' AND (ch.group_id IN (SELECT group_id FROM my_groups)
		                               OR ($2::bool AND cr.chat_id IS NOT NULL))))
		  AND ($3::text = '' OR COALESCE(g.name, ` + database.FullName("peer") + `) ILIKE '%' || $3::text || '%')
		ORDER BY lm.created_at DESC NULLS LAST, ch.id DESC` + pageTail(4)
	total := 0
	list, err := database.Collect(ctx, r.db, func(row pgx.Row) (Chat, error) {
		return scanChat(row, &total)
	}, sql, userID, isAdmin, likeEscape(q.Q), q.Limit, q.Offset)
	if err != nil {
		return nil, 0, err
	}
	if len(list) == 0 && q.Offset > 0 {
		total, err = r.countPast(ctx, sql, userID, isAdmin, likeEscape(q.Q))
	}
	return list, total, err
}

func (r *Repository) ChatView(ctx context.Context, chatID, userID int64) (Chat, error) {
	c, err := scanChat(r.db.QueryRow(ctx, chatQuery("")+` WHERE ch.id = $2`, userID, chatID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Chat{}, apperr.NewNotFound("chat_not_found")
	}
	return c, database.Translate(err)
}

// ---- контакты ----

// ListContacts — активные пользователи, с которыми viewer может начать личный чат
// (те же правила, что decideContact), кроме него самого. q — подстрока ФИО.
func (r *Repository) ListContacts(ctx context.Context, viewerID int64, q ListQuery) ([]Contact, int, error) {
	const teaches = `EXISTS (SELECT 1 FROM workload w JOIN course_offering c ON c.id = w.course_id
	                         WHERE w.teacher_id = %s AND c.group_id = %s)`
	sql := `
	WITH me AS (
		SELECT u.id, u.is_admin, t.user_id IS NOT NULL AS teacher, s.group_id
		FROM app_user u
		LEFT JOIN teacher t ON t.user_id = u.id
		LEFT JOIN student s ON s.user_id = u.id
		WHERE u.id = $1
	)
	SELECT u.id, ` + database.FullName("u") + `, u.avatar_file_id,
	       CASE WHEN t.user_id IS NOT NULL THEN d.name ELSE g.name END,
	       count(*) OVER ()
	FROM me, app_user u
	LEFT JOIN teacher t ON t.user_id = u.id
	LEFT JOIN department d ON d.id = t.department_id
	LEFT JOIN student s ON s.user_id = u.id
	LEFT JOIN study_group g ON g.id = s.group_id
	WHERE u.is_active AND u.id <> me.id
	  AND (me.is_admin OR u.is_admin
	       OR (me.teacher AND t.user_id IS NOT NULL)
	       OR (me.teacher AND s.group_id IS NOT NULL AND ` + fmt.Sprintf(teaches, "me.id", "s.group_id") + `)
	       OR (t.user_id IS NOT NULL AND me.group_id IS NOT NULL AND ` + fmt.Sprintf(teaches, "u.id", "me.group_id") + `)
	       OR (NOT me.teacher AND t.user_id IS NULL AND me.group_id IS NOT NULL AND s.group_id = me.group_id))
	  AND ($2::text = '' OR ` + database.FullName("u") + ` ILIKE '%' || $2::text || '%'
	                     OR (u.first_name || ' ' || u.last_name) ILIKE '%' || $2::text || '%')
	ORDER BY u.last_name, u.first_name, u.id` + pageTail(3)
	total := 0
	list, err := database.Collect(ctx, r.db, func(row pgx.Row) (Contact, error) {
		var (
			c      Contact
			avatar *int64
		)
		if err := row.Scan(&c.ID, &c.Name, &avatar, &c.Description, &total); err != nil {
			return Contact{}, err
		}
		if avatar != nil {
			c.Avatar = &Avatar{ID: *avatar, URL: "/api/v1/files/" + strconv.FormatInt(*avatar, 10)}
		}
		return c, nil
	}, sql, viewerID, likeEscape(q.Q), q.Limit, q.Offset)
	if err == nil && len(list) == 0 && q.Offset > 0 {
		total, err = r.countPast(ctx, sql, viewerID, likeEscape(q.Q))
	}
	return list, total, err
}

// pageTail — LIMIT/OFFSET с номерами параметров n и n+1. Отделён маркером, чтобы countPast
// мог посчитать total той же выборки без страницы.
func pageTail(n int) string {
	return fmt.Sprintf("\n\t-- page\n\tLIMIT $%d OFFSET $%d", n, n+1)
}

// countPast — total, когда offset ушёл за конец выборки (страница пуста и count(*) OVER () не пришёл).
func (r *Repository) countPast(ctx context.Context, sql string, args ...any) (int, error) {
	base := sql[:strings.LastIndex(sql, "\n\t-- page")]
	var n int
	err := r.db.QueryRow(ctx, `SELECT count(*) FROM (`+base+`) x`, args...).Scan(&n)
	return n, database.Translate(err)
}

func likeEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// ---- сообщения ----

func filesJSON() string {
	return `(SELECT COALESCE(json_agg(json_build_object('id', f.id, 'name', ` + database.FileName("f") + `) ORDER BY f.id), '[]'::json)
		FROM chat_message_file l JOIN file f ON f.id = l.file_id WHERE l.message_id = m.id)`
}

var messageSelect = `
	SELECT m.id, m.chat_id, m.sender_id, ` + database.FullName("u") + `, m.body, m.created_at, m.edited_at,
	       ` + filesJSON() + `
	FROM chat_message m JOIN app_user u ON u.id = m.sender_id`

func scanMessage(row pgx.Row) (Message, error) {
	var (
		m   Message
		raw []byte
	)
	if err := row.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.SenderName, &m.Body,
		&m.CreatedAt, &m.EditedAt, &raw); err != nil {
		return Message{}, err
	}
	m.Files = []FileRef{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &m.Files); err != nil {
			return Message{}, err
		}
	}
	return m, nil
}

// ListMessages — до limit сообщений старше before (или самые свежие), по возрастанию id.
// Берёт на одно больше, чтобы узнать, есть ли ещё более старые.
func (r *Repository) ListMessages(ctx context.Context, chatID int64, before *int64, limit int) (MessagePage, error) {
	list, err := database.Collect(ctx, r.db, scanMessage,
		messageSelect+` WHERE m.chat_id = $1 AND m.deleted_at IS NULL AND ($2::bigint IS NULL OR m.id < $2::bigint) ORDER BY m.id DESC LIMIT $3`,
		chatID, before, limit+1)
	if err != nil {
		return MessagePage{}, err
	}
	page := MessagePage{HasMore: len(list) > limit}
	if page.HasMore {
		list = list[:limit]
	}
	slices.Reverse(list)
	if list == nil {
		list = []Message{}
	}
	page.Items = list
	if page.HasMore {
		page.NextBefore = &list[0].ID
	}
	return page, nil
}

func (r *Repository) GetMessage(ctx context.Context, id int64) (Message, error) {
	m, err := scanMessage(r.db.QueryRow(ctx, messageSelect+` WHERE m.id = $1 AND m.deleted_at IS NULL`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Message{}, apperr.NewNotFound("message_not_found")
	}
	return m, database.Translate(err)
}

const markReadSQL = `
	INSERT INTO chat_read (chat_id, user_id, last_read_message_id) VALUES ($1, $2, $3)
	ON CONFLICT (chat_id, user_id)
	DO UPDATE SET last_read_message_id = GREATEST(chat_read.last_read_message_id, EXCLUDED.last_read_message_id)`

// CreateMessage сохраняет сообщение с вложениями; отправитель по определению прочитал чат до него.
func (r *Repository) CreateMessage(ctx context.Context, chatID, senderID int64, body string, fileIDs []int64) (int64, error) {
	var id int64
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx,
			`INSERT INTO chat_message (chat_id, sender_id, body) VALUES ($1, $2, $3) RETURNING id`,
			chatID, senderID, body).Scan(&id); err != nil {
			return database.Translate(err)
		}
		if len(fileIDs) > 0 {
			if _, err := tx.Exec(ctx,
				`INSERT INTO chat_message_file (message_id, file_id) SELECT $1::bigint, unnest($2::bigint[])::int`,
				id, fileIDs); err != nil {
				if database.ConstraintName(err) == "chat_message_file_file_id_fkey" {
					return apperr.NewInvalid("file_not_found", "attached file does not exist")
				}
				return database.Translate(err)
			}
		}
		_, err := tx.Exec(ctx, markReadSQL, chatID, senderID, id)
		return database.Translate(err)
	})
	return id, err
}

func (r *Repository) UpdateMessageBody(ctx context.Context, id int64, body string) error {
	tag, err := r.db.Exec(ctx,
		`UPDATE chat_message SET body = $2, edited_at = now() WHERE id = $1 AND deleted_at IS NULL`, id, body)
	return database.Affected(tag, err, "message_not_found")
}

func scanID(row pgx.Row) (int64, error) {
	var id int64
	err := row.Scan(&id)
	return id, err
}

// DeleteMessage удаляет сообщение и связи с вложениями (без tombstone). Возвращает ключи
// файлов, оставшихся без ссылок, — их нужно удалить с диска после коммита; файл, который
// используется ещё где-то, остаётся.
func (r *Repository) DeleteMessage(ctx context.Context, id int64) ([]string, error) {
	var keys []string
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		fileIDs, err := database.Collect(ctx, tx, scanID,
			`DELETE FROM chat_message_file WHERE message_id = $1 RETURNING file_id`, id)
		if err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `DELETE FROM chat_message WHERE id = $1`, id)
		if err := database.Affected(tag, err, "message_not_found"); err != nil {
			return err
		}
		keys, err = filegc.Purge(ctx, tx, fileIDs)
		return err
	})
	return keys, err
}

// Members — кому доставлять события чата сейчас: оба собеседника личного чата; для группы —
// студенты группы, преподаватели с нагрузкой у неё (в любом семестре, как CanAccessGroup)
// и администраторы, открывавшие этот чат. Только активные пользователи.
func (r *Repository) Members(ctx context.Context, chatID int64) ([]int64, error) {
	return database.Collect(ctx, r.db, scanID, `
		WITH m AS (
			SELECT unnest(ARRAY[ch.user_low, ch.user_high]) AS user_id
			FROM chat ch WHERE ch.id = $1 AND ch.kind = 'direct'
			UNION
			SELECT s.user_id FROM chat ch JOIN student s ON s.group_id = ch.group_id
			WHERE ch.id = $1 AND ch.kind = 'group'
			UNION
			SELECT w.teacher_id FROM chat ch
			JOIN course_offering c ON c.group_id = ch.group_id
			JOIN workload w ON w.course_id = c.id
			WHERE ch.id = $1 AND ch.kind = 'group'
			UNION
			SELECT cr.user_id FROM chat_read cr JOIN app_user a ON a.id = cr.user_id AND a.is_admin
			WHERE cr.chat_id = $1
		)
		SELECT m.user_id::bigint FROM m JOIN app_user u ON u.id = m.user_id AND u.is_active`, chatID)
}

func (r *Repository) Participants(ctx context.Context, chatID int64) ([]Person, error) {
	ids, err := r.Members(ctx, chatID)
	if err != nil {
		return nil, err
	}
	return database.Collect(ctx, r.db, func(row pgx.Row) (Person, error) {
		var person Person
		err := row.Scan(&person.ID, &person.Name)
		return person, err
	}, `SELECT u.id, `+database.FullName("u")+`
		FROM app_user u WHERE u.id = ANY($1::bigint[])
		ORDER BY u.last_name, u.first_name, u.id`, ids)
}
