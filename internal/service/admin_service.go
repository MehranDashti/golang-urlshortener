package service

import (
    "context"

    "urlshortener/internal/apperror"
    "urlshortener/internal/model"
)

type AdminURLRepository interface {
    FindAll(ctx context.Context) ([]*model.URL, error)
    Delete(ctx context.Context, id string) error
}

type AdminUserRepository interface {
    FindAll(ctx context.Context) ([]*model.User, error)
}

// ← this was missing
type AdminService struct {
    urlRepo  AdminURLRepository
    userRepo AdminUserRepository
}

// ← this was missing
func NewAdminService(urlRepo AdminURLRepository, userRepo AdminUserRepository) *AdminService {
    return &AdminService{urlRepo: urlRepo, userRepo: userRepo}
}

func (s *AdminService) GetAllLinks(
    ctx context.Context) ([]*model.URL, *apperror.AppError) {
    urls, err := s.urlRepo.FindAll(ctx)
    if err != nil {
        return nil, apperror.Internal("could not fetch links")
    }
    return urls, nil
}

func (s *AdminService) DeleteLink(
    ctx context.Context, id string) *apperror.AppError {
    if err := s.urlRepo.Delete(ctx, id); err != nil {
        return apperror.Internal("could not delete link")
    }
    return nil
}

func (s *AdminService) GetAllUsers(
    ctx context.Context) ([]*model.User, *apperror.AppError) {
    users, err := s.userRepo.FindAll(ctx)
    if err != nil {
        return nil, apperror.Internal("could not fetch users")
    }
    return users, nil
}