package service

import (
    "context"
    "fmt"
    "log/slog"

    "golang.org/x/sync/errgroup"
    "urlshortener/internal/apperror"
    "urlshortener/internal/model"
)

type AdminURLRepository interface {
    FindAll(ctx context.Context) ([]*model.URL, error)
    Delete(ctx context.Context, id string) error
    DeleteByUserID(ctx context.Context, userID string) error
}

type AdminUserRepository interface {
    FindAll(ctx context.Context) ([]*model.User, error)
    Delete(ctx context.Context, id string) error
}

type AdminService struct {
    urlRepo  AdminURLRepository
    userRepo AdminUserRepository
}

type DashboardData struct {
    Links []*model.URL
    Users []*model.User
}

func NewAdminService(
    urlRepo AdminURLRepository,
    userRepo AdminUserRepository) *AdminService {
    return &AdminService{urlRepo: urlRepo, userRepo: userRepo}
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

    // errgroup.WithContext — if either goroutine fails,
    // the context is cancelled so the other can stop early
    g, gCtx := errgroup.WithContext(ctx)

    // Delete all user's links
    g.Go(func() error {
        if err := s.urlRepo.DeleteByUserID(gCtx, userID); err != nil {
            return fmt.Errorf("delete links for user %s: %w",
                userID, err)
        }
        return nil
    })

    // Delete the user account
    g.Go(func() error {
        if err := s.userRepo.Delete(gCtx, userID); err != nil {
            return fmt.Errorf("delete user %s: %w", userID, err)
        }
        return nil
    })

    // Wait — returns first non-nil error or nil if all succeeded
    if err := g.Wait(); err != nil {
        slog.Error("DeleteUser failed", "error", err,
            "userID", userID)
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