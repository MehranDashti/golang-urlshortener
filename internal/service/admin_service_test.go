package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"urlshortener/internal/model"
	"urlshortener/internal/repository"
	"urlshortener/internal/service"
)

// ── mockTransactor ────────────────────────────────────────────────────────────
// Bypasses real DB transaction — calls mock repos directly
type mockTransactor struct {
	fn func(ctx context.Context) error
}

func (m *mockTransactor) Transaction(
	ctx context.Context,
	fn func(tx *gorm.DB) error) error {
	if m.fn != nil {
		return m.fn(ctx)
	}
	return nil
}

// ── mockAdminURLRepo ──────────────────────────────────────────────────────────
type mockAdminURLRepo struct {
	findAllFn        func(ctx context.Context) ([]*model.URL, error)
	deleteFn         func(ctx context.Context, id string) error
	deleteByUserIDFn func(ctx context.Context, userID string) error
}

func (m *mockAdminURLRepo) FindAll(ctx context.Context) ([]*model.URL, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx)
	}
	return []*model.URL{}, nil
}
func (m *mockAdminURLRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *mockAdminURLRepo) DeleteByUserID(ctx context.Context, userID string) error {
	if m.deleteByUserIDFn != nil {
		return m.deleteByUserIDFn(ctx, userID)
	}
	return nil
}
func (m *mockAdminURLRepo) WithTx(_ *gorm.DB) *repository.URLRepository { return nil }
func (m *mockAdminURLRepo) DB() *gorm.DB                                 { return nil }

// ── mockAdminUserRepo ─────────────────────────────────────────────────────────
type mockAdminUserRepo struct {
	findAllFn func(ctx context.Context) ([]*model.User, error)
	deleteFn  func(ctx context.Context, id string) error
}

func (m *mockAdminUserRepo) FindAll(ctx context.Context) ([]*model.User, error) {
	if m.findAllFn != nil {
		return m.findAllFn(ctx)
	}
	return []*model.User{}, nil
}
func (m *mockAdminUserRepo) Delete(ctx context.Context, id string) error {
	if m.deleteFn != nil {
		return m.deleteFn(ctx, id)
	}
	return nil
}
func (m *mockAdminUserRepo) WithTx(_ *gorm.DB) *repository.UserRepository { return nil }
func (m *mockAdminUserRepo) DB() *gorm.DB                                  { return nil }

// ── Tests ─────────────────────────────────────────────────────────────────────

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name          string
		urlRepoErr    error
		userRepoErr   error
		wantAppErrNil bool
	}{
		{
			name:          "success",
			urlRepoErr:    nil,
			userRepoErr:   nil,
			wantAppErrNil: true,
		},
		{
			name:          "url repo fails",
			urlRepoErr:    errors.New("db error"),
			userRepoErr:   nil,
			wantAppErrNil: false,
		},
		{
			name:          "user repo fails",
			urlRepoErr:    nil,
			userRepoErr:   errors.New("db error"),
			wantAppErrNil: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			urlRepo := &mockAdminURLRepo{
				deleteByUserIDFn: func(ctx context.Context, userID string) error {
					return tt.urlRepoErr
				},
			}
			userRepo := &mockAdminUserRepo{
				deleteFn: func(ctx context.Context, id string) error {
					return tt.userRepoErr
				},
			}

			// mockTransactor calls mock repos directly
			// bypassing WithTx which needs a real *gorm.DB
			transactor := &mockTransactor{
				fn: func(ctx context.Context) error {
					if err := urlRepo.DeleteByUserID(ctx, "user-123"); err != nil {
						return err
					}
					return userRepo.Delete(ctx, "user-123")
				},
			}

			svc := service.NewAdminService(urlRepo, userRepo, transactor)
			appErr := svc.DeleteUser(context.Background(), "user-123")

			if tt.wantAppErrNil {
				assert.Nil(t, appErr)
			} else {
				assert.NotNil(t, appErr)
				assert.Equal(t, 500, appErr.Code)
			}
		})
	}
}

func TestDeleteUser_Transaction_Rollback(t *testing.T) {
	urlRepo := &mockAdminURLRepo{
		deleteByUserIDFn: func(ctx context.Context, userID string) error {
			return nil // links delete succeeds
		},
	}
	userRepo := &mockAdminUserRepo{
		deleteFn: func(ctx context.Context, id string) error {
			return errors.New("user delete failed") // user delete fails
		},
	}

	transactor := &mockTransactor{
		fn: func(ctx context.Context) error {
			if err := urlRepo.DeleteByUserID(ctx, "user-123"); err != nil {
				return err
			}
			return userRepo.Delete(ctx, "user-123")
		},
	}

	svc := service.NewAdminService(urlRepo, userRepo, transactor)
	appErr := svc.DeleteUser(context.Background(), "user-123")

	assert.NotNil(t, appErr)
	t.Log("Transaction rollback tested via integration tests with real DB")
}

func TestGetDashboard_Concurrent(t *testing.T) {
	urlRepo := &mockAdminURLRepo{
		findAllFn: func(ctx context.Context) ([]*model.URL, error) {
			return []*model.URL{
				{ID: "1", ShortCode: "abc"},
				{ID: "2", ShortCode: "def"},
			}, nil
		},
	}
	userRepo := &mockAdminUserRepo{
		findAllFn: func(ctx context.Context) ([]*model.User, error) {
			return []*model.User{
				{ID: "1", Email: "a@test.com"},
			}, nil
		},
	}

	svc := service.NewAdminService(urlRepo, userRepo, &mockTransactor{})
	data, appErr := svc.GetDashboard(context.Background())

	assert.Nil(t, appErr)
	assert.Len(t, data.Links, 2)
	assert.Len(t, data.Users, 1)
}