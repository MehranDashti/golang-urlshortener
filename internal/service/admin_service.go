package service

import (
    "sync"
    "context"

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

// ← this was missing
type AdminService struct {
    urlRepo  AdminURLRepository
    userRepo AdminUserRepository
}

// ← this was missing
func NewAdminService(urlRepo AdminURLRepository, userRepo AdminUserRepository) *AdminService {
    return &AdminService{urlRepo: urlRepo, userRepo: userRepo}
}

func (s *AdminService) GetAllLinks(
    ctx context.Context) ([]*model.URL, *apperror.AppError) {
    urls, err := s.urlRepo.FindAll(ctx)
    if err != nil {
        return nil, apperror.InternalWithErr("could not fetch links", err)
    }
    return urls, nil
}

func (s *AdminService) DeleteLink(
    ctx context.Context, id string) *apperror.AppError {
    if err := s.urlRepo.Delete(ctx, id); err != nil {
        return apperror.InternalWithErr("could not delete link", err)
    }
    return nil
}

func (s *AdminService) GetAllUsers(
    ctx context.Context) ([]*model.User, *apperror.AppError) {
    users, err := s.userRepo.FindAll(ctx)
    if err != nil {
        return nil, apperror.InternalWithErr("could not fetch users", err)
    }
    return users, nil
}


func (s *AdminService) DeleteUser(
    ctx context.Context, userID string) *apperror.AppError {

    // Run both deletes concurrently with WaitGroup
    var wg sync.WaitGroup
    // errCh is a channel — Go's way to communicate between goroutines
    // buffer of 2 means it can hold 2 errors without blocking
    errCh := make(chan error, 2)

    // Delete all user's links
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := s.urlRepo.DeleteByUserID(ctx, userID); err != nil {
            errCh <- err // send error to channel
        }
    }()

    // Delete the user account
    wg.Add(1)
    go func() {
        defer wg.Done()
        if err := s.userRepo.Delete(ctx, userID); err != nil {
            errCh <- err
        }
    }()

    // Wait for both goroutines to finish
    wg.Wait()
    close(errCh) // close channel so we can range over it

    // Check if any errors occurred
    for err := range errCh {
        if err != nil {
            return apperror.Internal("could not delete user")
        }
    }

    return nil
}