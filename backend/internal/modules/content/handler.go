package content

import (
	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

// Handler — HTTP-слой модуля content. Каждый метод устроен одинаково:
// разобрать id из пути → (для записи) разобрать и провалидировать JSON →
// вызвать сервис, передав текущего пользователя (middleware.Me) → отправить ответ.
// Права на конкретный курс проверяет сервис, handler только передаёт пользователя.
type Handler struct{ svc *Service }

// topicRequest — тело запроса на создание и изменение темы.
// Указатели отличают «поле не передано» (nil) от пустого значения.
//
// PUT — полная замена: если description не передано, описание стирается.
// Position не указан — тема встаёт в конец списка (при обновлении — остаётся на месте).
type topicRequest struct {
	Title       string  `json:"title" binding:"required,max=200"`
	Description *string `json:"description"`
	Position    *int    `json:"position" binding:"omitempty,min=1"` // позиция в списке, с 1
}

// materialRequest — тело запроса на создание и изменение материала.
// Материал — это файл (FileID, загружен заранее через POST /files) и/или внешняя ссылка
// (ExternalURL); нужен хотя бы один из них, схему ссылки (http/https) проверяет сервис.
//
// PUT — полная замена: непереданные file_id и external_url обнуляются;
// если файл убран или заменён, старый файл удаляется.
// Position — как у темы: не указан → в конец списка (при обновлении — на месте).
type materialRequest struct {
	Title       string  `json:"title" binding:"required,max=200"`
	FileID      *int64  `json:"file_id"`
	ExternalURL *string `json:"external_url" binding:"omitempty,max=2048"`
	Position    *int    `json:"position" binding:"omitempty,min=1"`
}

// input переводит HTTP-запрос во входные данные сервиса.
// Прямое приведение типа работает, пока у структур совпадают поля (теги не учитываются).
// Если поля разойдутся, код перестанет компилироваться — это защита от рассинхрона.
func (r topicRequest) input() TopicInput { return TopicInput(r) }

// input — то же для материала.
func (r materialRequest) input() MaterialInput { return MaterialInput(r) }

// GET /courses/:id/topics — список тем курса с материалами. id — курс.
// Доступен любому авторизованному пользователю; доступ к курсу проверяет сервис.
func (h *Handler) ListTopics(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	list, err := h.svc.ListTopics(c.Request.Context(), middleware.Me(c), id)
	httpx.Reply(c, list, err)
}

// POST /courses/:id/topics — создать тему в курсе. id — курс. Ответ 201 с темой.
func (h *Handler) CreateTopic(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req topicRequest
	if !httpx.Bind(c, &req) {
		return
	}
	t, err := h.svc.CreateTopic(c.Request.Context(), middleware.Me(c), id, req.input())
	httpx.ReplyCreated(c, t, err)
}

// PUT /topics/:id — изменить тему. id — тема. Ответ 200 с обновлённой темой.
func (h *Handler) UpdateTopic(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req topicRequest
	if !httpx.Bind(c, &req) {
		return
	}
	t, err := h.svc.UpdateTopic(c.Request.Context(), middleware.Me(c), id, req.input())
	httpx.Reply(c, t, err)
}

// DELETE /topics/:id — удалить тему вместе с материалами. id — тема. Ответ 204 без тела.
func (h *Handler) DeleteTopic(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.svc.DeleteTopic(c.Request.Context(), middleware.Me(c), id))
}

// POST /topics/:id/materials — добавить материал в тему. id — тема. Ответ 201 с материалом.
func (h *Handler) CreateMaterial(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req materialRequest
	if !httpx.Bind(c, &req) {
		return
	}
	m, err := h.svc.CreateMaterial(c.Request.Context(), middleware.Me(c), id, req.input())
	httpx.ReplyCreated(c, m, err)
}

// PUT /materials/:id — изменить материал. id — материал. Ответ 200 с обновлённым материалом.
func (h *Handler) UpdateMaterial(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req materialRequest
	if !httpx.Bind(c, &req) {
		return
	}
	m, err := h.svc.UpdateMaterial(c.Request.Context(), middleware.Me(c), id, req.input())
	httpx.Reply(c, m, err)
}

// DELETE /materials/:id — удалить материал. id — материал. Ответ 204 без тела.
func (h *Handler) DeleteMaterial(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.svc.DeleteMaterial(c.Request.Context(), middleware.Me(c), id))
}
