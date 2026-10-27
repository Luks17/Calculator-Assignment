package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/luks17/sezzle-assignment/apps/calc-service/internal/infrastructure/config"
)

func TestLoad(t *testing.T) {
	for _, tt := range []struct{ host, port, want string }{
		{"", "", "0.0.0.0:8080"},
		{"127.0.0.1", "9090", "127.0.0.1:9090"},
		{"::1", "8080", "[::1]:8080"},
		{"localhost", "1", "localhost:1"},
		{"localhost", "65535", "localhost:65535"},
	} {
		t.Run(tt.want, func(t *testing.T) {
			t.Setenv("SERVER_HOST", tt.host)
			t.Setenv("SERVER_PORT", tt.port)
			cfg, err := config.Load()
			require.NoError(t, err)
			require.Equal(t, tt.want, cfg.Address())
		})
	}
}

func TestInvalidPort(t *testing.T) {
	for _, port := range []string{"0", "-1", "65536", "not-a-port", "8080.5", "99999999999999999999"} {
		t.Run(port, func(t *testing.T) {
			t.Setenv("SERVER_PORT", port)
			_, err := config.Load()
			require.Error(t, err)
		})
	}
}

func TestCORSConfiguration(t *testing.T) {
	for _, tt := range []struct {
		name, environment, origins, wantEnvironment string
		wantOrigins                                 []string
	}{
		{"secure defaults", "", "", "production", nil},
		{"explicit local mode", "local", "", "local", nil},
		{"configured origins", "production", "http://localhost:5173,http://localhost:8081", "production", []string{"http://localhost:5173", "http://localhost:8081"}},
		{"trim whitespace and empty entries", "staging", " https://app.example , , https://docs.example, ", "staging", []string{"https://app.example", "https://docs.example"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SERVER_HOST", "")
			t.Setenv("SERVER_PORT", "")
			t.Setenv("APP_ENV", tt.environment)
			t.Setenv("ALLOWED_ORIGINS", tt.origins)
			cfg, err := config.Load()
			require.NoError(t, err)
			require.Equal(t, tt.wantEnvironment, cfg.GetAppEnv())
			require.Equal(t, tt.wantOrigins, cfg.GetAllowedOrigins())
		})
	}
}
