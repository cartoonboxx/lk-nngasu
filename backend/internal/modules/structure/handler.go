package structure

import (
	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
)

type Handler struct{ repo *Repository }

type nameRequest struct {
	Name string `json:"name" binding:"required,max=150"`
}

type specializationRequest struct {
	Code string `json:"code" binding:"required,max=20"`
	Name string `json:"name" binding:"required,max=200"`
}

type groupRequest struct {
	Name             string `json:"name" binding:"required,max=50"`
	SpecializationID int64  `json:"specialization_id" binding:"required"`
	AdmissionYear    int    `json:"admission_year" binding:"required,min=2000,max=2100"`
}

type semesterRequest struct {
	AcademicYearStart int `json:"academic_year_start" binding:"required,min=2000,max=2100"`
	Number            int `json:"number" binding:"required,oneof=1 2"`
}

type roomRequest struct {
	Number string `json:"number" binding:"required,max=30"`
}

// catalogHandlers строит четыре хендлера для справочника «id + название».
func catalogHandlers(cat catalog) (list, create, update, remove gin.HandlerFunc) {
	list = func(c *gin.Context) {
		items, err := cat.list(c.Request.Context(), c.Query("q"))
		httpx.Reply(c, items, err)
	}
	create = func(c *gin.Context) {
		var req nameRequest
		if !httpx.Bind(c, &req) {
			return
		}
		n, err := cat.create(c.Request.Context(), req.Name)
		httpx.ReplyCreated(c, n, err)
	}
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
	list, err := h.repo.ListSpecializations(c.Request.Context())
	httpx.Reply(c, list, err)
}

func (h *Handler) CreateSpecialization(c *gin.Context) {
	var req specializationRequest
	if !httpx.Bind(c, &req) {
		return
	}
	s, err := h.repo.CreateSpecialization(c.Request.Context(), req.Code, req.Name)
	httpx.ReplyCreated(c, s, err)
}

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

func (h *Handler) DeleteSpecialization(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteSpecialization(c.Request.Context(), id))
}

// ---- groups ----

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

func (h *Handler) GetGroup(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	g, err := h.repo.GetGroup(c.Request.Context(), id)
	httpx.Reply(c, g, err)
}

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

func (h *Handler) DeleteGroup(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteGroup(c.Request.Context(), id))
}

// ---- semesters ----

func (h *Handler) ListSemesters(c *gin.Context) {
	list, err := h.repo.ListSemesters(c.Request.Context())
	httpx.Reply(c, list, err)
}

func (h *Handler) CurrentSemester(c *gin.Context) {
	s, err := h.repo.CurrentSemester(c.Request.Context())
	httpx.Reply(c, s, err)
}

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

func (h *Handler) DeleteSemester(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteSemester(c.Request.Context(), id))
}

// ---- rooms ----

func (h *Handler) ListRooms(c *gin.Context) {
	list, err := h.repo.ListRooms(c.Request.Context())
	httpx.Reply(c, list, err)
}

func (h *Handler) CreateRoom(c *gin.Context) {
	var req roomRequest
	if !httpx.Bind(c, &req) {
		return
	}
	rm, err := h.repo.CreateRoom(c.Request.Context(), req.Number)
	httpx.ReplyCreated(c, rm, err)
}

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

func (h *Handler) DeleteRoom(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	httpx.ReplyEmpty(c, h.repo.DeleteRoom(c.Request.Context(), id))
}
