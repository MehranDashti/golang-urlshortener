package service

import (
    "urlshortener/internal/apperror"
    "urlshortener/internal/model"
)

type AdminURLRepository interface {
    FindAll() ([]*model.URL, error)
    Delete(id string) error
}

type AdminUserRepository interface {
    FindAll() ([]*model.User, error)
}

type AdminService struct {
    urlRepo  AdminURLRepository
    userRepo AdminUserRepository
}

func NewAdminService(urlRepo AdminURLRepository, userRepo AdminUserRepository) *AdminService {
    return &AdminService{urlRepo: urlRepo, userRepo: userRepo}
}

func (s *AdminService) GetAllLinks() ([]*model.URL, *apperror.AppError) {
    urls, err := s.urlRepo.FindAll()
    if err != nil {
        return nil, apperror.Internal("could not fetch links")
    }
    return urls, nil
}

func (s *AdminService) DeleteLink(id string) *apperror.AppError {
    if err := s.urlRepo.Delete(id); err != nil {
        return apperror.Internal("could not delete link")
    }
    return nil
}

func (s *AdminService) GetAllUsers() ([]*model.User, *apperror.AppError) {
    users, err := s.userRepo.FindAll()
    if err != nil {
        return nil, apperror.Internal("could not fetch users")
    }
    return users, nil
}