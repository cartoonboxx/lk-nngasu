package users

import (
	"strconv"
	"strings"
	"unicode/utf8"

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
// record_book_number обязателен как ключ, но может быть null (зачётки ещё нет).
type studentWriteBody struct {
	GroupID          int64         `json:"group_id" binding:"required,gt=0"`
	RecordBookNumber Field[string] `json:"record_book_number"`
	IsHeadman        *bool         `json:"is_headman" binding:"required"` // указатель: required у bool отверг бы false
}

func (b studentWriteBody) role() StudentRole {
	return StudentRole{GroupID: b.GroupID, RecordBookNumber: b.RecordBookNumber.Value, IsHeadman: *b.IsHeadman}
}

// checkStudentBody — ключ record_book_number обязателен (null допустим), до 50 символов.
func checkStudentBody(b *studentWriteBody) error {
	if b == nil {
		return nil
	}
	if !b.RecordBookNumber.Set {
		return fieldErr("student.record_book_number", "required (null if there is no record book yet)")
	}
	if v := b.RecordBookNumber.Value; v != nil && utf8.RuneCountInString(*v) > 50 {
		return fieldErr("student.record_book_number", "must be at most 50 characters")
	}
	return nil
}

// teacherWriteBody — TeacherWrite.
type teacherWriteBody struct {
	DepartmentID int64 `json:"department_id" binding:"required,gt=0"`
}

func (b teacherWriteBody) role() TeacherRole { return TeacherRole{DepartmentID: b.DepartmentID} }

// userCreateBody — UserCreateInput (POST /users).
// student и teacher обязательны как ключи: null — профиля нет, объект — есть.
// role не передан — выводится из профиля (совместимость с прежней формой админки);
// password не передан — пользователь получит приглашение задать пароль.
type userCreateBody struct {
	Role         string                  `json:"role" binding:"omitempty,oneof=student teacher"`
	Password     string                  `json:"password" binding:"omitempty,min=8"`
	LastName     string                  `json:"last_name" binding:"required,max=50"`
	FirstName    string                  `json:"first_name" binding:"required,max=50"`
	MiddleNames  *string                 `json:"middle_names" binding:"omitempty,max=50"`
	Email        string                  `json:"email" binding:"required,email,max=255"`
	Passport     string                  `json:"passport" binding:"required,max=20"` // "45 09 123456" допустимо
	Snils        string                  `json:"snils" binding:"required,max=20"`    // "112-233-445 95" допустимо
	Student      Field[studentWriteBody] `json:"student"`
	Teacher      Field[teacherWriteBody] `json:"teacher"`
	AvatarFileID *int64                  `json:"avatar"`
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
	AvatarFileID Field[int64]            `json:"avatar"`
	IsActive     Field[bool]             `json:"is_active"`
	Role         Field[string]           `json:"role"`
}

// profilePatchBody — ProfilePatchInput (PATCH /me).
type profilePatchBody struct {
	LastName     Field[string] `json:"last_name"`
	FirstName    Field[string] `json:"first_name"`
	MiddleNames  Field[string] `json:"middle_names"`
	AvatarFileID Field[int64]  `json:"avatar"`
}

// passwordBody — PasswordInput (POST /me/password).
type passwordBody struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=8,max=72"` // и не больше 72 байт — проверяет сервис
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
	httpx.ReplyEmpty(c, h.svc.ChangePassword(c.Request.Context(), middleware.Me(c).UserID, req.OldPassword, req.NewPassword))
}

// --- Работа с пользователями по id ---

// List возвращает страницу UserSummary с фильтрами из query-параметров.
// Некорректный параметр — 400 invalid_query с details.fields.
//
// GET /users?q=&role=&is_headman=&is_active=&group_id=&department_id=&sort=&limit=&offset=
func (h *Handler) List(c *gin.Context) {
	f, ok := parseUserFilter(c)
	if !ok {
		return
	}
	list, total, err := h.svc.List(c.Request.Context(), f)
	httpx.ReplyPage(c, list, httpx.Meta{Limit: f.Limit, Offset: f.Offset, Total: total}, err)
}

// parseUserFilter разбирает и проверяет query-параметры списка пользователей.
// При ошибке ответ 400 уже отправлен.
func parseUserFilter(c *gin.Context) (ListFilter, bool) {
	f := ListFilter{Q: strings.TrimSpace(c.Query("q")), Role: c.Query("role"), Limit: 50}
	if utf8.RuneCountInString(f.Q) > 200 {
		failQuery(c, "q", "must be at most 200 characters")
		return f, false
	}
	switch f.Role {
	case "", RoleStudent, RoleTeacher, RoleAdmin:
	default:
		failQuery(c, "role", "must be student, teacher or admin")
		return f, false
	}
	for _, p := range []struct {
		name string
		dst  **bool
	}{{"is_headman", &f.IsHeadman}, {"is_active", &f.IsActive}} {
		raw := c.Query(p.name)
		if raw == "" {
			continue
		}
		v, err := strconv.ParseBool(raw)
		if err != nil || (raw != "true" && raw != "false") {
			failQuery(c, p.name, "must be true or false")
			return f, false
		}
		*p.dst = &v
	}
	for _, p := range []struct {
		name string
		dst  **int64
	}{{"group_id", &f.GroupID}, {"department_id", &f.DepartmentID}} {
		raw := c.Query(p.name)
		if raw == "" {
			continue
		}
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || v <= 0 {
			failQuery(c, p.name, "must be a positive integer")
			return f, false
		}
		*p.dst = &v
	}
	switch c.Query("sort") {
	case "", "full_name":
	case "-full_name":
		f.Desc = true
	default:
		failQuery(c, "sort", "must be full_name or -full_name")
		return f, false
	}
	var ok bool
	f.Limit, f.Offset, ok = parsePage(c)
	return f, ok
}

// GroupStudents возвращает страницу состава группы (CourseStudent): только данные,
// нужные для учебной работы, без email, документов и полной карточки.
// Нет группы или нет доступа к ней — 404 group_not_found (не раскрываем, что группа есть).
//
// GET /groups/:id/students?limit=&offset=
func (h *Handler) GroupStudents(c *gin.Context) {
	groupID, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	limit, offset, ok := parsePage(c)
	if !ok {
		return
	}
	list, total, err := h.svc.GroupStudents(c.Request.Context(), middleware.Me(c), groupID, limit, offset)
	httpx.ReplyPage(c, list, httpx.Meta{Limit: limit, Offset: offset, Total: total}, err)
}

// parsePage — limit (1–200, по умолчанию 50) и offset (>= 0). При ошибке 400 уже отправлен.
func parsePage(c *gin.Context) (limit, offset int, ok bool) {
	limit = 50
	if raw := c.Query("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 200 {
			failQuery(c, "limit", "must be an integer from 1 to 200")
			return 0, 0, false
		}
		limit = v
	}
	if raw := c.Query("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			failQuery(c, "offset", "must be an integer >= 0")
			return 0, 0, false
		}
		offset = v
	}
	return limit, offset, true
}

// Get возвращает полную карточку пользователя (AdminUser) — только администратору.
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

// Create создаёт пользователя: с паролем — сразу активного, без пароля — с приглашением.
// Ответ 201 с AdminUser.
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
	for _, b := range []*studentWriteBody{req.Student.Value} {
		if err := checkStudentBody(b); err != nil {
			httpx.Fail(c, err)
			return
		}
	}
	in := CreateUser{
		Role: req.Role, Password: req.Password,
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
	if in.Role == "" { // прежняя форма без role: роль по переданному профилю
		switch {
		case in.Teacher != nil && in.Student == nil:
			in.Role = RoleTeacher
		case in.Student != nil && in.Teacher == nil:
			in.Role = RoleStudent
		default:
			httpx.Fail(c, fieldErr("role", "required: student, teacher or admin"))
			return
		}
	}
	c.Header("Cache-Control", "no-store")
	u, err := h.svc.Create(c.Request.Context(), in)
	httpx.ReplyCreated(c, u, err)
}

// ResendInvitation отправляет новое приглашение пользователю без пароля.
// Прежние ссылки перестают работать. Ответ 202 с InvitationReceipt.
//
// POST /users/:id/invitations
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
	if err := checkStudentBody(req.Student.Value); err != nil {
		httpx.Fail(c, err)
		return
	}
	in := AdminUpdate{
		Role:     req.Role,
		LastName: req.LastName, FirstName: req.FirstName, MiddleNames: req.MiddleNames,
		Email: req.Email, Passport: req.Passport, Snils: req.Snils,
		Student:      mapField(req.Student, studentWriteBody.role),
		Teacher:      mapField(req.Teacher, teacherWriteBody.role),
		AvatarFileID: req.AvatarFileID, IsActive: req.IsActive,
	}
	c.Header("Cache-Control", "no-store")
	u, err := h.svc.AdminUpdate(c.Request.Context(), middleware.Me(c), id, in)
	httpx.Reply(c, u, err)
}
