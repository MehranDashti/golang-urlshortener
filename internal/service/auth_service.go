package service

import (
    "urlshortener/internal/apperror"
    "urlshortener/internal/model"
    "urlshortener/pkg/token"

    "golang.org/x/crypto/bcrypt"
)

// UserRepository is the interface the auth service needs.
type UserRepository interface {
    Create(user *model.User) error
    FindByEmail(email string) (*model.User, error)
}

type AuthService struct {
    repo         UserRepository
    tokenManager *token.Manager
}

func NewAuthService(repo UserRepository, tokenManager *token.Manager) *AuthService {
    return &AuthService{repo: repo, tokenManager: tokenManager}
}

func (s *AuthService) Signup(email, password string) (*model.User, *apperror.AppError) {
    existing, err := s.repo.FindByEmail(email)
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

    if err := s.repo.Create(user); err != nil {
        return nil, apperror.Internal("could not create user")
    }

    return user, nil
}

func (s *AuthService) Login(email, password string) (string, *apperror.AppError) {
    user, err := s.repo.FindByEmail(email)
    if err != nil {
        return "", apperror.Internal("could not find user")
    }

    if user == nil {
        return "", apperror.Unauthorized("invalid credentials")
    }

    err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
    if err != nil {
        return "", apperror.Unauthorized("invalid credentials")
    }

    tokenStr, err := s.tokenManager.Generate(user.ID)
    if err != nil {
        return "", apperror.Internal("could not generate token")
    }

    return tokenStr, nil
}