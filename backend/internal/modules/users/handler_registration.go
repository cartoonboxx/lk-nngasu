package users

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/httpx"
	"university-api/internal/platform/middleware"
)

// submitRequestBody — тело POST /auth/register (RegistrationInput).
// «Повтори пароль» проверяет фронт, на сервер оно не отправляется.
type submitRequestBody struct {
	Email         string `json:"email" binding:"required,email,max=255"`
	Password      string `json:"password" binding:"required,min=8,max=72"` // и не больше 72 байт — проверяет сервис
	Passport      string `json:"passport" binding:"required,max=20"`       // "45 09 123456" — пробелы допустимы
	Snils         string `json:"snils" binding:"required,max=20"`          // "112-233-445 95" — формат формы допустим
	AcceptedTerms bool   `json:"accepted_terms"`                           // должно быть true
	TermsVersion  string `json:"terms_version" binding:"required,max=32"`  // версия условий из конфигурации фронта
}

// decisionBody — тело POST /registration-requests/:id/decision (RegistrationDecision).
// Для reject поле user запрещено, для approve обязательно.
type decisionBody struct {
	Decision string             `json:"decision" binding:"required,oneof=approve reject"`
	User     *approvalUserInput `json:"user"`
}

// approvalUserInput — ApprovalUserInput: ФИО, уточнённые документы и роль.
// Ключи student и teacher обязательны (значение может быть null),
// при этом ровно одна роль должна быть объектом — это проверяет сервис.
type approvalUserInput struct {
	LastName    string                  `json:"last_name" binding:"required,max=50"`
	FirstName   string                  `json:"first_name" binding:"required,max=50"`
	MiddleNames *string                 `json:"middle_names" binding:"omitempty,max=50"`
	Passport    string                  `json:"passport" binding:"required,max=20"`
	Snils       string                  `json:"snils" binding:"required,max=20"`
	Student     Field[studentWriteBody] `json:"student"`
	Teacher     Field[teacherWriteBody] `json:"teacher"`
}

// SubmitRegistration — отправка заявки на регистрацию. Маршрут публичный.
// Ответ 201 с квитанцией {id, status, created_at}. Войти можно только после одобрения.
//
// POST /auth/register
func (h *Handler) SubmitRegistration(c *gin.Context) {
	var req submitRequestBody
	if !httpx.Bind(c, &req) {
		return
	}
	r, err := h.svc.SubmitRegistration(c.Request.Context(), SubmitRequest{
		Email: req.Email, Password: req.Password, Passport: req.Passport, Snils: req.Snils,
		AcceptedTerms: req.AcceptedTerms, TermsVersion: req.TermsVersion,
	})
	httpx.ReplyCreated(c, r, err)
}

// ListRegistrations — очередь заявок для администратора: {data: [...], meta: {limit, offset, total}}.
//
// GET /registration-requests?q=&status=pending|approved|rejected&sort=-created_at&limit=50&offset=0
func (h *Handler) ListRegistrations(c *gin.Context) {
	c.Header("Cache-Control", "no-store") // в ответе паспорта и СНИЛС

	f := RequestFilter{Q: c.Query("q"), Status: c.Query("status"), Newest: true, Limit: 50}
	switch f.Status {
	case "", "pending", "approved", "rejected":
	default:
		failQuery(c, "status", "must be one of: pending, approved, rejected")
		return
	}
	switch c.Query("sort") {
	case "", "-created_at":
	case "created_at":
		f.Newest = false
	default:
		failQuery(c, "sort", "must be created_at or -created_at")
		return
	}
	if raw := c.Query("limit"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 1 || v > 200 {
			failQuery(c, "limit", "must be an integer from 1 to 200")
			return
		}
		f.Limit = v
	}
	if raw := c.Query("offset"); raw != "" {
		v, err := strconv.Atoi(raw)
		if err != nil || v < 0 {
			failQuery(c, "offset", "must be an integer >= 0")
			return
		}
		f.Offset = v
	}

	list, total, err := h.svc.ListRegistrations(c.Request.Context(), f)
	httpx.ReplyPage(c, list, httpx.Meta{Limit: f.Limit, Offset: f.Offset, Total: total}, err)
}

// failQuery — 400 invalid_query с описанием параметра в details.fields.
func failQuery(c *gin.Context, name, problem string) {
	httpx.Fail(c, apperr.NewInvalid("invalid_query", "invalid query parameter: "+name).
		WithDetails(map[string]any{"fields": map[string][]string{name: {problem}}}))
}

// GetRegistration — одна заявка (окно «Заявка №…»).
//
// GET /registration-requests/:id
func (h *Handler) GetRegistration(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	r, err := h.svc.GetRegistration(c.Request.Context(), id)
	httpx.Reply(c, r, err)
}

// DecideRegistration — одобрить с созданием пользователя или отклонить.
// Ответ 200 с заявкой в итоговом состоянии.
//
// POST /registration-requests/:id/decision
func (h *Handler) DecideRegistration(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	id, ok := httpx.ID(c, "id")
	if !ok {
		return
	}
	var req decisionBody
	if !httpx.Bind(c, &req) {
		return
	}

	d := Decision{Approve: req.Decision == "approve"}
	switch {
	case !d.Approve && req.User != nil:
		httpx.Fail(c, fieldErr("user", "must be absent for reject"))
		return
	case d.Approve && req.User == nil:
		httpx.Fail(c, fieldErr("user", "required for approve"))
		return
	case d.Approve && !req.User.Student.Set:
		httpx.Fail(c, fieldErr("user.student", "required (null if the user is not a student)"))
		return
	case d.Approve && !req.User.Teacher.Set:
		httpx.Fail(c, fieldErr("user.teacher", "required (null if the user is not a teacher)"))
		return
	case d.Approve:
		u := req.User
		d.User = &ApprovalUser{
			LastName: u.LastName, FirstName: u.FirstName, MiddleNames: u.MiddleNames,
			Passport: u.Passport, Snils: u.Snils,
		}
		if u.Student.Value != nil {
			r := u.Student.Value.role()
			d.User.Student = &r
		}
		if u.Teacher.Value != nil {
			r := u.Teacher.Value.role()
			d.User.Teacher = &r
		}
	}

	r, err := h.svc.DecideRegistration(c.Request.Context(), middleware.Me(c), id, d)
	httpx.Reply(c, r, err)
}
