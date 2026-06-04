package service

import (
    "encoding/csv"
    "io"
    "fmt"
    "log/slog"
    "context"
    "time"
    "sync"
    "gorm.io/gorm"
    
    "urlshortener/internal/apperror"
    "urlshortener/internal/worker"
    "urlshortener/internal/model"
    "urlshortener/internal/util"
    "urlshortener/internal/repository"
)

type URLRepository interface {
    Create(ctx context.Context, url *model.URL) error
    FindByShortCode(ctx context.Context, code string) (*model.URL, error)
    IncrementClicks(ctx context.Context, id string) error
    FindByUserID(ctx context.Context, userID string) ([]*model.URL, error)
    FindByUserIDPaginated(ctx context.Context, userID string,
        params model.PaginationParams) ([]*model.URL, int64, error)
    WithTx(tx *gorm.DB) *repository.URLRepository
    DB() *gorm.DB                           
}

type URLService struct {
    repo URLRepository
    clickCh   chan string
    recentIDs sync.Map
}

// BulkShortenResult holds the result of one bulk shorten operation
type BulkShortenResult struct {
    OriginalURL string
    ShortCode   string
    Error       string
}

// ImportResult holds the outcome of one import row
type ImportResult struct {
    Row         int
    OriginalURL string
    ShortCode   string
    Error       string
}

func NewURLService(
    repo URLRepository,
    ctx context.Context) *URLService {
    s := &URLService{
        repo:    repo,
        clickCh: make(chan string, 100),
    }
    go s.clickWorker(ctx) // pass context — worker exits when ctx cancelled
    return s
}

func (s *URLService) clickWorker(ctx context.Context) {
    for {
        select {
        case id, ok := <-s.clickCh:
            if !ok {
                return // channel closed — exit cleanly
            }
            workCtx, cancel := context.WithTimeout(
                context.Background(), 5*time.Second)
            s.repo.IncrementClicks(workCtx, id)
            cancel()

        case <-ctx.Done():
            return // context cancelled — exit cleanly
        }
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

const maxShortCodeAttempts = 5

func (s *URLService) ShortenURL(
    ctx context.Context,
    originalURL string,
    userID string,
    expiresAt *time.Time) (*model.URL, *apperror.AppError) {

    url := &model.URL{
        UserID:      userID,
        OriginalURL: originalURL,
        ExpiresAt:   expiresAt,
    }

    // Retry up to maxShortCodeAttempts times on collision
    for attempt := 1; attempt <= maxShortCodeAttempts; attempt++ {
        url.ShortCode = util.GenerateShortCode()

        err := s.repo.Create(ctx, url)
        if err == nil {
            return url, nil // success
        }

        // Unwrap to check if it's a duplicate key error
        if repository.IsDuplicateKeyError(err) {
            slog.Warn("short code collision — retrying",
                "attempt", attempt,
                "code",    url.ShortCode,
            )
            continue // try again with a new code
        }

        // Any other error — don't retry
        slog.Error("ShortenURL failed",
            "error",  err,
            "url",    originalURL,
            "userID", userID,
        )
        return nil, apperror.InternalWithErr(
            "could not create short url", err)
    }

    // All attempts exhausted
    slog.Error("ShortenURL failed — max collision attempts reached",
        "attempts", maxShortCodeAttempts,
        "url",      originalURL,
    )
    return nil, apperror.Internal(
        "could not generate unique short code — please try again")
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

func (s *URLService) BulkShorten(
    ctx context.Context,
    urls []string,
    userID string,
    numWorkers int,
) []BulkShortenResult {

    results := make([]BulkShortenResult, 0, len(urls))

    // Wrap all creates in a single transaction
    // If any URL fails to create — all are rolled back
    err := s.repo.DB().WithContext(ctx).
        Transaction(func(tx *gorm.DB) error {
            txRepo := s.repo.WithTx(tx)

            pool := worker.NewPool[string, BulkShortenResult](
                numWorkers,
                len(urls),
                func(ctx context.Context,
                    job worker.Job[string]) (BulkShortenResult, error) {

                    url := &model.URL{
                        UserID:      userID,
                        OriginalURL: job.Payload,
                        ShortCode:   util.GenerateShortCode(),
                    }

                    if err := txRepo.Create(ctx, url); err != nil {
                        return BulkShortenResult{
                            OriginalURL: job.Payload,
                            Error:       "could not create short url",
                        }, err
                    }

                    return BulkShortenResult{
                        OriginalURL: job.Payload,
                        ShortCode:   url.ShortCode,
                    }, nil
                },
            )

            pool.Start(ctx)

            for i, url := range urls {
                pool.Submit(worker.Job[string]{
                    ID:      fmt.Sprintf("job-%d", i),
                    Payload: url,
                })
            }

            go pool.Close()

            for result := range pool.Results() {
                results = append(results, result.Value)
                if result.Err != nil {
                    return fmt.Errorf(
                        "bulk shorten failed at %s: %w",
                        result.Value.OriginalURL, result.Err)
                }
            }

            return nil
        })

    if err != nil {
        slog.Error("BulkShorten transaction failed", "error", err)
        // Return results with error marker
        return []BulkShortenResult{{
            Error: "bulk operation failed — all URLs rolled back",
        }}
    }

    return results
}

// ImportLinksCSV reads URLs from a CSV reader and creates short links.
// r can be an uploaded file, HTTP body, or any io.Reader.
func (s *URLService) ImportLinksCSV(
    ctx context.Context,
    r io.Reader,
    userID string) ([]ImportResult, error) {

    cr := csv.NewReader(r)

    // Skip header row
    if _, err := cr.Read(); err != nil {
        return nil, fmt.Errorf("read CSV header: %w", err)
    }

    var results []ImportResult
    row := 1

    for {
        record, err := cr.Read()
        if err == io.EOF {
            break // end of file — normal exit
        }
        if err != nil {
            return nil, fmt.Errorf(
                "read CSV row %d: %w", row, err)
        }

        row++
        if len(record) == 0 || record[0] == "" {
            continue // skip empty rows
        }

        originalURL := record[0]
        url := &model.URL{
            UserID:      userID,
            OriginalURL: originalURL,
            ShortCode:   util.GenerateShortCode(),
        }

        if err := s.repo.Create(ctx, url); err != nil {
            results = append(results, ImportResult{
                Row:         row,
                OriginalURL: originalURL,
                Error:       "could not create short url",
            })
            continue // skip failed rows, process the rest
        }

        results = append(results, ImportResult{
            Row:         row,
            OriginalURL: originalURL,
            ShortCode:   url.ShortCode,
        })
    }

    return results, nil
}