package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/database"
	"university-api/internal/platform/mail"
	"university-api/internal/platform/middleware"
)

// Срок сессии по документации: 1 сутки, с remember_me — 30. Не продлевается при refresh.
const (
	sessionTTL         = 24 * time.Hour
	sessionTTLRemember = 30 * 24 * time.Hour
)

// Ошибки входа и сессий — коды из документации API.
var (
	errInvalidCredentials = apperr.NewUnauthorized("invalid_credentials", "wrong email or password")
	errAccountPending     = apperr.NewForbidden("account_pending", "registration request is waiting for administrator approval")
	errAccountRejected    = apperr.NewForbidden("account_rejected", "registration request was rejected")
	errAccountInactive    = apperr.NewForbidden("account_inactive", "account is disabled")
	errSessionExpired     = apperr.NewUnauthorized("session_expired", "session has ended, sign in again")
	errRefreshReused      = apperr.NewUnauthorized("refresh_reused", "refresh token was already used; the session has been revoked")
	errCSRFInvalid        = apperr.NewForbidden("csrf_invalid", "missing or invalid CSRF token")
)

// dummyHash — bcrypt-хеш случайной строки. Если email не найден, пароль всё равно
// сравнивается с ним, чтобы время ответа не выдавало, есть ли такой email.
var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password-for-timing"), bcrypt.DefaultCost)

// Options — настройки модуля из конфигурации приложения.
type Options struct {
	Secret         []byte        // ключ подписи JWT и CSRF-токенов (JWT_SECRET)
	AccessTTL      time.Duration // срок access-токена (ACCESS_TTL)
	CookieSecure   bool          // флаг Secure у refresh-cookie
	AllowedOrigins []string      // дополнительные разрешённые Origin для refresh/logout
	AppURL         string        // адрес фронтенда для ссылки активации
	Mail           mail.Sender   // отправка писем-приглашений
}

// Service — вход, сессии и приглашения.
type Service struct {
	repo *repository
	opts Options
}

// NewService создаёт сервис авторизации.
func NewService(repo *repository, opts Options) *Service {
	return &Service{repo: repo, opts: opts}
}

// ===== Вход =====

// Login проверяет email и пароль и открывает сессию.
//
// Сначала проверяется пароль, и только потом сообщается причина недоступности
// аккаунта (pending, rejected, inactive): иначе по ответу можно было бы узнать
// состояние чужого email, не зная пароля.
func (s *Service) Login(ctx context.Context, email, password string, remember bool) (Token, session, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	u, err := s.repo.UserByEmail(ctx, email)
	if errors.Is(err, errNotFound) {
		return Token{}, session{}, s.loginWithoutAccount(ctx, email, password)
	}
	if err != nil {
		return Token{}, session{}, fmt.Errorf("auth: find user: %w", err)
	}

	// Пустой хеш — пароль ещё не задан (приглашение не активировано): войти нельзя.
	if u.PasswordHash == "" {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return Token{}, session{}, errInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return Token{}, session{}, errInvalidCredentials
	}
	if !u.IsActive {
		return Token{}, session{}, errAccountInactive
	}

	ttl := sessionTTL
	if remember {
		ttl = sessionTTLRemember
	}
	sess := session{refreshToken: newToken(), expiresAt: time.Now().Add(ttl)}
	sid, err := s.repo.CreateSession(ctx, u.ID, remember, sess.expiresAt, hashToken(sess.refreshToken))
	if err != nil {
		return Token{}, session{}, fmt.Errorf("auth: create session: %w", err)
	}
	tok, err := s.issue(u.UserInfo, sid, sess.expiresAt)
	return tok, sess, err
}

// loginWithoutAccount — учётки с таким email нет. Если есть заявка и пароль
// совпадает с паролем из неё, сообщаем её статус; иначе — неверные данные.
func (s *Service) loginWithoutAccount(ctx context.Context, email, password string) error {
	status, hash, err := s.repo.LatestRequest(ctx, email)
	if errors.Is(err, errNotFound) {
		_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
		return errInvalidCredentials
	}
	if err != nil {
		return fmt.Errorf("auth: find registration: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) != nil {
		return errInvalidCredentials
	}
	switch status {
	case "pending":
		return errAccountPending
	case "rejected":
		return errAccountRejected
	}
	return errInvalidCredentials
}

// ===== CSRF =====

// CSRF возвращает CSRF-токен сессии по refresh-cookie. Токен вычисляется из id
// сессии и секрета, поэтому стабилен в пределах сессии и не хранится в БД.
func (s *Service) CSRF(ctx context.Context, refreshToken string) (CsrfToken, error) {
	if refreshToken == "" {
		return CsrfToken{}, errSessionExpired
	}
	row, err := s.repo.RefreshByHash(ctx, hashToken(refreshToken))
	if errors.Is(err, errNotFound) {
		return CsrfToken{}, errSessionExpired
	}
	if err != nil {
		return CsrfToken{}, fmt.Errorf("auth: find session: %w", err)
	}
	if !row.active() || row.Used {
		return CsrfToken{}, errSessionExpired
	}
	return CsrfToken{CsrfToken: s.csrfFor(row.SessionID)}, nil
}

// csrfFor — HMAC-SHA256(секрет, "csrf:" + id сессии) в base64url.
func (s *Service) csrfFor(sid string) string {
	m := hmac.New(sha256.New, s.opts.Secret)
	m.Write([]byte("csrf:" + sid))
	return base64.RawURLEncoding.EncodeToString(m.Sum(nil))
}

// csrfValid сравнивает токен за постоянное время.
func (s *Service) csrfValid(sid, token string) bool {
	return token != "" && hmac.Equal([]byte(token), []byte(s.csrfFor(sid)))
}

// ===== Обновление сессии =====

// Refresh обменивает refresh-токен на новую пару (ротация). Активность пользователя
// и роли перечитываются из БД. Срок сессии не продлевается.
//
// Повторное использование уже обменянного токена означает, что его украли:
// сессия отзывается целиком, ответ 401 refresh_reused.
func (s *Service) Refresh(ctx context.Context, refreshToken, csrf string) (Token, session, error) {
	if refreshToken == "" {
		return Token{}, session{}, errSessionExpired
	}
	oldHash := hashToken(refreshToken)
	sess := session{refreshToken: newToken()}

	var (
		user    userRecord
		sid     string
		outcome error // ошибка, о которой сообщаем после фиксации транзакции
	)
	err := database.InTx(ctx, s.repo.db, func(tx pgx.Tx) error {
		row, err := lockRefresh(ctx, tx, oldHash)
		if errors.Is(err, errNotFound) {
			return errSessionExpired
		}
		if err != nil {
			return err
		}
		// CSRF — до любых изменений: чужой сайт не должен ни обновить, ни отозвать сессию.
		if !s.csrfValid(row.SessionID, csrf) {
			return errCSRFInvalid
		}
		if !row.active() {
			return errSessionExpired
		}
		if row.Used {
			outcome = errRefreshReused
			return revokeSession(ctx, tx, row.SessionID, "refresh_reused")
		}

		user, err = userByID(ctx, tx, row.UserID)
		if err != nil {
			return err
		}
		if !user.IsActive {
			outcome = errSessionExpired
			return revokeSession(ctx, tx, row.SessionID, "account_inactive")
		}

		sid, sess.expiresAt = row.SessionID, row.ExpiresAt
		return rotate(ctx, tx, oldHash, hashToken(sess.refreshToken), sid)
	})
	if err != nil {
		return Token{}, session{}, err
	}
	if outcome != nil {
		return Token{}, session{}, outcome
	}
	tok, err := s.issue(user.UserInfo, sid, sess.expiresAt)
	return tok, sess, err
}

// ===== Выход =====

// Logout отзывает текущую сессию. Если сессии уже нет (истекла, отозвана, cookie нет),
// это не ошибка — ответ всё равно 204. Для действующей сессии CSRF обязателен.
func (s *Service) Logout(ctx context.Context, refreshToken, csrf string) error {
	if refreshToken == "" {
		return nil
	}
	row, err := s.repo.RefreshByHash(ctx, hashToken(refreshToken))
	if errors.Is(err, errNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("auth: find session: %w", err)
	}
	if !row.active() {
		return nil
	}
	if !s.csrfValid(row.SessionID, csrf) {
		return errCSRFInvalid
	}
	return s.repo.RevokeSession(ctx, row.SessionID, "logout")
}

// SessionActive — для middleware: принимать ли access-токен этой сессии.
func (s *Service) SessionActive(ctx context.Context, sid string, userID int64) (bool, error) {
	return s.repo.SessionActive(ctx, sid, userID)
}

// ===== Вспомогательное =====

// issue подписывает access-токен сессии и собирает ответ. Роль в токене нужна только
// серверу (проверка прав в middleware); фронт смотрит на user.student/teacher/is_admin.
func (s *Service) issue(u UserInfo, sid string, sessionExpires time.Time) (Token, error) {
	role := middleware.RoleStudent
	switch {
	case u.IsAdmin:
		role = middleware.RoleAdmin
	case u.Teacher != nil:
		role = middleware.RoleTeacher
	}

	now := time.Now()
	expires := now.Add(s.opts.AccessTTL)
	if expires.After(sessionExpires) {
		expires = sessionExpires // access-токен не переживает свою сессию
	}
	claims := middleware.Claims{
		Role: role,
		Sid:  sid,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(u.ID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.opts.Secret)
	if err != nil {
		return Token{}, fmt.Errorf("auth: sign token: %w", err)
	}
	return Token{AccessToken: signed, TokenType: "Bearer", ExpiresAt: expires, User: u}, nil
}

// newToken — 32 случайных байта в base64url (refresh-токены, приглашения).
func newToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic("auth: crypto/rand failed: " + err.Error()) // без случайности токены выдавать нельзя
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// hashToken — SHA-256 токена. В БД хранится только он: утечка базы не даёт рабочих токенов.
func hashToken(t string) []byte {
	h := sha256.Sum256([]byte(t))
	return h[:]
}
