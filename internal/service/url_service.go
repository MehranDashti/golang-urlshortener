package service

import (
    "urlshortener/internal/apperror"
    "urlshortener/internal/model"
    "urlshortener/internal/util"
)

type URLRepository interface {
    Create(url *model.URL) error
    FindByShortCode(code string) (*model.URL, error)
    IncrementClicks(id string) error
}

type URLService struct {
    repo URLRepository
}

func NewURLService(repo URLRepository) *URLService {
    return &URLService{repo: repo}
}

func (s *URLService) ShortenURL(originalURL string) (*model.URL, *apperror.AppError) {
    url := &model.URL{
        OriginalURL: originalURL,
        ShortCode:   util.GenerateShortCode(),
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

    s.repo.IncrementClicks(url.ID)
    return url, nil
}