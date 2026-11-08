package chat

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Register(r *gin.RouterGroup, pool *pgxpool.Pool, guard Guard, blobs Blobs) {
	h := &Handler{svc: NewService(newRepository(pool), guard, blobs)}

	r.GET("/chats", h.List)
	r.POST("/chats/direct", h.OpenDirect)
	r.POST("/chats/group", h.OpenGroup)
	r.GET("/chats/:id", h.Get)

	r.GET("/chats/:id/messages", h.Messages)
	r.POST("/chats/:id/messages", h.Send)
	r.PATCH("/chats/:id/messages/:mid", h.Edit)
	r.DELETE("/chats/:id/messages/:mid", h.DeleteMessage)
	r.POST("/chats/:id/read", h.Read)
}
