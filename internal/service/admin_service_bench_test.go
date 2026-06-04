package service_test

import (
	"context"
	"io"
	"testing"

	"urlshortener/internal/model"
	"urlshortener/internal/service"
)

func BenchmarkWriteLinksCSV(b *testing.B) {
	urls := make([]*model.URL, 100)
	for i := range urls {
		urls[i] = &model.URL{
			ID:          "id",
			ShortCode:   "abc123",
			OriginalURL: "https://google.com",
			Clicks:      5,
		}
	}

	adminRepo := &mockAdminURLRepo{
		findAllFn: func(ctx context.Context) ([]*model.URL, error) {
			return urls, nil
		},
	}
	svc := service.NewAdminService(
		adminRepo, &mockAdminUserRepo{}, &mockTransactor{})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := svc.WriteLinksCSV(context.Background(), io.Discard); err != nil {
			b.Fatal(err)
		}
	}
}
