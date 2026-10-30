package users

import (
	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

type Handler struct{ svc *Service }

type updateProfileRequest struct {
	LastName     *string `json:"last_name" binding:"omitempty,min=1,max=50"`
	FirstName    *string `json:"first_name" binding:"omitempty,min=1,max=50"`
	MiddleNames  *string `json:"middle_names" binding:"omitempty,max=50"`
	AvatarFileID *int64  `json:"avatar_file_id"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"`
}

type changeEmailRequest struct {
	Password string `json:"password" binding:"required"`
	NewEmail string `json:"new_email" binding:"required,email,max=255"`
}

type createUserRequest struct {
	Role             string  `json:"role" binding:"required,oneof=student teacher admin"`
	LastName         string  `json:"last_name" binding:"required,max=50"`
	FirstName        string  `json:"first_name" binding:"required,max=50"`
	MiddleNames      *string `json:"middle_names" binding:"omitempty,max=50"`
	Email            string  `json:"email" binding:"required,email,max=255"`
	Password         string  `json:"password" binding:"required,min=8,max=72"`
	GroupID          *int64  `json:"group_id"`
	DepartmentID     *int64  `json:"department_id"`
	RecordBookNumber *string `json:"record_book_number" binding:"omitempty,max=50"`
	IsHeadman        bool    `json:"is_headman"`
}

type adminUpdateRequest struct {
	LastName         *string `json:"last_name" binding:"omitempty,min=1,max=50"`
	FirstName        *string `json:"first_name" binding:"omitempty,min=1,max=50"`
	MiddleNames      *string `json:"middle_names" binding:"omitempty,max=50"`
	Email            *string `json:"email" binding:"omitempty,email,max=255"`
	Password         *string `json:"password" binding:"omitempty,min=8,max=72"` // сброс пароля администратором
	IsActive         *bool   `json:"is_active"`
	IsAdmin          *bool   `json:"is_admin"`
	GroupID          *int64  `json:"group_id"`
	DepartmentID     *int64  `json:"department_id"`
	IsHeadman        *bool   `json:"is_headman"`
	RecordBookNumber *string `json:"record_book_number" binding:"omitempty,max=50"`
}

// GET /me
func (h *Handler) Me(c *gin.Context) {
	u, err := h.svc.Get(c.Request.Context(), middleware.Me(c).UserID)
	httpx.Reply(c, u, err)
}

// PATCH /me
func (h *Handler) UpdateMe(c *gin.Context) {
	var req updateProfileRequest
	if !httpx.Bind(c, &req) {
		return
	}
	u, err := h.svc.UpdateProfile(c.Request.Context(), middleware.Me(c).UserID, UpdateProfile{
		LastName: req.LastName, FirstName: req.FirstName, MiddleNames: req.MiddleNames,
		AvatarFileID: req.AvatarFileID,
	})
	httpx.Reply(c, u, err)
}

// POST /me/password
func (h *Handler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if !httpx.Bind(c, &req) {
		return
	}
	httpx.ReplyEmpty(c, h.svc.ChangePassword(c.Request.Context(), middleware.Me(c).UserID, req.OldPassword, req.NewPassword))
}

// POST /me/email
func (h *Handler) ChangeEmail(c *gin.Context) {
	var req changeEmailRequest
	if !httpx.Bind(c, &req) {
		return
	}
	u, err := h.svc.ChangeEmail(c.Request.Context(), middleware.Me(c).UserID, req.Password, req.NewEmail)
	httpx.Reply(c, u, err)
}

// GET /users?role=&group_id=&department_id=&q=&limit=&offset=
func (h *Handler) List(c *gin.Context) {
	groupID, ok := httpx.OptInt64(c, "group_id")
	if !ok {
		return
	}
	deptID, ok := httpx.OptInt64(c, "department_id")
	if !ok {
		return
	}
	limit, offset := httpx.Page(c)
	list, err := h.svc.List(c.Request.Context(), ListFilter{
		Role: c.Query("role"), GroupID: groupID, DepartmentID: deptID,
		Q: c.Query("q"), Limit: limit, Offset: offset,
	})
	httpx.Reply(c, list, err)
}

// GET /groups/:id/students
func (h *Handler) GroupStudents(c *gin.Context) {
	groupID, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	list, err := h.svc.List(c.Request.Context(), ListFilter{Role: "student", GroupID: &groupID, Limit: 500})
	httpx.Reply(c, list, err)
}

// GET /users/:id
func (h *Handler) Get(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	u, err := h.svc.Get(c.Request.Context(), id)
	httpx.Reply(c, u, err)
}

// POST /users
func (h *Handler) Create(c *gin.Context) {
	var req createUserRequest
	if !httpx.Bind(c, &req) {
		return
	}
	u, err := h.svc.Create(c.Request.Context(), CreateUser{
		Role: req.Role, LastName: req.LastName, FirstName: req.FirstName, MiddleNames: req.MiddleNames,
		Email: req.Email, Password: req.Password, GroupID: req.GroupID, DepartmentID: req.DepartmentID,
		RecordBookNumber: req.RecordBookNumber, IsHeadman: req.IsHeadman,
	})
	httpx.ReplyCreated(c, u, err)
}

// PATCH /users/:id
func (h *Handler) AdminUpdate(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req adminUpdateRequest
	if !httpx.Bind(c, &req) {
		return
	}
	u, err := h.svc.AdminUpdate(c.Request.Context(), middleware.Me(c), id, AdminUpdate{
		LastName: req.LastName, FirstName: req.FirstName, MiddleNames: req.MiddleNames,
		Email: req.Email, Password: req.Password, IsActive: req.IsActive, IsAdmin: req.IsAdmin,
		GroupID: req.GroupID, DepartmentID: req.DepartmentID, IsHeadman: req.IsHeadman,
		RecordBookNumber: req.RecordBookNumber,
	})
	httpx.Reply(c, u, err)
}
