package service

import (
    "encoding/csv"
    "io"
    "fmt"
    "log/slog"
    "context"
    "time"
    "sync"
    
    "urlshortener/internal/apperror"
    "urlshortener/internal/worker"
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

// BulkShorten shortens multiple URLs concurrently using a worker pool.
// numWorkers controls how many DB writes happen simultaneously.
func (s *URLService) BulkShorten(
    ctx context.Context,
    urls []string,
    userID string,
    numWorkers int,
) []BulkShortenResult {

    // Create pool — T=string (URL), R=BulkShortenResult
    pool := worker.NewPool[string, BulkShortenResult](
        numWorkers,
        len(urls), // buffer = number of jobs
        func(ctx context.Context,
            job worker.Job[string]) (BulkShortenResult, error) {

            url := &model.URL{
                UserID:      userID,
                OriginalURL: job.Payload,
                ShortCode:   util.GenerateShortCode(),
            }

            if err := s.repo.Create(ctx, url); err != nil {
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

    // Start workers
    pool.Start(ctx)

    // Submit all jobs
    for i, url := range urls {
        pool.Submit(worker.Job[string]{
            ID:      fmt.Sprintf("job-%d", i),
            Payload: url,
        })
    }

    // Signal no more jobs — workers will exit after draining
    // Run in goroutine because Close() blocks until workers finish
    // but we also need to read results — deadlock if we block here
    go pool.Close()

    // Collect results
    results := make([]BulkShortenResult, 0, len(urls))
    for result := range pool.Results() {
        results = append(results, result.Value)
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