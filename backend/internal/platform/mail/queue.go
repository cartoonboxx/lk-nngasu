package mail

import (
	"context"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Очередь писем (таблица mail_outbox, миграция 007).
//
// Письмо записывается в БД в той же транзакции, что и действие, которое его вызвало
// (создание пользователя, повторное приглашение). Поэтому либо есть и пользователь,
// и письмо в очереди, либо нет ничего: сбой почтового сервера не оставляет учётку
// без приглашения. Отправляет письма фоновый Worker с повторными попытками.

// Enqueue ставит письмо в очередь внутри переданной транзакции.
func Enqueue(ctx context.Context, tx pgx.Tx, to, subject, body string) error {
	_, err := tx.Exec(ctx, `INSERT INTO mail_outbox (recipient, subject, body) VALUES ($1, $2, $3)`, to, subject, body)
	return err
}

// Параметры повторов: после maxAttempts неудачных попыток письмо помечается
// как не отправленное и больше не трогается (видно по failed_at и last_error).
const (
	maxAttempts  = 8
	pollInterval = 5 * time.Second
	batchSize    = 20
)

// Worker забирает письма из очереди и отправляет через Sender.
type Worker struct {
	db     *pgxpool.Pool
	sender Sender
}

func NewWorker(db *pgxpool.Pool, sender Sender) *Worker { return &Worker{db: db, sender: sender} }

// Run отправляет письма, пока не отменён ctx. Запускать в отдельной горутине.
func (w *Worker) Run(ctx context.Context) {
	t := time.NewTicker(pollInterval)
	defer t.Stop()
	for {
		for w.sendBatch(ctx) == batchSize { // полная пачка — возможно, в очереди есть ещё
		}
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}
	}
}

// sendBatch отправляет до batchSize готовых писем и возвращает, сколько обработано.
//
// FOR UPDATE SKIP LOCKED: если запущено несколько экземпляров сервера,
// каждое письмо возьмёт только один из них, и оно не уйдёт дважды.
func (w *Worker) sendBatch(ctx context.Context) int {
	n := 0
	err := pgx.BeginFunc(ctx, w.db, func(tx pgx.Tx) error {
		rows, err := tx.Query(ctx, `
			SELECT id, recipient, subject, body, attempts FROM mail_outbox
			WHERE sent_at IS NULL AND failed_at IS NULL AND next_attempt_at <= now()
			ORDER BY id LIMIT $1
			FOR UPDATE SKIP LOCKED`, batchSize)
		if err != nil {
			return err
		}
		type msg struct {
			id                int64
			to, subject, body string
			attempts          int
		}
		var batch []msg
		for rows.Next() {
			var m msg
			if err := rows.Scan(&m.id, &m.to, &m.subject, &m.body, &m.attempts); err != nil {
				rows.Close()
				return err
			}
			batch = append(batch, m)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}

		for _, m := range batch {
			n++
			if sendErr := w.sender.Send(m.to, m.subject, m.body); sendErr != nil {
				attempts := m.attempts + 1
				slog.Error("email send failed", "outbox_id", m.id, "attempt", attempts, "err", sendErr)
				if attempts >= maxAttempts {
					_, err = tx.Exec(ctx, `UPDATE mail_outbox SET attempts = $2, last_error = $3, failed_at = now() WHERE id = $1`,
						m.id, attempts, sendErr.Error())
				} else {
					// Экспоненциальная пауза: 1, 2, 4 … минут, не больше часа.
					delay := min(time.Minute<<(attempts-1), time.Hour)
					_, err = tx.Exec(ctx, `UPDATE mail_outbox SET attempts = $2, last_error = $3, next_attempt_at = now() + $4::interval WHERE id = $1`,
						m.id, attempts, sendErr.Error(), delay.String())
				}
			} else {
				// Тело стираем: в нём одноразовая ссылка, хранить её после отправки незачем.
				_, err = tx.Exec(ctx, `UPDATE mail_outbox SET sent_at = now(), body = '' WHERE id = $1`, m.id)
			}
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil && ctx.Err() == nil {
		slog.Error("mail queue", "err", err)
	}
	return n
}
