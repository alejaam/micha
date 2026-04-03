package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseAllowedOrigins(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		raw        string
		production bool
		want       []string
		wantErr    string
	}{
		{
			name:       "development defaults to wildcard when empty",
			raw:        "",
			production: false,
			want:       []string{"*"},
		},
		{
			name:       "production requires explicit origins",
			raw:        "",
			production: true,
			wantErr:    "ALLOWED_ORIGINS environment variable is required in production",
		},
		{
			name:       "production rejects wildcard",
			raw:        "https://app.example.com,*",
			production: true,
			wantErr:    "ALLOWED_ORIGINS cannot contain wildcard '*' in production",
		},
		{
			name:       "trims and skips empty origins",
			raw:        " https://app.example.com , , https://admin.example.com ",
			production: true,
			want:       []string{"https://app.example.com", "https://admin.example.com"},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseAllowedOrigins(tc.raw, tc.production)
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestIsProductionEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		appEnv string
		env    string
		want   bool
	}{
		{name: "APP_ENV production", appEnv: "production", env: "", want: true},
		{name: "APP_ENV prod", appEnv: "prod", env: "", want: true},
		{name: "fallback to ENV", appEnv: "", env: "production", want: true},
		{name: "non-production value", appEnv: "development", env: "", want: false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tc.want, isProductionEnv(tc.appEnv, tc.env))
		})
	}
}
