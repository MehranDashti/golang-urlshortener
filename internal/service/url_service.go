package service

import (
    "time"

    "urlshortener/internal/apperror"
    "urlshortener/internal/model"
    "urlshortener/internal/util"
)

type URLRepository interface {
    Create(url *model.URL) error
    FindByShortCode(code string) (*model.URL, error)
    IncrementClicks(id string) error
    FindByUserID(userID string) ([]*model.URL, error)
}

type URLService struct {
    repo URLRepository
}

func NewURLService(repo URLRepository) *URLService {
    return &URLService{repo: repo}
}

func (s *URLService) ShortenURL(originalURL string, userID string, expiresAt *time.Time) (*model.URL, *apperror.AppError) {
    url := &model.URL{
        UserID:      userID,
        OriginalURL: originalURL,
        ShortCode:   util.GenerateShortCode(),
        ExpiresAt:   expiresAt,
    }

    if err := s.repo.Create(url); err != nil {
        return nil, apperror.Internal("could not create short url")
    }

    return url, nil
}

func (s *URLService) GetByShortCode(code string) (*model.URL, *apperror.AppError) {
    url, err := s.repo.FindByShortCode(code)
    if err != nil {
        return nil, apperror.Internal("something went wrong")
    }
    if url == nil {
        return nil, apperror.NotFound("short url not found")
    }
    if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
        return nil, apperror.Gone("short url has expired")
    }

    s.repo.IncrementClicks(url.ID)
    return url, nil
}

func (s *URLService) GetUserLinks(userID string) ([]*model.URL, *apperror.AppError) {
    urls, err := s.repo.FindByUserID(userID)
    if err != nil {
        return nil, apperror.Internal("could not fetch links")
    }
    return urls, nil
}