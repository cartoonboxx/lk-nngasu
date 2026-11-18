package news

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
)

type Repository struct{ db *pgxpool.Pool }

func newRepository(db *pgxpool.Pool) *Repository { return &Repository{db: db} }

var (
	errNotFound    = apperr.NewNotFound("news_not_found")
	errArchived    = apperr.NewConflict("news_archived", "archived news cannot be edited; move it to draft or published first")
	errIsPublished = apperr.NewConflict("news_is_published", "unpublish or archive the news before deleting it")
)

var newsSelect = `
	SELECT n.id, n.body, n.status::text, n.published_at, n.created_at, n.updated_at,
	       n.author_id, ` + database.FullName("u") + `
	FROM news n
	JOIN app_user u ON u.id = n.author_id`

func scanNews(row pgx.Row) (AdminNews, error) {
	var n AdminNews
	err := row.Scan(&n.ID, &n.Body, &n.Status, &n.PublishedAt, &n.CreatedAt, &n.UpdatedAt,
		&n.Author.ID, &n.Author.Name)
	n.Excerpt = Excerpt(n.Body)
	return n, err
}

// ListQuery — параметры списков.
type ListQuery struct {
	Q             string
	Status        *string // только для админского списка
	Sort          string  // updated_at | -updated_at | published_at | -published_at (админский список)
	Public        bool    // только опубликованные с published_at <= now(), порядок ленты
	Limit, Offset int
}

func likeEscape(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// orderBy — сортировка админского списка; null published_at всегда в конце, затем id.
func orderBy(sort string) string {
	switch sort {
	case "updated_at":
		return "n.updated_at ASC, n.id ASC"
	case "published_at":
		return "n.published_at ASC NULLS LAST, n.id ASC"
	case "-published_at":
		return "n.published_at DESC NULLS LAST, n.id DESC"
	default: // -updated_at
		return "n.updated_at DESC, n.id DESC"
	}
}

func (r *Repository) List(ctx context.Context, q ListQuery) ([]AdminNews, int, error) {
	var w database.Filter
	order := orderBy(q.Sort)
	if q.Public {
		w.Raw("n.status = 'published' AND n.published_at <= now()")
		order = "n.published_at DESC, n.id DESC"
	}
	if q.Status != nil {
		w.Add("n.status = ?::news_status", *q.Status)
	}
	if q.Q != "" {
		w.Add("n.body ILIKE ?", "%"+likeEscape(q.Q)+"%")
	}
	var total int
	if err := r.db.QueryRow(ctx, `SELECT count(*) FROM news n`+w.Where(), w.Args()...).Scan(&total); err != nil {
		return nil, 0, database.Translate(err)
	}
	n := len(w.Args())
	args := append(w.Args(), q.Limit, q.Offset)
	list, err := database.Collect(ctx, r.db, scanNews,
		newsSelect+w.Where()+` ORDER BY `+order+fmt.Sprintf(` LIMIT $%d OFFSET $%d`, n+1, n+2), args...)
	return list, total, err
}

// Get — новость по id; public — только опубликованная (черновик и архив неотличимы от отсутствующей).
func (r *Repository) Get(ctx context.Context, id int64, public bool) (AdminNews, error) {
	where := ` WHERE n.id = $1`
	if public {
		where += ` AND n.status = 'published' AND n.published_at <= now()`
	}
	n, err := scanNews(r.db.QueryRow(ctx, newsSelect+where, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return AdminNews{}, errNotFound
	}
	return n, database.Translate(err)
}

// publishedAt — published_at после перехода из from в to: публикация ставит now(),
// уход из published очищает, published → published сохраняет дату.
const publishedAtExpr = `CASE WHEN $2::news_status = 'published'
	THEN CASE WHEN n.status = 'published' THEN n.published_at ELSE now() END END`

func (r *Repository) Create(ctx context.Context, authorID int64, body, status string) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		INSERT INTO news (author_id, body, status, published_at)
		VALUES ($1, $2, $3::news_status, CASE WHEN $3 = 'published' THEN now() END)
		RETURNING id`, authorID, body, status).Scan(&id)
	return id, database.Translate(err)
}

// Replace — полная замена текста и статуса (draft | published). Архивную новость
// менять нельзя (409 news_archived): сначала вернуть её из архива командой PATCH.
func (r *Repository) Replace(ctx context.Context, id int64, body, status string) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		cur, err := lockStatus(ctx, tx, id)
		if err != nil {
			return err
		}
		if cur == StatusArchived {
			return errArchived
		}
		_, err = tx.Exec(ctx, `
			UPDATE news n SET body = $3, status = $2::news_status, published_at = `+publishedAtExpr+`,
			       updated_at = now()
			WHERE n.id = $1`, id, status, body)
		return database.Translate(err)
	})
}

// SetStatus — смена состояния: любое из draft/published/archived в любое другое.
// Тот же статус — ничего не меняется (повтор команды безопасен).
func (r *Repository) SetStatus(ctx context.Context, id int64, status string) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		cur, err := lockStatus(ctx, tx, id)
		if err != nil || cur == status {
			return err
		}
		_, err = tx.Exec(ctx, `
			UPDATE news n SET status = $2::news_status, published_at = `+publishedAtExpr+`, updated_at = now()
			WHERE n.id = $1`, id, status)
		return database.Translate(err)
	})
}

// Delete — физическое удаление черновика или архивной новости; опубликованную — 409.
func (r *Repository) Delete(ctx context.Context, id int64) error {
	return database.InTx(ctx, r.db, func(tx pgx.Tx) error {
		cur, err := lockStatus(ctx, tx, id)
		if err != nil {
			return err
		}
		if cur == StatusPublished {
			return errIsPublished
		}
		_, err = tx.Exec(ctx, `DELETE FROM news WHERE id = $1`, id)
		return database.Translate(err)
	})
}

func lockStatus(ctx context.Context, tx pgx.Tx, id int64) (string, error) {
	var status string
	err := tx.QueryRow(ctx, `SELECT status::text FROM news WHERE id = $1 FOR UPDATE`, id).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", errNotFound
	}
	return status, database.Translate(err)
}
