package auth

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"university-api/internal/platform/apperr"
	"university-api/internal/platform/middleware"
)

// errInvalidCredentials возвращается одинаково, когда пользователь не найден,
// неактивен или пароль неверный. Так по ответу нельзя узнать,
// зарегистрирован ли email. Отдаётся клиенту как 401.
var errInvalidCredentials = apperr.NewUnauthorized("invalid_credentials", "wrong email or password")

// Token — результат успешного входа.
type Token struct {
	AccessToken string    `json:"access_token"` // подписанный JWT
	TokenType   string    `json:"token_type"`   // всегда "Bearer"
	ExpiresAt   time.Time `json:"expires_at"`   // когда токен перестанет действовать
	Role        string    `json:"role"`         // дублирует роль из токена, чтобы фронту не разбирать JWT
}

// Service — бизнес-логика входа: проверяет учётные данные,
// определяет роль и выпускает JWT. Права доступа здесь не проверяются —
// это делает middleware по роли из токена.
type Service struct {
	repo   Repository
	secret []byte        // ключ подписи JWT (HS256), тот же использует middleware
	ttl    time.Duration // время жизни токена
}

// NewService создаёт сервис авторизации.
func NewService(repo Repository, secret []byte, ttl time.Duration) *Service {
	return &Service{repo: repo, secret: secret, ttl: ttl}
}

// Login проверяет email и пароль и при успехе возвращает JWT.
// При неверных данных — errInvalidCredentials, при сбое БД — внутренняя ошибка.
func (s *Service) Login(ctx context.Context, email, password string) (Token, error) {
	// Нормализуем email: репозиторий сравнивает его с учётом регистра.
	u, err := s.repo.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))

	// Пользователь не найден — не раскрываем это, отвечаем как на неверный пароль.
	if errors.Is(err, errUserNotFound) {
		return Token{}, errInvalidCredentials
	}
	// Любая другая ошибка — внутренняя (например, недоступна БД), клиент получит 500.
	if err != nil {
		return Token{}, err
	}

	// Неактивный пользователь и неверный пароль дают ту же ошибку, что и выше.
	if !u.IsActive || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return Token{}, errInvalidCredentials
	}

	// Роль одна. Порядок проверок задаёт приоритет: admin > teacher > student.
	// Администратор — надмножество: ему доступны все управляющие операции.
	role := middleware.RoleStudent
	switch {
	case u.IsAdmin:
		role = middleware.RoleAdmin
	case u.IsTeacher:
		role = middleware.RoleTeacher
	}

	now := time.Now()         // момент выдачи токена
	expires := now.Add(s.ttl) // момент истечения токена

	claims := middleware.Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(u.ID, 10), // sub — ID пользователя (по спецификации JWT строка)
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}

	// Подписываем HS256 общим секретом; middleware проверит подпись тем же ключом.
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return Token{}, err
	}
	return Token{AccessToken: signed, TokenType: "Bearer", ExpiresAt: expires, Role: role}, nil
}
