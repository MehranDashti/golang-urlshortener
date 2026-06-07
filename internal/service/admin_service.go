package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"gorm.io/gorm"
	"io"
	"log/slog"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"urlshortener/internal/apperror"
	"urlshortener/internal/model"
	"urlshortener/internal/repository"
)

var csvBufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

type AdminURLRepository interface {
	FindAll(ctx context.Context) ([]*model.URL, error)
	FindAllPaginated(ctx context.Context,
		params model.PaginationParams) ([]*model.URL, int64, error)
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
	WithTx(tx *gorm.DB) *repository.URLRepository
	DB() *gorm.DB
	TopLinks(ctx context.Context, limit int) ([]*model.URL, error)
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

type DashboardData struct {
	Links    []*model.URL
	Users    []*model.User
	TopLinks []*model.URL // ← add this
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
	urlRepo    AdminURLRepository
	userRepo   AdminUserRepository
	transactor Transactor
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
			urlRepo := s.urlRepo.WithTx(tx)
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

func (s *AdminService) GetDashboard(
	ctx context.Context) (*DashboardData, *apperror.AppError) {

	var data DashboardData
	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		links, err := s.urlRepo.FindAll(gCtx)
		if err != nil {
			return fmt.Errorf("fetch links: %w", err)
		}
		data.Links = links
		return nil
	})

	g.Go(func() error {
		users, err := s.userRepo.FindAll(gCtx)
		if err != nil {
			return fmt.Errorf("fetch users: %w", err)
		}
		data.Users = users
		return nil
	})

	g.Go(func() error {
		top, err := s.urlRepo.TopLinks(gCtx, 5)
		if err != nil {
			return fmt.Errorf("fetch top links: %w", err)
		}
		data.TopLinks = top
		return nil
	})

	if err := g.Wait(); err != nil {
		slog.Error("GetDashboard failed", "error", err)
		return nil, apperror.InternalWithErr(
			"could not fetch dashboard data", err)
	}

	return &data, nil
}

func (s *AdminService) WriteLinksCSV(
	ctx context.Context,
	w io.Writer) error {

	links, err := s.urlRepo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("WriteLinksCSV: %w", err)
	}

	// Get buffer from pool — reused across requests
	buf := csvBufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer csvBufPool.Put(buf) // return to pool when done

	cw := csv.NewWriter(buf) // write to buffer first

	if err := cw.Write([]string{
		"id", "short_code", "original_url",
		"clicks", "created_at", "expires_at",
	}); err != nil {
		return fmt.Errorf("write CSV header: %w", err)
	}

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

	cw.Flush()
	if err := cw.Error(); err != nil {
		return err
	}

	// Write buffered content to the actual writer
	_, err = buf.WriteTo(w)
	return err
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
