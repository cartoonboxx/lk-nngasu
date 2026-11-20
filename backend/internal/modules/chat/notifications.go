package chat

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
)

type messageNotification struct {
	MessageID         int64  `json:"message_id,omitempty"`
	UserID            int64  `json:"user_id,omitempty"`
	ChatID            int64  `json:"chat_id,omitempty"`
	LastReadMessageID int64  `json:"last_read_message_id,omitempty"`
	InstanceID        string `json:"instance_id"`
}

// У каждого API свой WebSocket-хаб. Уведомления БД доставляют сообщения
// пользователям, подключённым к другому экземпляру API с той же базой.
func (s *Service) listenMessages(ctx context.Context) {
	delay := time.Second
	for ctx.Err() == nil {
		err := s.receiveNotifications(ctx)
		if ctx.Err() != nil {
			return
		}
		slog.Error("chat: notification listener disconnected", "err", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}
		delay = min(delay*2, 30*time.Second)
	}
}

func (s *Service) receiveNotifications(ctx context.Context) error {
	config := s.repo.db.Config().ConnConfig.Copy()
	config.ConnectTimeout = 10 * time.Second
	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return err
	}
	defer func() {
		closeCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = conn.Close(closeCtx)
	}()
	if _, err := conn.Exec(ctx, `LISTEN lk_chat_messages`); err != nil {
		return err
	}
	slog.Info("chat: notification listener connected")
	// При обрыве LISTEN уведомления не накапливаются: перечитываем снимки из БД.
	s.resync()
	for {
		notification, err := conn.WaitForNotification(ctx)
		if err != nil {
			return err
		}
		var event messageNotification
		if err := json.Unmarshal([]byte(notification.Payload), &event); err != nil || (event.MessageID <= 0 && (event.UserID <= 0 || event.ChatID <= 0 || event.LastReadMessageID <= 0)) {
			slog.Warn("chat: invalid message notification")
			continue
		}
		// Локальные события уже рассылаются обработчиками.
		if event.InstanceID == s.instanceID {
			continue
		}
		if event.UserID > 0 {
			s.publish([]int64{event.UserID}, "chat.read", ReadPosition{ChatID: event.ChatID, LastReadMessageID: event.LastReadMessageID})
			continue
		}
		queryCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		message, err := s.repo.GetMessage(queryCtx, event.MessageID)
		if err == nil {
			var users []int64
			users, err = s.repo.Members(queryCtx, message.ChatID)
			if err == nil {
				s.publish(users, "message.created", MessageMutation{Message: message})
			}
		}
		cancel()
		if err != nil {
			slog.Error("chat: notification delivery failed", "message_id", event.MessageID, "err", err)
		}
	}
}
