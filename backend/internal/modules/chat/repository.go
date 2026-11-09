package chat

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
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

// OpenDirect находит или создаёт личный чат пары пользователей.
func (r *Repository) OpenDirect(ctx context.Context, a, b int64) (int64, error) {
	low, high := pair(a, b)
	if _, err := r.db.Exec(ctx, `
		INSERT INTO chat (kind, user_low, user_high) VALUES ('direct', $1, $2)
		ON CONFLICT (user_low, user_high) WHERE kind = 'direct' DO NOTHING`, low, high); err != nil {
		return 0, database.Translate(err)
	}
	var id int64
	err := r.db.QueryRow(ctx,
		`SELECT id FROM chat WHERE kind = 'direct' AND user_low = $1 AND user_high = $2`, low, high).Scan(&id)
	return id, database.Translate(err)
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

// $1 — зритель. Личный чат без сообщений в список не попадает (его видно только создателю по ответу POST).
var chatSelect = `
	SELECT ch.id, ch.kind::text, ch.group_id, g.name,
	       peer.id, ` + database.FullName("peer") + `,
	       lm.id, lm.sender_id, ` + database.FullName("ls") + `,
	       lm.body, lm.created_at, lm.deleted_at IS NOT NULL,
	       EXISTS (SELECT 1 FROM chat_message_file mf WHERE mf.message_id = lm.id),
	       (SELECT count(*) FROM chat_message m
	         WHERE m.chat_id = ch.id AND m.id > COALESCE(cr.last_read_message_id, 0)
	           AND m.sender_id <> $1 AND m.deleted_at IS NULL)
	FROM chat ch
	LEFT JOIN study_group g ON g.id = ch.group_id
	LEFT JOIN app_user peer ON ch.kind = 'direct'
	                       AND peer.id = CASE WHEN ch.user_low = $1 THEN ch.user_high ELSE ch.user_low END
	LEFT JOIN chat_read cr ON cr.chat_id = ch.id AND cr.user_id = $1
	LEFT JOIN LATERAL (
		SELECT m.id, m.sender_id, m.body, m.created_at, m.deleted_at
		FROM chat_message m WHERE m.chat_id = ch.id ORDER BY m.id DESC LIMIT 1
	) lm ON true
	LEFT JOIN app_user ls ON ls.id = lm.sender_id`

func scanChat(row pgx.Row) (Chat, error) {
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
		lmDel    bool
		lmFiles  bool
	)
	if err := row.Scan(&c.ID, &c.Kind, &c.GroupID, &groupNm, &peerID, &peerName,
		&lmID, &lmSender, &lmName, &lmBody, &lmAt, &lmDel, &lmFiles, &c.Unread); err != nil {
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
			HasFiles: lmFiles, Deleted: lmDel, CreatedAt: *lmAt}
	}
	return c, nil
}

// ListChats — чаты пользователя: личные с перепиской и чаты его групп (по свежести).
// Для администратора групповые чаты — только те, что он открывал сам (chat_read).
func (r *Repository) ListChats(ctx context.Context, userID int64, isAdmin bool) ([]Chat, error) {
	q := myGroupsCTE + chatSelect + `
		WHERE (ch.kind = 'direct' AND $1::int IN (ch.user_low, ch.user_high) AND lm.id IS NOT NULL)
		   OR (ch.kind = 'group' AND (ch.group_id IN (SELECT group_id FROM my_groups)
		                              OR ($2::bool AND cr.chat_id IS NOT NULL)))
		ORDER BY COALESCE(lm.created_at, ch.created_at) DESC, ch.id DESC`
	return database.Collect(ctx, r.db, scanChat, q, userID, isAdmin)
}

func (r *Repository) ChatView(ctx context.Context, chatID, userID int64) (Chat, error) {
	c, err := scanChat(r.db.QueryRow(ctx, chatSelect+` WHERE ch.id = $2`, userID, chatID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Chat{}, apperr.NewNotFound("chat_not_found")
	}
	return c, database.Translate(err)
}

// ---- сообщения ----

func filesJSON() string {
	return `(SELECT COALESCE(json_agg(json_build_object('id', f.id, 'name', ` + database.FileName("f") + `) ORDER BY f.id), '[]'::json)
		FROM chat_message_file l JOIN file f ON f.id = l.file_id WHERE l.message_id = m.id)`
}

var messageSelect = `
	SELECT m.id, m.chat_id, m.sender_id, ` + database.FullName("u") + `, m.body, m.created_at, m.edited_at,
	       m.deleted_at IS NOT NULL, ` + filesJSON() + `
	FROM chat_message m JOIN app_user u ON u.id = m.sender_id`

func scanMessage(row pgx.Row) (Message, error) {
	var (
		m   Message
		raw []byte
	)
	if err := row.Scan(&m.ID, &m.ChatID, &m.SenderID, &m.SenderName, &m.Body,
		&m.CreatedAt, &m.EditedAt, &m.Deleted, &raw); err != nil {
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

// ListMessages: after — сообщения новее id (по возрастанию, для опроса); иначе последние limit
// сообщений до before (или самые свежие) — тоже по возрастанию.
func (r *Repository) ListMessages(ctx context.Context, chatID int64, before, after *int64, limit int) ([]Message, error) {
	if after != nil {
		return database.Collect(ctx, r.db, scanMessage,
			messageSelect+` WHERE m.chat_id = $1 AND m.id > $2 ORDER BY m.id LIMIT $3`, chatID, *after, limit)
	}
	list, err := database.Collect(ctx, r.db, scanMessage,
		messageSelect+` WHERE m.chat_id = $1 AND ($2::bigint IS NULL OR m.id < $2::bigint) ORDER BY m.id DESC LIMIT $3`,
		chatID, before, limit)
	if err != nil {
		return nil, err
	}
	slices.Reverse(list)
	return list, nil
}

func (r *Repository) GetMessage(ctx context.Context, id int64) (Message, error) {
	m, err := scanMessage(r.db.QueryRow(ctx, messageSelect+` WHERE m.id = $1`, id))
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

// DeleteMessage помечает сообщение удалённым, стирает текст и вложения. Возвращает ключи файлов,
// оставшихся без ссылок, — их нужно удалить с диска после коммита.
func (r *Repository) DeleteMessage(ctx context.Context, id int64) ([]string, error) {
	var keys []string
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx,
			`UPDATE chat_message SET body = '', deleted_at = now() WHERE id = $1 AND deleted_at IS NULL`, id)
		if err := database.Affected(tag, err, "message_not_found"); err != nil {
			return err
		}
		fileIDs, err := database.Collect(ctx, tx, scanID,
			`DELETE FROM chat_message_file WHERE message_id = $1 RETURNING file_id`, id)
		if err != nil {
			return err
		}
		keys, err = filegc.Purge(ctx, tx, fileIDs)
		return err
	})
	return keys, err
}

// MarkRead сдвигает «прочитано до» вперёд (назад не двигает) и не выходит за последнее сообщение чата.
func (r *Repository) MarkRead(ctx context.Context, chatID, userID, messageID int64) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO chat_read (chat_id, user_id, last_read_message_id)
		VALUES ($1, $2, LEAST($3::bigint, COALESCE((SELECT MAX(m.id) FROM chat_message m WHERE m.chat_id = $1), 0)))
		ON CONFLICT (chat_id, user_id)
		DO UPDATE SET last_read_message_id = GREATEST(chat_read.last_read_message_id, EXCLUDED.last_read_message_id)`,
		chatID, userID, messageID)
	return database.Translate(err)
}
