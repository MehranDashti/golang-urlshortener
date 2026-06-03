package service

import (
    "log/slog"
    "context"
    "time"
    "sync"

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
    recentIDs sync.Map
}
func NewURLService(repo URLRepository) *URLService {
    s := &URLService{
        repo:    repo,
        clickCh: make(chan string, 100),
    }
    go s.clickWorker()
    return s
}

func (s *URLService) clickWorker() {
    for id := range s.clickCh {
        ctx, cancel := context.WithTimeout(
            context.Background(), 5*time.Second)
        s.repo.IncrementClicks(ctx, id)
        cancel()

        // Store the click time — sync.Map is safe for concurrent use
        s.recentIDs.Store(id, time.Now())
    }
}

// cleanRecentIDs removes entries older than 1 minute
// Called periodically — prevents recentIDs growing forever
func (s *URLService) cleanRecentIDs() {
    s.recentIDs.Range(func(key, value interface{}) bool {
        t := value.(time.Time)
        if time.Since(t) > time.Minute {
            s.recentIDs.Delete(key)
        }
        return true // continue ranging
    })
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
        return nil, apperror.InternalWithErr("could not create short url", err)
    }
    return url, nil
}

func (s *URLService) GetByShortCode(
    ctx context.Context,
    code string) (*model.URL, *apperror.AppError) {

    url, err := s.repo.FindByShortCode(ctx, code)
    if err != nil {
        return nil, apperror.InternalWithErr("something went wrong", err)
    }
    if url == nil {
        return nil, apperror.NotFound("short url not found")
    }
    if url.ExpiresAt != nil && time.Now().After(*url.ExpiresAt) {
        return nil, apperror.Gone("short url has expired")
    }

    // Non-blocking click increment using select
    // If channel is full (100 pending) — drop the click rather than block
    if _, alreadyClicked := s.recentIDs.Load(url.ID); !alreadyClicked {
        select {
        case s.clickCh <- url.ID:
        default:
            slog.Warn("click channel full, dropping increment",
                "url_id", url.ID)
        }
    }

    return url, nil
}

func (s *URLService) GetUserLinks(
    ctx context.Context,
    userID string) ([]*model.URL, *apperror.AppError) {

    urls, err := s.repo.FindByUserID(ctx, userID)
    if err != nil {
        return nil, apperror.InternalWithErr("could not fetch links", err)
    }
    return urls, nil
}

func (s *URLService) GetUserLinksPaginated(
    ctx context.Context,
    userID string,
    params model.PaginationParams) (*model.PaginatedResult[*model.URL], *apperror.AppError) {

    urls, total, err := s.repo.FindByUserIDPaginated(
        ctx, userID, params)
    if err != nil {
        slog.Error("GetUserLinksPaginated failed",
            "error", err, "userID", userID)
        return nil, apperror.InternalWithErr(
            "could not fetch links", err)
    }

    result := model.NewPaginatedResult(urls, total, params)
    return &result, nil
}