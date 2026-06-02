package service

import (
    "context"
    "log/slog"
    
    "urlshortener/internal/apperror"
    "urlshortener/internal/model"
    "urlshortener/pkg/token"

    "golang.org/x/crypto/bcrypt"
    
)

type UserRepository interface {
    Create(ctx context.Context, user *model.User) error
    FindByEmail(ctx context.Context, email string) (*model.User, error)
}

type AuthService struct {
    repo         UserRepository
    tokenManager *token.Manager
}

type TokenPair struct {
    AccessToken  string
    RefreshToken string
}


func NewAuthService(repo UserRepository, tokenManager *token.Manager) *AuthService {
    return &AuthService{repo: repo, tokenManager: tokenManager}
}

func (s *AuthService) Signup(
    ctx context.Context,
    email, password string) (*model.User, *apperror.AppError) {

    existing, err := s.repo.FindByEmail(ctx, email)
    if err != nil {
        return nil, apperror.Internal("could not check email")
    }
    if existing != nil {
        return nil, apperror.BadRequest("email already in use")
    }

    hashed, err := bcrypt.GenerateFromPassword([]byte(password), 12)
    if err != nil {
        return nil, apperror.Internal("could not hash password")
    }

    user := &model.User{
        Email:    email,
        Password: string(hashed),
    }

    if err := s.repo.Create(ctx, user); err != nil {
        return nil, apperror.Internal("could not create user")
    }

    go func(email string) {
        slog.Info("welcome email sent", "email", email)
    }(user.Email)

    return user, nil
}

func (s *AuthService) Login(
    ctx context.Context,
    email, password string) (*TokenPair, *apperror.AppError) {

    user, err := s.repo.FindByEmail(ctx, email)
    if err != nil {
        return nil, apperror.Internal("could not find user")
    }
    if user == nil {
        return nil, apperror.Unauthorized("invalid credentials")
    }

    err = bcrypt.CompareHashAndPassword(
        []byte(user.Password), []byte(password))
    if err != nil {
        return nil, apperror.Unauthorized("invalid credentials")
    }

    accessToken, err := s.tokenManager.GenerateAccessToken(
        user.ID, user.Role)
    if err != nil {
        return nil, apperror.Internal("could not generate access token")
    }

    refreshToken, err := s.tokenManager.GenerateRefreshToken(
        user.ID, user.Role)
    if err != nil {
        return nil, apperror.Internal("could not generate refresh token")
    }

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: refreshToken,
    }, nil
}

func (s *AuthService) Refresh(refreshToken string) (*TokenPair, *apperror.AppError) {
    claims, err := s.tokenManager.Validate(refreshToken)
    if err != nil {
        return nil, apperror.Unauthorized("invalid or expired refresh token")
    }

    if claims.TokenType != token.RefreshToken {
        return nil, apperror.Unauthorized("invalid token type")
    }

    accessToken, err := s.tokenManager.GenerateAccessToken(
        claims.UserID, claims.Role)
    if err != nil {
        return nil, apperror.Internal("could not generate access token")
    }

    newRefreshToken, err := s.tokenManager.GenerateRefreshToken(
        claims.UserID, claims.Role)
    if err != nil {
        return nil, apperror.Internal("could not generate refresh token")
    }

    return &TokenPair{
        AccessToken:  accessToken,
        RefreshToken: newRefreshToken,
    }, nil
}