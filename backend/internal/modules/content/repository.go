package content

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
	"university-api/internal/platform/filegc"
)

// Repository — доступ к таблицам topic и material.
// Записи file не создаёт, но удаляет файлы, оставшиеся без ссылок (через filegc).
type Repository struct{ db *pgxpool.Pool }

// newRepository создаёт репозиторий поверх общего пула подключений (пул создаётся в main).
func newRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

// topicCols — колонки темы в том же порядке, в каком их читает scanTopic.
const topicCols = `t.id, t.course_id, t.title, t.description, t.position`

// scanTopic читает строку темы. Materials — пустой срез, а не nil,
// чтобы в JSON тема без материалов отдавалась как "materials": [], а не null.
func scanTopic(row pgx.Row) (Topic, error) {
	var t Topic
	err := row.Scan(&t.ID, &t.CourseID, &t.Title, &t.Description, &t.Position)
	t.Materials = []Material{}
	return t, err
}

// materialSelect — общий SELECT материала, порядок колонок совпадает со scanMaterial.
// JOIN topic — чтобы сразу получить course_id для проверки прав.
// LEFT JOIN file — у материала может не быть файла (только ссылка).
// Имя файла хранится в ключе, его извлекает SQL-выражение database.FileName.
// var, а не const: строка собирается с вызовом функции.
var materialSelect = `
	SELECT m.id, m.topic_id, t.course_id, m.title, m.file_id, ` + database.FileName("f") + `,
	       m.external_url, m.position, m.created_by, m.created_at
	FROM material m
	JOIN topic t ON t.id = m.topic_id
	LEFT JOIN file f ON f.id = m.file_id`

// scanMaterial читает строку, полученную через materialSelect.
func scanMaterial(row pgx.Row) (Material, error) {
	var m Material
	err := row.Scan(&m.ID, &m.TopicID, &m.CourseID, &m.Title, &m.FileID, &m.FileName,
		&m.ExternalURL, &m.Position, &m.CreatedBy, &m.CreatedAt)
	return m, err
}

// ListTopics возвращает темы курса с вложенными материалами (двумя запросами, без N+1).
// Первый запрос — все темы курса, второй — все материалы всех тем курса;
// затем материалы раскладываются по темам в памяти. Порядок — по position, при равенстве по id.
func (r *Repository) ListTopics(ctx context.Context, courseID int64) ([]Topic, error) {
	topics, err := database.Collect(ctx, r.db, scanTopic,
		`SELECT `+topicCols+` FROM topic t WHERE t.course_id = $1 ORDER BY t.position, t.id`, courseID)
	if err != nil {
		return nil, err
	}
	materials, err := database.Collect(ctx, r.db, scanMaterial,
		materialSelect+` WHERE t.course_id = $1 ORDER BY m.position, m.id`, courseID)
	if err != nil {
		return nil, err
	}

	// Индекс тем по id. Указатели на элементы среза безопасны:
	// в topics больше ничего не добавляется, срез не переаллоцируется.
	byID := make(map[int64]*Topic, len(topics))
	for i := range topics {
		byID[topics[i].ID] = &topics[i]
	}
	// Материалы уже отсортированы, поэтому внутри каждой темы порядок сохраняется.
	for _, m := range materials {
		if t := byID[m.TopicID]; t != nil {
			t.Materials = append(t.Materials, m)
		}
	}
	return topics, nil
}

// GetTopic возвращает тему по id (без материалов). Нет темы — 404 topic_not_found.
func (r *Repository) GetTopic(ctx context.Context, id int64) (Topic, error) {
	t, err := scanTopic(r.db.QueryRow(ctx, `SELECT `+topicCols+` FROM topic t WHERE t.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Topic{}, apperr.NewNotFound("topic_not_found")
	}
	return t, database.Translate(err)
}

// CreateTopic создаёт тему и возвращает её id.
// Если позиция не указана — ставит тему в конец: MAX(position) + 1 среди тем курса.
// position — ключ сортировки, а не уникальный номер: при параллельных вставках
// или явно заданной позиции возможны дубли, порядок тогда задаёт id.
// $4::int — явный тип нужен, потому что при nil Postgres не может вывести тип параметра.
func (r *Repository) CreateTopic(ctx context.Context, courseID int64, in TopicInput) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO topic (course_id, title, description, position)
		VALUES ($1, $2, $3, COALESCE($4::int, (SELECT COALESCE(MAX(position), 0) + 1 FROM topic WHERE course_id = $1)))
		RETURNING id`, courseID, in.Title, in.Description, in.Position).Scan(&id)
	return id, database.Translate(err)
}

// UpdateTopic полностью заменяет название и описание (description = nil стирает описание);
// позиция меняется, только если передана. Нет темы — 404 topic_not_found.
func (r *Repository) UpdateTopic(ctx context.Context, id int64, in TopicInput) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE topic SET title = $2, description = $3, position = COALESCE($4::int, position) WHERE id = $1`,
		id, in.Title, in.Description, in.Position)
	return database.Affected(tag, err, "topic_not_found")
}

// DeleteTopic удаляет тему вместе с материалами (ON DELETE CASCADE). Если на тему ссылаются
// задания — 409 in_use. Возвращает ключи файлов, оставшихся без ссылок, — их надо стереть с диска.
//
// Всё в одной транзакции: id файлов собираем до удаления, потому что каскад
// удалит материалы вместе со ссылками на файлы.
// При ошибке (в том числе при неудачном коммите) ключи не возвращаются.
func (r *Repository) DeleteTopic(ctx context.Context, id int64) ([]string, error) {
	var keys []string
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		fileIDs, err := database.Collect(ctx, tx, scanID,
			`SELECT file_id FROM material WHERE topic_id = $1 AND file_id IS NOT NULL`, id)
		if err != nil {
			return err
		}
		tag, err := tx.Exec(ctx, `DELETE FROM topic WHERE id = $1`, id)
		// AffectedDelete: 0 строк → 404, нарушение внешнего ключа (задания) → 409 in_use.
		if err := database.AffectedDelete(tag, err, "topic_not_found"); err != nil {
			return err
		}
		// Purge удаляет записи file, на которые больше никто не ссылается,
		// и возвращает их ключи. Файлы, используемые где-то ещё, не трогает.
		keys, err = filegc.Purge(ctx, tx, fileIDs)
		return err
	})
	if err != nil {
		// keys могли заполниться до неудачного коммита — после отката
		// записи file остались в БД, поэтому файлы с диска удалять нельзя.
		return nil, err
	}
	return keys, nil
}

// scanID читает одну колонку int64 — для запросов, возвращающих только id.
func scanID(row pgx.Row) (int64, error) {
	var id int64
	err := row.Scan(&id)
	return id, err
}

// GetMaterial возвращает материал по id (с course_id и именем файла). Нет — 404 material_not_found.
func (r *Repository) GetMaterial(ctx context.Context, id int64) (Material, error) {
	m, err := scanMaterial(r.db.QueryRow(ctx, materialSelect+` WHERE m.id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return Material{}, apperr.NewNotFound("material_not_found")
	}
	return m, database.Translate(err)
}

// CreateMaterial создаёт материал в теме и возвращает его id.
// Позиция по умолчанию — в конец темы (как в CreateTopic).
// createdBy — id преподавателя; колонка ссылается на teacher.
func (r *Repository) CreateMaterial(ctx context.Context, topicID, createdBy int64, in MaterialInput) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO material (topic_id, title, file_id, external_url, position, created_by)
		VALUES ($1, $2, $3, $4,
		        COALESCE($5::int, (SELECT COALESCE(MAX(position), 0) + 1 FROM material WHERE topic_id = $1)), $6)
		RETURNING id`, topicID, in.Title, in.FileID, in.ExternalURL, in.Position, createdBy).Scan(&id)
	return id, database.Translate(err)
}

// UpdateMaterial обновляет материал; если файл заменён или убран, старый стирается, когда на него
// никто больше не ссылается. Возвращает ключи файлов для удаления с диска.
// При ошибке (в том числе при неудачном коммите) ключи не возвращаются.
func (r *Repository) UpdateMaterial(ctx context.Context, id int64, in MaterialInput) ([]string, error) {
	var keys []string
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		// Запоминаем текущий файл и блокируем строку (FOR UPDATE), чтобы параллельное
		// обновление того же материала не прочитало тот же старый файл.
		var oldFile *int64
		err := tx.QueryRow(ctx, `SELECT file_id FROM material WHERE id = $1 FOR UPDATE`, id).Scan(&oldFile)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NewNotFound("material_not_found")
		}
		if err != nil {
			return database.Translate(err)
		}
		// Название, файл и ссылка заменяются полностью; позиция — только если передана.
		if _, err := tx.Exec(ctx, `
			UPDATE material SET title = $2, file_id = $3, external_url = $4, position = COALESCE($5::int, position)
			WHERE id = $1`, id, in.Title, in.FileID, in.ExternalURL, in.Position); err != nil {
			return database.Translate(err)
		}
		// Файл убрали или заменили — пробуем удалить старый.
		if oldFile != nil && (in.FileID == nil || *in.FileID != *oldFile) {
			keys, err = filegc.Purge(ctx, tx, []int64{*oldFile})
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err // транзакция откатилась — файлы с диска не трогаем
	}
	return keys, nil
}

// DeleteMaterial удаляет материал и, если его файл больше нигде не используется,
// запись этого файла. Возвращает ключи файлов для удаления с диска.
// При ошибке (в том числе при неудачном коммите) ключи не возвращаются.
func (r *Repository) DeleteMaterial(ctx context.Context, id int64) ([]string, error) {
	var keys []string
	err := database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		// RETURNING — удаляем и сразу узнаём, какой файл был привязан.
		var fileID *int64
		err := tx.QueryRow(ctx, `DELETE FROM material WHERE id = $1 RETURNING file_id`, id).Scan(&fileID)
		if errors.Is(err, pgx.ErrNoRows) {
			return apperr.NewNotFound("material_not_found")
		}
		if err != nil {
			return database.TranslateDelete(err)
		}
		if fileID != nil {
			keys, err = filegc.Purge(ctx, tx, []int64{*fileID})
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err // транзакция откатилась — файлы с диска не трогаем
	}
	return keys, nil
}
