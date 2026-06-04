package util_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"urlshortener/internal/util"
)

func TestNormaliseURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "lowercase scheme",
			input: "HTTPS://google.com",
			want:  "https://google.com",
		},
		{
			name:  "lowercase host",
			input: "https://GOOGLE.COM",
			want:  "https://google.com",
		},
		{
			name:  "remove trailing slash",
			input: "https://google.com/path/",
			want:  "https://google.com/path",
		},
		{
			name:  "keep root slash",
			input: "https://google.com/",
			want:  "https://google.com/",
		},
		{
			name:  "sort query params",
			input: "https://google.com?z=3&a=1&m=2",
			want:  "https://google.com?a=1&m=2&z=3",
		},
		{
			name:  "remove fragment",
			input: "https://google.com/page#section",
			want:  "https://google.com/page",
		},
		{
			name:  "already normalised",
			input: "https://google.com/path",
			want:  "https://google.com/path",
		},
		{
			name:    "reject javascript scheme",
			input:   "javascript:alert(1)",
			wantErr: true,
		},
		{
			name:    "reject ftp scheme",
			input:   "ftp://files.example.com",
			wantErr: true,
		},
		{
			name:    "reject invalid URL",
			input:   "not a url",
			wantErr: true,
		},
		{
			name:  "full normalisation",
			input: "HTTPS://GOOGLE.COM/Path/?B=2&A=1#fragment",
			want:  "https://google.com/Path?A=1&B=2",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := util.NormaliseURL(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
