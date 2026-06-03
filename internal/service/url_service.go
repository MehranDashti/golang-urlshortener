package service

import (
    "log/slog"
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
    FindByUserIDPaginated(ctx context.Context, userID string,
        params model.PaginationParams) ([]*model.URL, int64, error)
}

type URLService struct {
    repo URLRepository
    clickCh   chan string
}

func NewURLService(repo URLRepository) *URLService {
    s := &URLService{
        repo:    repo,
        clickCh: make(chan string, 100), // buffered — non-blocking
    }
    go s.clickWorker() // start background worker
    return s
}

// clickWorker drains the click channel in the background
func (s *URLService) clickWorker() {
    for id := range s.clickCh {
        // context.Background() — not tied to any request
        ctx, cancel := context.WithTimeout(
            context.Background(), 5*time.Second)
        s.repo.IncrementClicks(ctx, id)
        cancel()
    }
}

func (s *URLService) Close() {
    close(s.clickCh) // closing the channel stops the range loop in clickWorker
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

    // Non-blocking click increment using select
    // If channel is full (100 pending) — drop the click rather than block
    select {
    case s.clickCh <- url.ID:
        // sent to worker — will be processed asynchronously
    default:
        // channel full — skip this click increment
        // better to lose a click than to slow down a redirect
        slog.Warn("click channel full, dropping increment",
            "url_id", url.ID)
    }

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

func (s *URLService) GetUserLinksPaginated(
    ctx context.Context,
    userID string,
    params model.PaginationParams) (*model.PaginatedResult, *apperror.AppError) {

    urls, total, err := s.repo.FindByUserIDPaginated(
        ctx, userID, params)
    if err != nil {
        return nil, apperror.Internal("could not fetch links")
    }

    result := model.NewPaginatedResult(urls, total, params)
    return &result, nil
}