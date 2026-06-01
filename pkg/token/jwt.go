package token

import (
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "urlshortener/internal/model"
)

type Claims struct {
    UserID string     `json:"user_id"`
    Role   model.Role `json:"role"`
    jwt.RegisteredClaims
}

type Manager struct {
    secret   []byte
    duration time.Duration
}

func NewManager(secret string, duration time.Duration) *Manager {
    return &Manager{
        secret:   []byte(secret),
        duration: duration,
    }
}

func (m *Manager) Generate(userID string, role model.Role) (string, error) {
    claims := &Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.duration)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(m.secret)
}

func (m *Manager) Validate(tokenStr string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
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