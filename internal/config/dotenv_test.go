package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDotEnvLoadsFromParentDirectory(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "cmd", "api")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested: %v", err)
	}

	envContent := "DB_DSN=postgres://user:pass@localhost:5432/app\nREDIS_URL=localhost:6379\n"
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(envContent), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	prevDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.Chdir(nested); err != nil {
		t.Fatalf("chdir nested: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevDir)
	})

	restoreEnv(t, "DB_DSN")
	restoreEnv(t, "REDIS_URL")

	if err := os.Unsetenv("DB_DSN"); err != nil {
		t.Fatalf("unset DB_DSN: %v", err)
	}

	if err := os.Unsetenv("REDIS_URL"); err != nil {
		t.Fatalf("unset REDIS_URL: %v", err)
	}

	if err := loadDotEnv(); err != nil {
		t.Fatalf("loadDotEnv: %v", err)
	}

	if got := os.Getenv("DB_DSN"); got != "postgres://user:pass@localhost:5432/app" {
		t.Fatalf("DB_DSN = %q", got)
	}

	if got := os.Getenv("REDIS_URL"); got != "localhost:6379" {
		t.Fatalf("REDIS_URL = %q", got)
	}
}

func TestLoadDotEnvDoesNotOverrideExistingEnvironment(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("DB_DSN=from-file\n"), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	prevDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir root: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevDir)
	})

	t.Setenv("DB_DSN", "from-env")

	if err := loadDotEnv(); err != nil {
		t.Fatalf("loadDotEnv: %v", err)
	}

	if got := os.Getenv("DB_DSN"); got != "from-env" {
		t.Fatalf("DB_DSN = %q", got)
	}
}

func TestLoadUsesDefaultAppPort(t *testing.T) {
	root := t.TempDir()

	prevDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}

	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir root: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(prevDir)
	})

	restoreEnv(t, "APP_PORT")
	if err := os.Unsetenv("APP_PORT"); err != nil {
		t.Fatalf("unset APP_PORT: %v", err)
	}

	cfg := Load()
	if cfg.APP_PORT != "3005" {
		t.Fatalf("APP_PORT = %q", cfg.APP_PORT)
	}

	if cfg.AppAddr() != ":3005" {
		t.Fatalf("AppAddr = %q", cfg.AppAddr())
	}
}

func TestAppAddr(t *testing.T) {
	tests := []struct {
		name string
		port string
		want string
	}{
		{name: "plain port", port: "8080", want: ":8080"},
		{name: "prefixed port", port: ":8081", want: ":8081"},
		{name: "host and port", port: "127.0.0.1:8082", want: "127.0.0.1:8082"},
		{name: "empty port", port: "", want: ":3005"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{APP_PORT: tt.port}
			if got := cfg.AppAddr(); got != tt.want {
				t.Fatalf("AppAddr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseEnvLine(t *testing.T) {
	key, value, ok := parseEnvLine(`export DB_DSN="postgres://user:pass@localhost:5432/app"`)
	if !ok {
		t.Fatal("parseEnvLine returned ok=false")
	}

	if key != "DB_DSN" {
		t.Fatalf("key = %q", key)
	}

	if value != "postgres://user:pass@localhost:5432/app" {
		t.Fatalf("value = %q", value)
	}
}

func restoreEnv(t *testing.T, key string) {
	t.Helper()

	value, exists := os.LookupEnv(key)
	t.Cleanup(func() {
		if !exists {
			_ = os.Unsetenv(key)
			return
		}

		_ = os.Setenv(key, value)
	})
}
