package users

import (
	"github.com/gin-gonic/gin"

	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

// Handler — HTTP-слой пакета users: разбирает и валидирует запросы,
// переводит их в типы сервиса и отдаёт ответ через httpx.
// Бизнес-логики и проверок прав доступа здесь нет: права задаются
// middleware в роутере, правила предметной области проверяет Service.
type Handler struct{ svc *Service }

// --- Тела запросов ---
//
// Валидация формата (длины, email, обязательность) сделана через теги binding
// и выполняется в httpx.Bind. Указатели в PATCH-запросах означают:
// поле не передано (nil) — не менять.

// updateProfileRequest — тело PATCH /me: что пользователь меняет о себе.
type updateProfileRequest struct {
	LastName     *string `json:"last_name" binding:"omitempty,min=1,max=50"`
	FirstName    *string `json:"first_name" binding:"omitempty,min=1,max=50"`
	MiddleNames  *string `json:"middle_names" binding:"omitempty,max=50"` // без min=1: отчество может быть пустым
	AvatarFileID *int64  `json:"avatar_file_id"`                          // id ранее загруженного файла
}

// changePasswordRequest — тело POST /me/password.
// Для смены пароля нужен текущий пароль.
type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"` // 72 — предел длины пароля у bcrypt
}

// changeEmailRequest — тело POST /me/email.
// Смена email подтверждается паролем, чтобы украденная сессия
// не давала перехватить учётку через восстановление пароля.
type changeEmailRequest struct {
	Password string `json:"password" binding:"required"`
	NewEmail string `json:"new_email" binding:"required,email,max=255"`
}

// createUserRequest — тело POST /users (создание пользователя администратором).
// Поля группы и кафедры проверяются на соответствие роли в сервисе,
// здесь только формат.
type createUserRequest struct {
	Role             string  `json:"role" binding:"required,oneof=student teacher admin"`
	LastName         string  `json:"last_name" binding:"required,max=50"`
	FirstName        string  `json:"first_name" binding:"required,max=50"`
	MiddleNames      *string `json:"middle_names" binding:"omitempty,max=50"`
	Email            string  `json:"email" binding:"required,email,max=255"`
	Password         string  `json:"password" binding:"required,min=8,max=72"`
	GroupID          *int64  `json:"group_id"`                                      // для студента
	DepartmentID     *int64  `json:"department_id"`                                 // для преподавателя
	RecordBookNumber *string `json:"record_book_number" binding:"omitempty,max=50"` // для студента
	IsHeadman        bool    `json:"is_headman"`                                    // для студента
}

// registerRequest — тело POST /auth/register (самостоятельная регистрация).
// Как createUserRequest, но роль только student или teacher и без is_headman.
type registerRequest struct {
	Role             string  `json:"role" binding:"required,oneof=student teacher"`
	LastName         string  `json:"last_name" binding:"required,max=50"`
	FirstName        string  `json:"first_name" binding:"required,max=50"`
	MiddleNames      *string `json:"middle_names" binding:"omitempty,max=50"`
	Email            string  `json:"email" binding:"required,email,max=255"`
	Password         string  `json:"password" binding:"required,min=8,max=72"`
	GroupID          *int64  `json:"group_id"`                                      // для студента
	DepartmentID     *int64  `json:"department_id"`                                 // для преподавателя
	RecordBookNumber *string `json:"record_book_number" binding:"omitempty,max=50"` // для студента
}

// adminUpdateRequest — тело PATCH /users/:id (правка пользователя администратором).
type adminUpdateRequest struct {
	LastName         *string `json:"last_name" binding:"omitempty,min=1,max=50"`
	FirstName        *string `json:"first_name" binding:"omitempty,min=1,max=50"`
	MiddleNames      *string `json:"middle_names" binding:"omitempty,max=50"`
	Email            *string `json:"email" binding:"omitempty,email,max=255"`
	Password         *string `json:"password" binding:"omitempty,min=8,max=72"` // сброс пароля администратором, старый не нужен
	IsActive         *bool   `json:"is_active"`                                 // блокировка/разблокировка
	IsAdmin          *bool   `json:"is_admin"`                                  // выдать/снять права администратора
	GroupID          *int64  `json:"group_id"`
	DepartmentID     *int64  `json:"department_id"`
	IsHeadman        *bool   `json:"is_headman"`
	RecordBookNumber *string `json:"record_book_number" binding:"omitempty,max=50"`
}

// --- Текущий пользователь (/me) ---
// Id берётся из сессии (middleware.Me), а не из URL, поэтому
// пользователь физически может работать только со своей учёткой.

// Me возвращает профиль текущего пользователя.
//
// GET /me
func (h *Handler) Me(c *gin.Context) {
	u, err := h.svc.Get(c.Request.Context(), middleware.Me(c).UserID)
	httpx.Reply(c, u, err)
}

// UpdateMe частично обновляет профиль текущего пользователя (ФИО, аватар)
// и возвращает обновлённый профиль.
//
// PATCH /me
func (h *Handler) UpdateMe(c *gin.Context) {
	var req updateProfileRequest
	if !httpx.Bind(c, &req) {
		return // ответ с ошибкой валидации уже отправлен в Bind
	}
	u, err := h.svc.UpdateProfile(c.Request.Context(), middleware.Me(c).UserID, UpdateProfile{
		LastName: req.LastName, FirstName: req.FirstName, MiddleNames: req.MiddleNames,
		AvatarFileID: req.AvatarFileID,
	})
	httpx.Reply(c, u, err)
}

// ChangePassword меняет пароль текущего пользователя после проверки старого.
// В ответе тела нет.
//
// POST /me/password
func (h *Handler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if !httpx.Bind(c, &req) {
		return
	}
	httpx.ReplyEmpty(c, h.svc.ChangePassword(c.Request.Context(), middleware.Me(c).UserID, req.OldPassword, req.NewPassword))
}

// ChangeEmail меняет email текущего пользователя после проверки пароля
// и возвращает обновлённый профиль.
//
// POST /me/email
func (h *Handler) ChangeEmail(c *gin.Context) {
	var req changeEmailRequest
	if !httpx.Bind(c, &req) {
		return
	}
	u, err := h.svc.ChangeEmail(c.Request.Context(), middleware.Me(c).UserID, req.Password, req.NewEmail)
	httpx.Reply(c, u, err)
}

// --- Работа с пользователями по id ---

// List возвращает страницу пользователей с фильтрами из query-параметров.
// Все параметры необязательны; некорректный group_id или department_id
// даёт 400 (ответ отправляют OptInt64).
//
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
	limit, offset := httpx.Page(c) // значения по умолчанию и верхний предел limit — в httpx.Page
	list, err := h.svc.List(c.Request.Context(), ListFilter{
		Role: c.Query("role"), GroupID: groupID, DepartmentID: deptID,
		Q: c.Query("q"), Limit: limit, Offset: offset,
	})
	httpx.Reply(c, list, err)
}

// GroupStudents возвращает всех студентов группы одним списком, без пагинации.
// Лимит 500 — защитная верхняя граница: в реальной группе студентов
// заведомо меньше.
//
// GET /groups/:id/students
func (h *Handler) GroupStudents(c *gin.Context) {
	groupID, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	list, err := h.svc.List(c.Request.Context(), ListFilter{Role: "student", GroupID: &groupID, Limit: 500})
	httpx.Reply(c, list, err)
}

// Get возвращает пользователя по id.
//
// GET /users/:id
func (h *Handler) Get(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	u, err := h.svc.Get(c.Request.Context(), id)
	httpx.Reply(c, u, err)
}

// Create создаёт пользователя (действие администратора) и отвечает 201
// с созданным пользователем.
//
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

// SelfRegister — самостоятельная регистрация по email и паролю. Маршрут публичный.
// Ответ 201 с созданным пользователем; поле is_active показывает, можно ли сразу
// войти (студент) или нужно дождаться одобрения администратора (преподаватель).
//
// POST /auth/register
func (h *Handler) SelfRegister(c *gin.Context) {
	var req registerRequest
	if !httpx.Bind(c, &req) {
		return
	}
	u, err := h.svc.SelfRegister(c.Request.Context(), CreateUser{
		Role: req.Role, LastName: req.LastName, FirstName: req.FirstName, MiddleNames: req.MiddleNames,
		Email: req.Email, Password: req.Password, GroupID: req.GroupID, DepartmentID: req.DepartmentID,
		RecordBookNumber: req.RecordBookNumber,
	})
	httpx.ReplyCreated(c, u, err)
}

// AdminUpdate частично обновляет пользователя от имени администратора.
// В сервис передаётся и сам администратор (middleware.Me), чтобы проверить
// действия над собой, например запретить снять с себя админку
// или заблокировать собственную учётку.
//
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
