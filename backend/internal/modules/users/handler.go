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
// Формат (обязательность, email, длины) проверяется тегами binding в httpx.Bind.
// Правила, которые тегами не выразить (trim имён, цифры документов, null там,
// где он запрещён), проверяет сервис. Поля PATCH обёрнуты в Field, чтобы отличать
// «ключа нет» от «передан null».

// studentWriteBody — StudentWrite: полный объект роли при создании или замене.
// Название группы не передаётся, его вычисляет сервер.
type studentWriteBody struct {
	GroupID          int64  `json:"group_id" binding:"required,gt=0"`
	RecordBookNumber string `json:"record_book_number" binding:"required,max=50"`
	IsHeadman        *bool  `json:"is_headman" binding:"required"` // указатель: required у bool отверг бы false
}

func (b studentWriteBody) role() StudentRole {
	n := b.RecordBookNumber
	return StudentRole{GroupID: b.GroupID, RecordBookNumber: &n, IsHeadman: *b.IsHeadman}
}

// teacherWriteBody — TeacherWrite.
type teacherWriteBody struct {
	DepartmentID int64 `json:"department_id" binding:"required,gt=0"`
}

func (b teacherWriteBody) role() TeacherRole { return TeacherRole{DepartmentID: b.DepartmentID} }

// userCreateBody — UserCreateInput (POST /users). Пароля нет: сервер отправляет приглашение.
// student и teacher обязательны как ключи: null — роли нет, объект — роль есть.
type userCreateBody struct {
	LastName     string                  `json:"last_name" binding:"required,max=50"`
	FirstName    string                  `json:"first_name" binding:"required,max=50"`
	MiddleNames  *string                 `json:"middle_names" binding:"omitempty,max=50"`
	Email        string                  `json:"email" binding:"required,email,max=255"`
	Passport     string                  `json:"passport" binding:"required,max=20"` // "45 09 123456" допустимо
	Snils        string                  `json:"snils" binding:"required,max=20"`    // "112-233-445 95" допустимо
	Student      Field[studentWriteBody] `json:"student"`
	Teacher      Field[teacherWriteBody] `json:"teacher"`
	AvatarFileID *int64                  `json:"avatar_file_id"`
}

// userPatchBody — UserPatchInput (PATCH /users/:id). Все поля необязательны, нужен хотя бы один.
type userPatchBody struct {
	LastName     Field[string]           `json:"last_name"`
	FirstName    Field[string]           `json:"first_name"`
	MiddleNames  Field[string]           `json:"middle_names"`
	Email        Field[string]           `json:"email"`
	Passport     Field[string]           `json:"passport"`
	Snils        Field[string]           `json:"snils"`
	Student      Field[studentWriteBody] `json:"student"`
	Teacher      Field[teacherWriteBody] `json:"teacher"`
	AvatarFileID Field[int64]            `json:"avatar_file_id"`
	IsActive     Field[bool]             `json:"is_active"`
	IsAdmin      Field[bool]             `json:"is_admin"` // не описано в документации, оставлено для выдачи прав
}

// profilePatchBody — ProfilePatchInput (PATCH /me).
type profilePatchBody struct {
	LastName     Field[string] `json:"last_name"`
	FirstName    Field[string] `json:"first_name"`
	MiddleNames  Field[string] `json:"middle_names"`
	AvatarFileID Field[int64]  `json:"avatar_file_id"`
}

// passwordBody — PasswordInput (POST /me/password).
type passwordBody struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"` // и не больше 72 байт — проверяет сервис
}

// changeEmailRequest — тело POST /me/email (в документации не описано).
// Смена email подтверждается паролем, чтобы украденная сессия
// не давала перехватить учётку через восстановление пароля.
type changeEmailRequest struct {
	Password string `json:"password" binding:"required"`
	NewEmail string `json:"new_email" binding:"required,email,max=255"`
}

// --- Текущий пользователь (/me) ---
// Id берётся из сессии (middleware.Me), а не из URL, поэтому
// пользователь физически может работать только со своей учёткой.

// Me возвращает свой профиль (Profile).
//
// GET /me
func (h *Handler) Me(c *gin.Context) {
	u, err := h.svc.Get(c.Request.Context(), middleware.Me(c).UserID)
	httpx.Reply(c, u.User, err)
}

// UpdateMe частично обновляет свой профиль (ФИО, аватар) и возвращает его.
//
// PATCH /me
func (h *Handler) UpdateMe(c *gin.Context) {
	var req profilePatchBody
	if !httpx.Bind(c, &req) {
		return // ответ с ошибкой валидации уже отправлен в Bind
	}
	u, err := h.svc.UpdateProfile(c.Request.Context(), middleware.Me(c).UserID, UpdateProfile{
		LastName: req.LastName, FirstName: req.FirstName, MiddleNames: req.MiddleNames,
		AvatarFileID: req.AvatarFileID,
	})
	httpx.Reply(c, u, err)
}

// ChangePassword меняет свой пароль после проверки старого. Ответ 204.
//
// POST /me/password
func (h *Handler) ChangePassword(c *gin.Context) {
	var req passwordBody
	if !httpx.Bind(c, &req) {
		return
	}
	me := middleware.Me(c)
	httpx.ReplyEmpty(c, h.svc.ChangePassword(c.Request.Context(), me.UserID, me.Sid, req.OldPassword, req.NewPassword))
}

// ChangeEmail меняет свой email после проверки пароля и возвращает профиль.
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

// List возвращает страницу UserSummary с фильтрами из query-параметров.
// Некорректный group_id или department_id даёт 400 (ответ отправляют OptInt64).
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
	list, total, err := h.svc.List(c.Request.Context(), ListFilter{
		Role: c.Query("role"), GroupID: groupID, DepartmentID: deptID,
		Q: c.Query("q"), Limit: limit, Offset: offset,
	})
	httpx.ReplyPage(c, list, httpx.Meta{Limit: limit, Offset: offset, Total: total}, err)
}

// GroupStudents возвращает всех студентов группы одним списком (UserSummary), без пагинации.
// Лимит 500 — защитная верхняя граница: в реальной группе студентов заведомо меньше.
//
// GET /groups/:id/students
func (h *Handler) GroupStudents(c *gin.Context) {
	groupID, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	list, _, err := h.svc.List(c.Request.Context(), ListFilter{Role: "student", GroupID: &groupID, Limit: 500})
	httpx.Reply(c, list, err)
}

// Get возвращает карточку пользователя. Администратор получает AdminUser
// (с паспортом, СНИЛС и состоянием пароля), преподаватель — UserSummary без документов.
//
// GET /users/:id
func (h *Handler) Get(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	u, err := h.svc.Get(c.Request.Context(), id)
	if err != nil || middleware.Me(c).IsAdmin() {
		c.Header("Cache-Control", "no-store") // в ответе документы
		httpx.Reply(c, u, err)
		return
	}
	httpx.Reply(c, u.User, nil)
}

// Create создаёт пользователя без пароля и отправляет ему приглашение.
// Ответ 201 с AdminUser (credential_state = "invited").
//
// POST /users
func (h *Handler) Create(c *gin.Context) {
	var req userCreateBody
	if !httpx.Bind(c, &req) {
		return
	}
	// Ключи student и teacher обязательны, хотя значение может быть null.
	if !req.Student.Set {
		httpx.Fail(c, fieldErr("student", "required (null if the user is not a student)"))
		return
	}
	if !req.Teacher.Set {
		httpx.Fail(c, fieldErr("teacher", "required (null if the user is not a teacher)"))
		return
	}
	in := CreateUser{
		LastName: req.LastName, FirstName: req.FirstName, MiddleNames: req.MiddleNames,
		Email: req.Email, Passport: req.Passport, Snils: req.Snils,
		AvatarFileID: req.AvatarFileID, CreatedBy: middleware.Me(c).UserID,
	}
	if req.Student.Value != nil {
		r := req.Student.Value.role()
		in.Student = &r
	}
	if req.Teacher.Value != nil {
		r := req.Teacher.Value.role()
		in.Teacher = &r
	}
	c.Header("Cache-Control", "no-store")
	u, err := h.svc.Create(c.Request.Context(), in)
	httpx.ReplyCreated(c, u, err)
}

// ResendInvitation отправляет новое приглашение пользователю без пароля.
// Прежние ссылки перестают работать. Ответ 202 с InvitationReceipt.
//
// POST /users/:id/invitation
func (h *Handler) ResendInvitation(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	r, err := h.svc.ResendInvitation(c.Request.Context(), id, middleware.Me(c).UserID)
	httpx.ReplyAccepted(c, r, err)
}

// AdminUpdate частично обновляет пользователя от имени администратора
// и возвращает AdminUser. В сервис передаётся и сам администратор, чтобы
// запретить ему заблокировать себя или снять с себя права.
//
// PATCH /users/:id
func (h *Handler) AdminUpdate(c *gin.Context) {
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req userPatchBody
	if !httpx.Bind(c, &req) {
		return
	}
	in := AdminUpdate{
		LastName: req.LastName, FirstName: req.FirstName, MiddleNames: req.MiddleNames,
		Email: req.Email, Passport: req.Passport, Snils: req.Snils,
		Student:      mapField(req.Student, studentWriteBody.role),
		Teacher:      mapField(req.Teacher, teacherWriteBody.role),
		AvatarFileID: req.AvatarFileID, IsActive: req.IsActive, IsAdmin: req.IsAdmin,
	}
	c.Header("Cache-Control", "no-store")
	u, err := h.svc.AdminUpdate(c.Request.Context(), middleware.Me(c), id, in)
	httpx.Reply(c, u, err)
}
