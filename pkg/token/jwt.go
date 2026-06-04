package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"urlshortener/internal/model"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims is the JWT payload.
// jwt.RegisteredClaims includes the ID field (JTI) — unique per token.
type Claims struct {
	UserID    string     `json:"user_id"`
	Role      model.Role `json:"role"`
	TokenType TokenType  `json:"token_type"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret               []byte
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

func NewManager(
	secret string,
	accessDuration, refreshDuration time.Duration) *Manager {
	return &Manager{
		secret:               []byte(secret),
		accessTokenDuration:  accessDuration,
		refreshTokenDuration: refreshDuration,
	}
}

func (m *Manager) GenerateAccessToken(
	userID string, role model.Role) (string, error) {
	return m.generate(userID, role, AccessToken, m.accessTokenDuration)
}

func (m *Manager) GenerateRefreshToken(
	userID string, role model.Role) (string, error) {
	return m.generate(userID, role, RefreshToken, m.refreshTokenDuration)
}

// generate is the internal helper — creates a signed JWT with a unique JTI.
func (m *Manager) generate(
	userID string,
	role model.Role,
	tokenType TokenType,
	duration time.Duration) (string, error) {

	claims := &Claims{
		UserID:    userID,
		Role:      role,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.New().String(), // JTI — unique per token, used for blacklist
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) Validate(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr, &Claims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return m.secret, nil
		})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}