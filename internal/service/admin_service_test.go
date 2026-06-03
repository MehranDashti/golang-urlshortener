package service_test

import (
    "context"
    "errors"
    "testing"

    "github.com/stretchr/testify/assert"
    "urlshortener/internal/apperror"
    "urlshortener/internal/model"
    "urlshortener/internal/service"
)

// mockAdminURLRepo — functional mock
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
    if m.deleteFn != nil { return m.deleteFn(ctx, id) }
    return nil
}
func (m *mockAdminURLRepo) DeleteByUserID(ctx context.Context, userID string) error {
    if m.deleteByUserIDFn != nil { return m.deleteByUserIDFn(ctx, userID) }
    return nil
}

// mockAdminUserRepo — functional mock
type mockAdminUserRepo struct {
    findAllFn func(ctx context.Context) ([]*model.User, error)
    deleteFn  func(ctx context.Context, id string) error
}

func (m *mockAdminUserRepo) FindAll(ctx context.Context) ([]*model.User, error) {
    if m.findAllFn != nil { return m.findAllFn(ctx) }
    return []*model.User{}, nil
}
func (m *mockAdminUserRepo) Delete(ctx context.Context, id string) error {
    if m.deleteFn != nil { return m.deleteFn(ctx, id) }
    return nil
}

func TestDeleteUser(t *testing.T) {
    tests := []struct {
        name          string
        urlRepoErr    error
        userRepoErr   error
        wantAppErrNil bool
    }{
        {
            name:          "success — both deletes succeed",
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
        {
            name:          "both fail — first error returned",
            urlRepoErr:    errors.New("links db error"),
            userRepoErr:   errors.New("users db error"),
            wantAppErrNil: false,
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            urlRepo := &mockAdminURLRepo{
                deleteByUserIDFn: func(ctx context.Context,
                    userID string) error {
                    return tt.urlRepoErr
                },
            }
            userRepo := &mockAdminUserRepo{
                deleteFn: func(ctx context.Context,
                    id string) error {
                    return tt.userRepoErr
                },
            }

            svc := service.NewAdminService(urlRepo, userRepo)
            appErr := svc.DeleteUser(context.Background(), "user-123")

            if tt.wantAppErrNil {
                assert.Nil(t, appErr)
            } else {
                assert.NotNil(t, appErr)
                assert.Equal(t, 500, appErr.Code)

                // errors.As walks the chain — finds AppError
                var ae *apperror.AppError
                assert.True(t, errors.As(appErr, &ae))
            }
        })
    }
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

    svc := service.NewAdminService(urlRepo, userRepo)
    data, appErr := svc.GetDashboard(context.Background())

    assert.Nil(t, appErr)
    assert.Len(t, data.Links, 2)
    assert.Len(t, data.Users, 1)
}