package service

import (
    "gorm.io/gorm"
    "encoding/csv"
    "io"
    "time"
    "context"
    "fmt"
    "log/slog"

    "golang.org/x/sync/errgroup"
    "urlshortener/internal/apperror"
    "urlshortener/internal/model"
    "urlshortener/internal/repository"
)

type AdminURLRepository interface {
    FindAll(ctx context.Context) ([]*model.URL, error)
    FindAllPaginated(ctx context.Context,
        params model.PaginationParams) ([]*model.URL, int64, error) 
    Delete(ctx context.Context, id string) error
    DeleteByUserID(ctx context.Context, userID string) error
    WithTx(tx *gorm.DB) *repository.URLRepository
    DB() *gorm.DB                              
}

type AdminUserRepository interface {
    FindAll(ctx context.Context) ([]*model.User, error)
    FindAllPaginated(ctx context.Context,
        params model.PaginationParams) ([]*model.User, int64, error)
    Delete(ctx context.Context, id string) error
    WithTx(tx *gorm.DB) *repository.UserRepository
    DB() *gorm.DB                              
}

type Transactor interface {
    Transaction(ctx context.Context,
        fn func(tx *gorm.DB) error) error
}

type DBTransactor struct {
    db *gorm.DB
}

func NewDBTransactor(db *gorm.DB) *DBTransactor {
    return &DBTransactor{db: db}
}

func (t *DBTransactor) Transaction(
    ctx context.Context,
    fn func(tx *gorm.DB) error) error {
    return t.db.WithContext(ctx).Transaction(fn)
}

type AdminService struct {
    urlRepo  AdminURLRepository
    userRepo AdminUserRepository
    transactor Transactor
}

type DashboardData struct {
    Links []*model.URL
    Users []*model.User
}

func NewAdminService(
    urlRepo AdminURLRepository,
    userRepo AdminUserRepository,
    transactor Transactor) *AdminService {
    return &AdminService{
        urlRepo:    urlRepo,
        userRepo:   userRepo,
        transactor: transactor,
    }
}

func (s *AdminService) GetAllLinks(
    ctx context.Context) ([]*model.URL, *apperror.AppError) {
    urls, err := s.urlRepo.FindAll(ctx)
    if err != nil {
        slog.Error("GetAllLinks failed", "error", err)
        return nil, apperror.InternalWithErr(
            "could not fetch links", err)
    }
    return urls, nil
}

func (s *AdminService) DeleteLink(
    ctx context.Context, id string) *apperror.AppError {
    if err := s.urlRepo.Delete(ctx, id); err != nil {
        slog.Error("DeleteLink failed", "error", err, "id", id)
        return apperror.InternalWithErr(
            "could not delete link", err)
    }
    return nil
}

func (s *AdminService) GetAllUsers(
    ctx context.Context) ([]*model.User, *apperror.AppError) {
    users, err := s.userRepo.FindAll(ctx)
    if err != nil {
        slog.Error("GetAllUsers failed", "error", err)
        return nil, apperror.InternalWithErr(
            "could not fetch users", err)
    }
    return users, nil
}

func (s *AdminService) DeleteUser(
    ctx context.Context,
    userID string) *apperror.AppError {

    err := s.transactor.Transaction(ctx,
        func(tx *gorm.DB) error {
            urlRepo  := s.urlRepo.WithTx(tx)
            userRepo := s.userRepo.WithTx(tx)

            if err := urlRepo.DeleteByUserID(ctx, userID); err != nil {
                return fmt.Errorf("delete links: %w", err)
            }
            if err := userRepo.Delete(ctx, userID); err != nil {
                return fmt.Errorf("delete user: %w", err)
            }
            return nil
        })

    if err != nil {
        slog.Error("DeleteUser transaction failed",
            "error", err, "userID", userID)
        return apperror.InternalWithErr(
            "could not delete user", err)
    }
    return nil
}

// GetDashboard fetches links and users concurrently.
// If either fails the whole operation fails fast.
func (s *AdminService) GetDashboard(
    ctx context.Context) (*DashboardData, *apperror.AppError) {

    var data DashboardData
    g, gCtx := errgroup.WithContext(ctx)

    g.Go(func() error {
        links, err := s.urlRepo.FindAll(gCtx)
        if err != nil {
            return fmt.Errorf("fetch links: %w", err)
        }
        data.Links = links // safe — only this goroutine writes Links
        return nil
    })

    g.Go(func() error {
        users, err := s.userRepo.FindAll(gCtx)
        if err != nil {
            return fmt.Errorf("fetch users: %w", err)
        }
        data.Users = users // safe — only this goroutine writes Users
        return nil
    })

    if err := g.Wait(); err != nil {
        slog.Error("GetDashboard failed", "error", err)
        return nil, apperror.InternalWithErr(
            "could not fetch dashboard data", err)
    }

    return &data, nil
}

// WriteLinksCSV writes all links as CSV to any io.Writer.
// Works with HTTP response, file, buffer — anything.
func (s *AdminService) WriteLinksCSV(
    ctx context.Context,
    w io.Writer) error {

    links, err := s.urlRepo.FindAll(ctx)
    if err != nil {
        return fmt.Errorf("WriteLinksCSV: %w", err)
    }

    // csv.NewWriter wraps any io.Writer
    cw := csv.NewWriter(w)
    defer cw.Flush() // ensure buffered data is written

    // Header row
    if err := cw.Write([]string{
        "id", "short_code", "original_url",
        "clicks", "created_at", "expires_at",
    }); err != nil {
        return fmt.Errorf("write CSV header: %w", err)
    }

    // Data rows
    for _, link := range links {
        expiresAt := ""
        if link.ExpiresAt != nil {
            expiresAt = link.ExpiresAt.Format(time.RFC3339)
        }

        if err := cw.Write([]string{
            link.ID,
            link.ShortCode,
            link.OriginalURL,
            fmt.Sprintf("%d", link.Clicks),
            link.CreatedAt.Format(time.RFC3339),
            expiresAt,
        }); err != nil {
            return fmt.Errorf("write CSV row %s: %w",
                link.ID, err)
        }
    }

    // Check for any errors during writes
    cw.Flush()
    return cw.Error()
}

func (s *AdminService) GetAllLinksPaginated(
    ctx context.Context,
    params model.PaginationParams) (*model.PaginatedResult[*model.URL], *apperror.AppError) {

    urls, total, err := s.urlRepo.FindAllPaginated(ctx, params)
    if err != nil {
        slog.Error("GetAllLinksPaginated failed", "error", err)
        return nil, apperror.InternalWithErr(
            "could not fetch links", err)
    }

    result := model.NewPaginatedResult(urls, total, params)
    return &result, nil
}

func (s *AdminService) GetAllUsersPaginated(
    ctx context.Context,
    params model.PaginationParams) (*model.PaginatedResult[*model.User], *apperror.AppError) {

    users, total, err := s.userRepo.FindAllPaginated(ctx, params)
    if err != nil {
        slog.Error("GetAllUsersPaginated failed", "error", err)
        return nil, apperror.InternalWithErr(
            "could not fetch users", err)
    }

    result := model.NewPaginatedResult(users, total, params)
    return &result, nil
}

