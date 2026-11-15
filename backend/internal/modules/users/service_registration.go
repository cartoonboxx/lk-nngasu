package users

import (
	"context"
	"strings"
	"time"
	"unicode"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/middleware"
)

// Регистрация через заявку (раздел «Заявки на регистрацию» документации API).
//
//  1. Человек отправляет заявку: email, пароль, паспорт, СНИЛС, согласие с условиями
//     (POST /auth/register). Учётки ещё нет: вход отвечает 403 account_pending.
//  2. Администратор видит очередь (GET /registration-requests), открывает заявку,
//     уточняет ФИО и документы, выбирает ровно одну учебную роль и одобряет
//     (POST /registration-requests/:id/decision, decision = approve). В одной
//     транзакции создаётся пользователь с email и паролем из заявки, заявка закрывается.
//  3. Или отклоняет (decision = reject) — учётка не создаётся, вход отвечает
//     403 account_rejected, для доступа нужна новая заявка.
//
// Роль пользователь не выбирает, администраторские права так не выдаются.

// RegistrationRequest — заявка в административной схеме (список, карточка, ответ на решение).
// Паспорт и СНИЛС — только цифры. Хеш пароля наружу не отдаётся.
type RegistrationRequest struct {
	ID          int64      `json:"id"`
	Email       string     `json:"email"`
	Passport    string     `json:"passport"` // "4509123456"
	Snils       string     `json:"snils"`    // "11223344595"
	LastName    *string    `json:"last_name"`
	FirstName   *string    `json:"first_name"`
	MiddleNames *string    `json:"middle_names"`
	Comment     *string    `json:"comment"`
	Status      string     `json:"status"` // pending | approved | rejected
	CreatedAt   time.Time  `json:"created_at"`
	ReviewedAt  *time.Time `json:"reviewed_at"`
	ReviewedBy  *PersonRef `json:"reviewed_by"` // администратор, принявший решение
	UserID      *int64     `json:"user_id"`     // созданный пользователь, если заявка одобрена
}

// PersonRef — безопасная ссылка на человека: только id и ФИО, без email и документов.
type PersonRef struct {
	ID   int64  `json:"id"`
	Name string `json:"name"` // ФИО
}

// RegistrationReceipt — ответ на подачу заявки. Ни email, ни паспорта и СНИЛС:
// документы доступны только администраторам, обрабатывающим заявки.
type RegistrationReceipt struct {
	ID        int64     `json:"id"`
	Status    string    `json:"status"` // pending
	CreatedAt time.Time `json:"created_at"`
}

// SubmitRequest — данные заявки от пользователя.
type SubmitRequest struct {
	Email, Password, Passport, Snils string
	Comment                          *string // необязательный комментарий для администратора
	AcceptedTerms                    bool
	TermsVersion                     string
}

// RequestFilter — параметры списка заявок.
type RequestFilter struct {
	Q      string // подстрока email или ФИО
	Status string // pending | approved | rejected | "" (все)
	Newest bool   // true — новые сверху (sort=-created_at, по умолчанию)
	Limit  int
	Offset int
}

// Decision — решение администратора по заявке.
type Decision struct {
	Approve bool
	User    *ApprovalUser // только для approve
}

// ApprovalUser — данные пользователя при одобрении (ApprovalUserInput).
// Email и пароль берутся из заявки. Ровно одна учебная роль: Student или Teacher.
type ApprovalUser struct {
	LastName, FirstName string
	MiddleNames         *string
	Passport, Snils     string
	Student             *StudentRole // StudentWrite
	Teacher             *TeacherRole // TeacherWrite
}

// fieldErr — 400 invalid_request с описанием в details.fields, как требует документация API.
func fieldErr(field, problem string) error {
	return apperr.NewInvalid("invalid_request", field+": "+problem).
		WithDetails(map[string]any{"fields": map[string][]string{field: {problem}}})
}

var (
	errTermsNotAccepted = fieldErr("accepted_terms", "must be true")
	errOneRole          = apperr.NewInvalid("invalid_role", "exactly one of user.student or user.teacher must be set")
)

// digitsOnly оставляет в строке только цифры: пользователь может ввести
// "45 10 200001" или "200-000-001 00" — в БД попадёт "4510200001" и "20000000100".
func digitsOnly(s string) string {
	var b strings.Builder
	for _, r := range s {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// normPassport и normSnils нормализуют и проверяют длину. field — имя поля для
// details.fields (у одобрения поля вложены в user). Контрольное число СНИЛС
// пока не проверяется: тестовые данные часто его не соблюдают.
func normPassport(field, s string) (string, error) {
	d := digitsOnly(s)
	if len(d) != 10 {
		return "", fieldErr(field, "must contain 10 digits: series and number")
	}
	return d, nil
}

func normSnils(field, s string) (string, error) {
	d := digitsOnly(s)
	if len(d) != 11 {
		return "", fieldErr(field, "must contain 11 digits")
	}
	return d, nil
}

// SubmitRegistration сохраняет заявку. Пароль сразу хешируется: в открытом виде
// он нигде не хранится, при одобрении хеш переносится в учётку как есть.
// Ответ — квитанция без документов.
func (s *Service) SubmitRegistration(ctx context.Context, in SubmitRequest) (RegistrationReceipt, error) {
	if !in.AcceptedTerms {
		return RegistrationReceipt{}, errTermsNotAccepted
	}
	// Фронт берёт версию условий из своей конфигурации; если она не совпадает
	// с серверной, пользователь соглашался не с теми условиями.
	if in.TermsVersion != s.opts.TermsVersion {
		return RegistrationReceipt{}, fieldErr("terms_version", "unsupported version, current is "+s.opts.TermsVersion)
	}
	if len(in.Password) > maxPasswordBytes {
		return RegistrationReceipt{}, fieldErr("password", "must be at most 72 bytes in UTF-8")
	}
	passport, err := normPassport("passport", in.Passport)
	if err != nil {
		return RegistrationReceipt{}, err
	}
	snils, err := normSnils("snils", in.Snils)
	if err != nil {
		return RegistrationReceipt{}, err
	}
	if in.Comment != nil && strings.TrimSpace(*in.Comment) == "" {
		in.Comment = nil
	}
	hash, err := hashPassword(in.Password)
	if err != nil {
		return RegistrationReceipt{}, err
	}
	in.Email = normEmail(in.Email)
	in.Passport, in.Snils = passport, snils

	return s.repo.CreateRequest(ctx, in, hash)
}

// ListRegistrations — страница заявок и общее число под фильтрами.
func (s *Service) ListRegistrations(ctx context.Context, f RequestFilter) ([]RegistrationRequest, int, error) {
	f.Q = strings.TrimSpace(f.Q)
	return s.repo.ListRequests(ctx, f)
}

// GetRegistration — одна заявка.
func (s *Service) GetRegistration(ctx context.Context, id int64) (RegistrationRequest, error) {
	return s.repo.GetRequest(ctx, id)
}

// DecideRegistration — одобрить с созданием пользователя или отклонить.
// Возвращает заявку в итоговом состоянии.
func (s *Service) DecideRegistration(ctx context.Context, admin middleware.Principal, id int64, d Decision) (RegistrationRequest, error) {
	if d.Approve {
		u := d.User
		if (u.Student == nil) == (u.Teacher == nil) {
			return RegistrationRequest{}, errOneRole
		}
		var err error
		if u.LastName, err = normName("user.last_name", u.LastName); err != nil {
			return RegistrationRequest{}, err
		}
		if u.FirstName, err = normName("user.first_name", u.FirstName); err != nil {
			return RegistrationRequest{}, err
		}
		if u.MiddleNames, err = normMiddle("user.middle_names", u.MiddleNames); err != nil {
			return RegistrationRequest{}, err
		}
		if err = normRecordBook("user.student.record_book_number", u.Student); err != nil {
			return RegistrationRequest{}, err
		}
		if u.Passport, err = normPassport("user.passport", u.Passport); err != nil {
			return RegistrationRequest{}, err
		}
		if u.Snils, err = normSnils("user.snils", u.Snils); err != nil {
			return RegistrationRequest{}, err
		}
		if err := s.repo.ApproveRequest(ctx, id, admin.UserID, *u); err != nil {
			return RegistrationRequest{}, err
		}
	} else if err := s.repo.RejectRequest(ctx, id, admin.UserID); err != nil {
		return RegistrationRequest{}, err
	}
	return s.repo.GetRequest(ctx, id)
}
