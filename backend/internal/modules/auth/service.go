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

var errInvalidCredentials = apperr.NewUnauthorized("invalid_credentials", "wrong email or password")

// Token — результат успешного входа.
type Token struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
	Role        string    `json:"role"`
}

type Service struct {
	repo   Repository
	secret []byte
	ttl    time.Duration
}

func NewService(repo Repository, secret []byte, ttl time.Duration) *Service {
	return &Service{repo: repo, secret: secret, ttl: ttl}
}

func (s *Service) Login(ctx context.Context, email, password string) (Token, error) {
	u, err := s.repo.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, errUserNotFound) {
		return Token{}, errInvalidCredentials
	}
	if err != nil {
		return Token{}, err
	}
	if !u.IsActive || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return Token{}, errInvalidCredentials
	}

	// Администратор — надмножество: ему доступны все управляющие операции.
	role := middleware.RoleStudent
	switch {
	case u.IsAdmin:
		role = middleware.RoleAdmin
	case u.IsTeacher:
		role = middleware.RoleTeacher
	}

	now := time.Now()
	expires := now.Add(s.ttl)
	claims := middleware.Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(u.ID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expires),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return Token{}, err
	}
	return Token{AccessToken: signed, TokenType: "Bearer", ExpiresAt: expires, Role: role}, nil
}
