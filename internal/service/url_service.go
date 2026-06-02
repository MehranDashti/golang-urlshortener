package service

import (
    "context"
    "time"

    "urlshortener/internal/apperror"
    "urlshortener/internal/model"
    "urlshortener/internal/util"
)

type URLRepository interface {
    Create(ctx context.Context, url *model.URL) error
    FindByShortCode(ctx context.Context, code string) (*model.URL, error)
    IncrementClicks(ctx context.Context, id string) error
    FindByUserID(ctx context.Context, userID string) ([]*model.URL, error)
}

type URLService struct {
    repo URLRepository
}

func NewURLService(repo URLRepository) *URLService {
    return &URLService{repo: repo}
}

func (s *URLService) ShortenURL(
    ctx context.Context,
    originalURL string,
    userID string,
    expiresAt *time.Time) (*model.URL, *apperror.AppError) {

    url := &model.URL{
        UserID:      userID,
        OriginalURL: originalURL,
        ShortCode:   util.GenerateShortCode(),
        ExpiresAt:   expiresAt,
    }

    if err := s.repo.Create(ctx, url); err != nil {
        return nil, apperror.Internal("could not create short url")
    }
    return url, nil
}

func (s *URLService) GetByShortCode(
    ctx context.Context,
    code string) (*model.URL, *apperror.AppError) {

    url, err := s.repo.FindByShortCode(ctx, code)
    if err != nil {
        return nil, apperror.Internal("something went wrong")
    }
    if url == nil {
        return nil, apperror.NotFound("short url not found")
    }
    if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
        return nil, apperror.Gone("short url has expired")
    }

    s.repo.IncrementClicks(ctx, url.ID)
    return url, nil
}

func (s *URLService) GetUserLinks(
    ctx context.Context,
    userID string) ([]*model.URL, *apperror.AppError) {

    urls, err := s.repo.FindByUserID(ctx, userID)
    if err != nil {
        return nil, apperror.Internal("could not fetch links")
    }
    return urls, nil
}