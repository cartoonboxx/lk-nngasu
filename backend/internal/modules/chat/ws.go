package chat

import (
	"context"
	"encoding/json"
	"log/slog"

	"university-api/internal/modules/realtime"
	"university-api/internal/platform/middleware"
)

// Команды WebSocket (lk-chat.v1). Поля обязательны все: указатели отличают «нет поля» от нуля.

type sendCommand struct {
	ChatID  *int64   `json:"chat_id"`
	Body    *string  `json:"body"`
	FileIDs *[]int64 `json:"file_ids"`
}

// MessageMutation — данные события message.created.
type MessageMutation struct {
	Message Message `json:"message"`
}

func requireID(v *int64, field string) (int64, error) {
	if v == nil || *v <= 0 {
		return 0, realtime.FieldError(field, "is required and must be a positive integer")
	}
	return *v, nil
}

// registerCommands подключает команды чата к хабу и рассылает события участникам.
func registerCommands(hub *realtime.Hub, svc *Service) {
	// publish — после COMMIT, всем подключённым участникам (включая отправителя).
	// Состав считается в момент доставки, поэтому вышедший из группы событий не получит.
	publish := func(chatID int64, typ string, data any) func() {
		return func() {
			ctx, cancel := context.WithTimeout(context.Background(), realtime.CommandTimeout)
			defer cancel()
			users, err := svc.repo.Members(ctx, chatID)
			if err != nil {
				slog.Error("chat: members for event", "chat_id", chatID, "err", err)
				return
			}
			hub.Publish(users, typ, data)
		}
	}

	hub.Handle("message.send", func(ctx context.Context, p middleware.Principal, raw json.RawMessage) (any, func(), error) {
		var cmd sendCommand
		if err := realtime.Decode(raw, &cmd); err != nil {
			return nil, nil, err
		}
		chatID, err := requireID(cmd.ChatID, "chat_id")
		if err != nil {
			return nil, nil, err
		}
		if cmd.Body == nil {
			return nil, nil, realtime.FieldError("body", "is required")
		}
		if cmd.FileIDs == nil {
			return nil, nil, realtime.FieldError("file_ids", "is required ([] without files)")
		}
		m, err := svc.Send(ctx, p, chatID, *cmd.Body, *cmd.FileIDs)
		if err != nil {
			return nil, nil, err
		}
		return m, publish(chatID, "message.created", MessageMutation{Message: m}), nil
	})

}
