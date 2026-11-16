package structure

import (
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/httpx"
)

// Handler — HTTP-слой справочников. Сервисного слоя нет: правила простые
// (уникальность, ссылочная целостность, архив) и обеспечиваются БД и репозиторием.
// Права доступа (правка — только admin) навешиваются в Register.
type Handler struct{ repo *Repository }

// --- Тела запросов ---
//
// Создание — POST, полная замена — PUT, архивирование — PATCH с {"archived": true}.
// Длины названий проверяются после trim в normText: тег max считал бы и пробелы.

// namedBody — NamedInput: кафедра или дисциплина.
type namedBody struct {
	Name string `json:"name" binding:"required"`
}

// specializationBody — SpecializationInput.
type specializationBody struct {
	Name string `json:"name" binding:"required"`
	Code string `json:"code" binding:"required"`
}

// groupBody — GroupInput.
type groupBody struct {
	Name             string `json:"name" binding:"required"`
	SpecializationID int64  `json:"specialization_id" binding:"required,gt=0"`
	AdmissionYear    int    `json:"admission_year" binding:"required,min=2000,max=2100"`
}

// semesterBody — SemesterInput. is_current меняется отдельной командой.
type semesterBody struct {
	Name              string `json:"name" binding:"required"`
	AcademicYearStart int    `json:"academic_year_start" binding:"required,min=2000,max=2100"`
	Number            int    `json:"number" binding:"required,oneof=1 2"`
}

// roomBody — аудитория: обозначение, которое видят в расписании.
type roomBody struct {
	Name string `json:"name" binding:"required"`
}

// archiveBody — ArchiveInput. Допустимо только true: восстановления пока нет.
type archiveBody struct {
	Archived *bool `json:"archived" binding:"required"`
}

// normText обрезает пробелы и проверяет длину 1–max символов.
func normText(field, s string, max int) (string, error) {
	s = strings.TrimSpace(s)
	if s == "" || utf8.RuneCountInString(s) > max {
		return "", apperr.NewInvalid("invalid_request", "request body is invalid").
			WithDetails(map[string]any{"fields": map[string][]string{field: {"must be 1-" + strconv.Itoa(max) + " characters after trim"}}})
	}
	return s, nil
}

// failQuery — 400 invalid_query с указанием параметра.
func failQuery(c *gin.Context, name, problem string) {
	httpx.Fail(c, apperr.NewInvalid("invalid_query", "invalid query parameters").
		WithDetails(map[string]any{"fields": map[string][]string{name: {problem}}}))
}

// optBool читает "true"/"false"; пусто — nil. Другое значение — 400.
func optBool(c *gin.Context, name string) (*bool, bool) {
	switch c.Query(name) {
	case "":
		return nil, true
	case "true":
		v := true
		return &v, true
	case "false":
		v := false
		return &v, true
	}
	failQuery(c, name, "must be true or false")
	return nil, false
}

// parseList читает общие параметры списков: q, archived, include_archived, limit, offset.
//
// По умолчанию — только действующие записи. archived=true — только архив,
// include_archived=true — оба состояния; сочетать archived и include_archived нельзя.
func parseList(c *gin.Context) (ListQuery, bool) {
	lq := ListQuery{Q: strings.TrimSpace(c.Query("q")), Limit: 50}
	archived, ok := optBool(c, "archived")
	if !ok {
		return lq, false
	}
	include, ok := optBool(c, "include_archived")
	if !ok {
		return lq, false
	}
	switch {
	case archived != nil && include != nil:
		failQuery(c, "include_archived", "cannot be combined with archived")
		return lq, false
	case include != nil && *include:
		lq.Archived = nil // оба состояния
	case archived != nil:
		lq.Archived = archived
	default:
		f := false
		lq.Archived = &f
	}
	if raw := c.Query("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 200 {
			failQuery(c, "limit", "must be an integer from 1 to 200")
			return lq, false
		}
		lq.Limit = v
	}
	if raw := c.Query("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			failQuery(c, "offset", "must be an integer >= 0")
			return lq, false
		}
		lq.Offset = v
	}
	return lq, true
}

// replyPage — ответ Page<T> с meta по ListQuery.
func replyPage[T any](c *gin.Context, list []T, total int, lq ListQuery, err error) {
	httpx.ReplyPage(c, list, httpx.Meta{Limit: lq.Limit, Offset: lq.Offset, Total: total}, err)
}

// bindArchive разбирает ArchiveInput; false — 400 (восстановление не поддерживается).
func bindArchive(c *gin.Context) bool {
	var req archiveBody
	if !httpx.Bind(c, &req) {
		return false
	}
	if !*req.Archived {
		httpx.Fail(c, apperr.NewInvalid("invalid_request", "request body is invalid").
			WithDetails(map[string]any{"fields": map[string][]string{"archived": {"only true is supported: restoring is not available"}}}))
		return false
	}
	return true
}

// catalogHandlers строит хендлеры для справочника «id + название + archived».
// Хендлеры замыкаются на cat, так что один код обслуживает любую таблицу такого вида.
func catalogHandlers(cat catalog) (list, create, update, arch, remove gin.HandlerFunc) {
	// GET path?q=&archived=&include_archived=&limit=&offset= — страница NamedCatalog.
	list = func(c *gin.Context) {
		lq, ok := parseList(c)
		if !ok {
			return
		}
		items, total, err := cat.list(c.Request.Context(), lq)
		replyPage(c, items, total, lq, err)
	}
	// POST path — создать (201).
	create = func(c *gin.Context) {
		var req namedBody
		if !httpx.Bind(c, &req) {
			return
		}
		name, err := normText("name", req.Name, 200)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		n, err := cat.create(c.Request.Context(), name)
		httpx.ReplyCreated(c, n, err)
	}
	// PUT path/:id — полная замена (переименование).
	update = func(c *gin.Context) {
		id, ok := httpx.ID(c, "id")
		if !ok {
			return
		}
		var req namedBody
		if !httpx.Bind(c, &req) {
			return
		}
		name, err := normText("name", req.Name, 200)
		if err != nil {
			httpx.Fail(c, err)
			return
		}
		n, err := cat.update(c.Request.Context(), id, name)
		httpx.Reply(c, n, err)
	}
	// PATCH path/:id {"archived": true} — в архив.
	arch = func(c *gin.Context) {
		id, ok := httpx.ID(c, "id")
		if !ok || !bindArchive(c) {
			return
		}
		n, err := cat.archive(c.Request.Context(), id)
		httpx.Reply(c, n, err)
	}
	// DELETE path/:id — удалить (204); используемую запись не удалить.
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

func (h *Handler) ListSpecializations(c *gin.Context) {
	lq, ok := parseList(c)
	if !ok {
		return
	}
	list, total, err := h.repo.ListSpecializations(c.Request.Context(), lq)
	replyPage(c, list, total, lq, err)
}

func (h *Handler) bindSpecialization(c *gin.Context) (code, name string, ok bool) {
	var req specializationBody
	if !httpx.Bind(c, &req) {
		return "", "", false
	}
	var err error
	if name, err = normText("name", req.Name, 200); err == nil {
		code, err = normText("code", req.Code, 20)
	}
	if err != nil {
		httpx.Fail(c, err)
		return "", "", false
	}
	return code, name, true
}

func (h *Handler) CreateSpecialization(c *gin.Context) {
	code, name, ok := h.bindSpecialization(c)
	if !ok {
		return
	}
	s, err := h.repo.CreateSpecialization(c.Request.Context(), code, name)
	httpx.ReplyCreated(c, s, err)
}

func (h *Handler) UpdateSpecialization(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	code, name, ok := h.bindSpecialization(c)
	if !ok {
		return
	}
	s, err := h.repo.UpdateSpecialization(c.Request.Context(), id, code, name)
	httpx.Reply(c, s, err)
}

func (h *Handler) ArchiveSpecialization(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok || !bindArchive(c) {
		return
	}
	s, err := h.repo.ArchiveSpecialization(c.Request.Context(), id)
	httpx.Reply(c, s, err)
}

func (h *Handler) DeleteSpecialization(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteSpecialization(c.Request.Context(), id))
}

// ---- groups ----

// ListGroups — страница групп (по умолчанию только действующие).
//
// GET /groups?q=&archived=&include_archived=&specialization_id=&admission_year=&limit=&offset=
func (h *Handler) ListGroups(c *gin.Context) {
	lq, ok := parseList(c)
	if !ok {
		return
	}
	spec, ok := httpx.OptInt64(c, "specialization_id")
	if !ok {
		return
	}
	year, ok := httpx.OptInt64(c, "admission_year")
	if !ok {
		return
	}
	list, total, err := h.repo.ListGroups(c.Request.Context(), GroupFilter{ListQuery: lq, SpecializationID: spec, AdmissionYear: year})
	replyPage(c, list, total, lq, err)
}

func (h *Handler) GetGroup(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	g, err := h.repo.GetGroup(c.Request.Context(), id)
	httpx.Reply(c, g, err)
}

func bindGroup(c *gin.Context) (GroupInput, bool) {
	var req groupBody
	if !httpx.Bind(c, &req) {
		return GroupInput{}, false
	}
	name, err := normText("name", req.Name, 200)
	if err != nil {
		httpx.Fail(c, err)
		return GroupInput{}, false
	}
	return GroupInput{Name: name, SpecializationID: req.SpecializationID, AdmissionYear: req.AdmissionYear}, true
}

// CreateGroup — направление должно быть действующим (иначе 400 invalid_reference).
func (h *Handler) CreateGroup(c *gin.Context) {
	in, ok := bindGroup(c)
	if !ok {
		return
	}
	id, err := h.repo.CreateGroup(c.Request.Context(), in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	g, err := h.repo.GetGroup(c.Request.Context(), id)
	httpx.ReplyCreated(c, g, err)
}

// UpdateGroup — полная замена; оставить прежнее направление можно, даже если оно в архиве.
func (h *Handler) UpdateGroup(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	in, ok := bindGroup(c)
	if !ok {
		return
	}
	if err := h.repo.UpdateGroup(c.Request.Context(), id, in); err != nil {
		httpx.Fail(c, err)
		return
	}
	g, err := h.repo.GetGroup(c.Request.Context(), id)
	httpx.Reply(c, g, err)
}

func (h *Handler) ArchiveGroup(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok || !bindArchive(c) {
		return
	}
	g, err := h.repo.ArchiveGroup(c.Request.Context(), id)
	httpx.Reply(c, g, err)
}

func (h *Handler) DeleteGroup(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteGroup(c.Request.Context(), id))
}

// ---- semesters ----

func (h *Handler) ListSemesters(c *gin.Context) {
	lq, ok := parseList(c)
	if !ok {
		return
	}
	list, total, err := h.repo.ListSemesters(c.Request.Context(), lq)
	replyPage(c, list, total, lq, err)
}

func (h *Handler) CurrentSemester(c *gin.Context) {
	s, err := h.repo.CurrentSemester(c.Request.Context())
	httpx.Reply(c, s, err)
}

func bindSemester(c *gin.Context) (SemesterInput, bool) {
	var req semesterBody
	if !httpx.Bind(c, &req) {
		return SemesterInput{}, false
	}
	name, err := normText("name", req.Name, 200)
	if err != nil {
		httpx.Fail(c, err)
		return SemesterInput{}, false
	}
	return SemesterInput{Name: name, AcademicYearStart: req.AcademicYearStart, Number: req.Number}, true
}

func (h *Handler) CreateSemester(c *gin.Context) {
	in, ok := bindSemester(c)
	if !ok {
		return
	}
	id, err := h.repo.CreateSemester(c.Request.Context(), in)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	s, err := h.repo.GetSemester(c.Request.Context(), id)
	httpx.ReplyCreated(c, s, err)
}

func (h *Handler) UpdateSemester(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	in, ok := bindSemester(c)
	if !ok {
		return
	}
	if err := h.repo.UpdateSemester(c.Request.Context(), id, in); err != nil {
		httpx.Fail(c, err)
		return
	}
	s, err := h.repo.GetSemester(c.Request.Context(), id)
	httpx.Reply(c, s, err)
}

// ArchiveSemester — текущий семестр архивировать нельзя (409 current_semester_in_use).
func (h *Handler) ArchiveSemester(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok || !bindArchive(c) {
		return
	}
	s, err := h.repo.ArchiveSemester(c.Request.Context(), id)
	httpx.Reply(c, s, err)
}

// SetCurrentSemester — отдельная команда: снять флаг с прежнего и поставить на новый атомарно.
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

func (h *Handler) DeleteSemester(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteSemester(c.Request.Context(), id))
}

// ---- rooms ----

func (h *Handler) ListRooms(c *gin.Context) {
	lq, ok := parseList(c)
	if !ok {
		return
	}
	list, total, err := h.repo.ListRooms(c.Request.Context(), lq)
	replyPage(c, list, total, lq, err)
}

func bindRoom(c *gin.Context) (string, bool) {
	var req roomBody
	if !httpx.Bind(c, &req) {
		return "", false
	}
	name, err := normText("name", req.Name, 200)
	if err != nil {
		httpx.Fail(c, err)
		return "", false
	}
	return name, true
}

func (h *Handler) CreateRoom(c *gin.Context) {
	name, ok := bindRoom(c)
	if !ok {
		return
	}
	rm, err := h.repo.CreateRoom(c.Request.Context(), name)
	httpx.ReplyCreated(c, rm, err)
}

func (h *Handler) UpdateRoom(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	name, ok := bindRoom(c)
	if !ok {
		return
	}
	rm, err := h.repo.UpdateRoom(c.Request.Context(), id, name)
	httpx.Reply(c, rm, err)
}

func (h *Handler) ArchiveRoom(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok || !bindArchive(c) {
		return
	}
	rm, err := h.repo.ArchiveRoom(c.Request.Context(), id)
	httpx.Reply(c, rm, err)
}

func (h *Handler) DeleteRoom(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteRoom(c.Request.Context(), id))
}
