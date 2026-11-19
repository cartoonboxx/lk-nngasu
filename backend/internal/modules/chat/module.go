package chat

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"university-api/internal/modules/realtime"
)

// Возвращает сервис: его список чатов нужен главной странице (GET /dashboard).
// Отправка, правка и удаление сообщений — команды WebSocket, регистрируются в hub.
func Register(r *gin.RouterGroup, pool *pgxpool.Pool, guard Guard, blobs Blobs, hub *realtime.Hub) *Service {
	h := &Handler{svc: NewService(newRepository(pool), guard, blobs)}
	h.svc.publish = hub.Publish
	registerCommands(hub, h.svc)

	r.GET("/contacts", h.Contacts)
	r.GET("/chats", h.List)
	r.POST("/chats/direct", h.OpenDirect)
	r.POST("/chats/group", h.OpenGroup)
	r.GET("/chats/:id", h.Get)

	r.GET("/chats/:id/messages", h.Messages)
	r.GET("/chats/:id/participants", h.Participants)

	return h.svc
}
