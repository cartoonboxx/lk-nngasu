package structure

import (
	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
)

// Handler — HTTP-слой справочников. Сервисного слоя нет: правила простые
// (уникальность, ссылочная целостность) и обеспечиваются ограничениями БД,
// поэтому хендлеры вызывают репозиторий напрямую.
// Права доступа (правка — только admin) навешиваются в Register.
type Handler struct{ repo *Repository }

// --- Тела запросов ---
//
// Все изменения через PUT: клиент присылает объект целиком,
// поэтому поля здесь обычные значения, а не указатели.

// nameRequest — тело создания и переименования для справочников «id + название»
// (кафедры, дисциплины).
type nameRequest struct {
	Name string `json:"name" binding:"required,max=150"`
}

// specializationRequest — специальность: шифр направления
// (например, «09.03.04») и название.
type specializationRequest struct {
	Code string `json:"code" binding:"required,max=20"`
	Name string `json:"name" binding:"required,max=200"`
}

// groupRequest — учебная группа.
// Поля должны совпадать с GroupInput один в один: хендлеры конвертируют
// GroupInput(req) напрямую, и при расхождении код не скомпилируется.
type groupRequest struct {
	Name             string `json:"name" binding:"required,max=50"`
	SpecializationID int64  `json:"specialization_id" binding:"required"`
	AdmissionYear    int    `json:"admission_year" binding:"required,min=2000,max=2100"` // год набора
}

// semesterRequest — семестр задаётся учебным годом и номером в нём:
// {2026, 1} — осенний семестр 2026/27, {2026, 2} — весенний.
type semesterRequest struct {
	AcademicYearStart int `json:"academic_year_start" binding:"required,min=2000,max=2100"` // год начала учебного года
	Number            int `json:"number" binding:"required,oneof=1 2"`
}

// roomRequest — аудитория. Номер строкой: бывают «301а», «Б-12», «спортзал».
type roomRequest struct {
	Number string `json:"number" binding:"required,max=30"`
}

// catalogHandlers строит четыре хендлера для справочника «id + название».
// Хендлеры замыкаются на cat, так что один код обслуживает любую таблицу
// такого вида (см. registerCatalog).
func catalogHandlers(cat catalog) (list, create, update, remove gin.HandlerFunc) {
	// GET path?q= — список; q — необязательный поиск по названию.
	list = func(c *gin.Context) {
		items, err := cat.list(c.Request.Context(), c.Query("q"))
		httpx.Reply(c, items, err)
	}
	// POST path — создать, ответ 201 с созданной записью.
	create = func(c *gin.Context) {
		var req nameRequest
		if !httpx.Bind(c, &req) {
			return // ответ с ошибкой валидации уже отправлен в Bind
		}
		n, err := cat.create(c.Request.Context(), req.Name)
		httpx.ReplyCreated(c, n, err)
	}
	// PUT path/:id — переименовать.
	update = func(c *gin.Context) {
		id, ok := httpx.ID(c, "id")
		if !ok {
			return
		}
		var req nameRequest
		if !httpx.Bind(c, &req) {
			return
		}
		n, err := cat.update(c.Request.Context(), id, req.Name)
		httpx.Reply(c, n, err)
	}
	// DELETE path/:id — удалить, ответ без тела.
	remove = func(c *gin.Context) {
		id, ok := httpx.ID(c, "id")
		if !ok {
			return
		}
		httpx.ReplyEmpty(c, cat.remove(c.Request.Context(), id))
	}
	return
}

// ---- specializations ----

// ListSpecializations возвращает все специальности, без фильтров и пагинации:
// справочник небольшой.
//
// GET /specializations
func (h *Handler) ListSpecializations(c *gin.Context) {
	list, err := h.repo.ListSpecializations(c.Request.Context())
	httpx.Reply(c, list, err)
}

// CreateSpecialization создаёт специальность.
//
// POST /specializations
func (h *Handler) CreateSpecialization(c *gin.Context) {
	var req specializationRequest
	if !httpx.Bind(c, &req) {
		return
	}
	s, err := h.repo.CreateSpecialization(c.Request.Context(), req.Code, req.Name)
	httpx.ReplyCreated(c, s, err)
}

// UpdateSpecialization заменяет шифр и название специальности.
//
// PUT /specializations/:id
func (h *Handler) UpdateSpecialization(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req specializationRequest
	if !httpx.Bind(c, &req) {
		return
	}
	s, err := h.repo.UpdateSpecialization(c.Request.Context(), id, req.Code, req.Name)
	httpx.Reply(c, s, err)
}

// DeleteSpecialization удаляет специальность. Если на неё ссылаются группы,
// удаление упрётся во внешний ключ.
//
// DELETE /specializations/:id
func (h *Handler) DeleteSpecialization(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteSpecialization(c.Request.Context(), id))
}

// ---- groups ----

// ListGroups возвращает группы с необязательными фильтрами
// по специальности и году набора.
//
// GET /groups?specialization_id=&admission_year=
func (h *Handler) ListGroups(c *gin.Context) {
	spec, ok := httpx.OptInt64(c, "specialization_id")
	if !ok {
		return
	}
	year, ok := httpx.OptInt64(c, "admission_year")
	if !ok {
		return
	}
	list, err := h.repo.ListGroups(c.Request.Context(), GroupFilter{SpecializationID: spec, AdmissionYear: year})
	httpx.Reply(c, list, err)
}

// GetGroup возвращает группу по id.
//
// GET /groups/:id
func (h *Handler) GetGroup(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	g, err := h.repo.GetGroup(c.Request.Context(), id)
	httpx.Reply(c, g, err)
}

// CreateGroup создаёт группу и возвращает её перечитанной через GetGroup,
// чтобы в ответе были вычисляемые поля (например, название специальности).
//
// POST /groups
func (h *Handler) CreateGroup(c *gin.Context) {
	var req groupRequest
	if !httpx.Bind(c, &req) {
		return
	}
	id, err := h.repo.CreateGroup(c.Request.Context(), GroupInput(req))
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	g, err := h.repo.GetGroup(c.Request.Context(), id)
	httpx.ReplyCreated(c, g, err)
}

// UpdateGroup заменяет данные группы и возвращает её перечитанной.
//
// PUT /groups/:id
func (h *Handler) UpdateGroup(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req groupRequest
	if !httpx.Bind(c, &req) {
		return
	}
	if err := h.repo.UpdateGroup(c.Request.Context(), id, GroupInput(req)); err != nil {
		httpx.Fail(c, err)
		return
	}
	g, err := h.repo.GetGroup(c.Request.Context(), id)
	httpx.Reply(c, g, err)
}

// DeleteGroup удаляет группу. Если в группе есть студенты,
// удаление упрётся во внешний ключ.
//
// DELETE /groups/:id
func (h *Handler) DeleteGroup(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteGroup(c.Request.Context(), id))
}

// ---- semesters ----

// ListSemesters возвращает все семестры.
//
// GET /semesters
func (h *Handler) ListSemesters(c *gin.Context) {
	list, err := h.repo.ListSemesters(c.Request.Context())
	httpx.Reply(c, list, err)
}

// CurrentSemester возвращает семестр, отмеченный текущим.
// Если текущий не выбран — «не найдено».
//
// GET /semesters/current
func (h *Handler) CurrentSemester(c *gin.Context) {
	s, err := h.repo.CurrentSemester(c.Request.Context())
	httpx.Reply(c, s, err)
}

// CreateSemester создаёт семестр (текущим он не становится)
// и возвращает его перечитанным.
//
// POST /semesters
func (h *Handler) CreateSemester(c *gin.Context) {
	var req semesterRequest
	if !httpx.Bind(c, &req) {
		return
	}
	id, err := h.repo.CreateSemester(c.Request.Context(), req.AcademicYearStart, req.Number)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	s, err := h.repo.GetSemester(c.Request.Context(), id)
	httpx.ReplyCreated(c, s, err)
}

// UpdateSemester меняет учебный год и номер семестра.
// Признак «текущий» здесь не меняется, для этого есть SetCurrentSemester.
//
// PUT /semesters/:id
func (h *Handler) UpdateSemester(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req semesterRequest
	if !httpx.Bind(c, &req) {
		return
	}
	if err := h.repo.UpdateSemester(c.Request.Context(), id, req.AcademicYearStart, req.Number); err != nil {
		httpx.Fail(c, err)
		return
	}
	s, err := h.repo.GetSemester(c.Request.Context(), id)
	httpx.Reply(c, s, err)
}

// SetCurrentSemester делает семестр текущим; с прежнего текущего флаг снимается.
// Тела запроса нет, в ответе — обновлённый семестр.
//
// POST /semesters/:id/current
func (h *Handler) SetCurrentSemester(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	if err := h.repo.SetCurrentSemester(c.Request.Context(), id); err != nil {
		httpx.Fail(c, err)
		return
	}
	s, err := h.repo.GetSemester(c.Request.Context(), id)
	httpx.Reply(c, s, err)
}

// DeleteSemester удаляет семестр. Если удалить текущий,
// GET /semesters/current начнёт отвечать «не найдено».
//
// DELETE /semesters/:id
func (h *Handler) DeleteSemester(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteSemester(c.Request.Context(), id))
}

// ---- rooms ----

// ListRooms возвращает все аудитории.
//
// GET /rooms
func (h *Handler) ListRooms(c *gin.Context) {
	list, err := h.repo.ListRooms(c.Request.Context())
	httpx.Reply(c, list, err)
}

// CreateRoom создаёт аудиторию.
//
// POST /rooms
func (h *Handler) CreateRoom(c *gin.Context) {
	var req roomRequest
	if !httpx.Bind(c, &req) {
		return
	}
	rm, err := h.repo.CreateRoom(c.Request.Context(), req.Number)
	httpx.ReplyCreated(c, rm, err)
}

// UpdateRoom меняет номер аудитории.
//
// PUT /rooms/:id
func (h *Handler) UpdateRoom(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req roomRequest
	if !httpx.Bind(c, &req) {
		return
	}
	rm, err := h.repo.UpdateRoom(c.Request.Context(), id, req.Number)
	httpx.Reply(c, rm, err)
}

// DeleteRoom удаляет аудиторию.
//
// DELETE /rooms/:id
func (h *Handler) DeleteRoom(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteRoom(c.Request.Context(), id))
}
