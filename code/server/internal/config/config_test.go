package config_test

import (
	"os"
	"testing"

	"github.com/family/crab-server/internal/config"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear environment variables
	os.Unsetenv("SERVER_PORT")
	os.Unsetenv("SERVER_HOST")
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("ENVIRONMENT")

	cfg := config.Load()

	if cfg.ServerPort != "8080" {
		t.Errorf("ServerPort = %s, want 8080", cfg.ServerPort)
	}
	if cfg.ServerHost != "0.0.0.0" {
		t.Errorf("ServerHost = %s, want 0.0.0.0", cfg.ServerHost)
	}
	// DATABASE_URL should be empty by default (no hardcoded credentials)
	if cfg.DatabaseURL != "" {
		t.Errorf("DatabaseURL = %s, want empty string (no default)", cfg.DatabaseURL)
	}
	if cfg.Environment != "development" {
		t.Errorf("Environment = %s, want development", cfg.Environment)
	}
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	os.Setenv("SERVER_PORT", "3000")
	os.Setenv("SERVER_HOST", "127.0.0.1")
	os.Setenv("ENVIRONMENT", "production")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("ENVIRONMENT")
	}()

	cfg := config.Load()

	if cfg.ServerPort != "3000" {
		t.Errorf("ServerPort = %s, want 3000", cfg.ServerPort)
	}
	if cfg.ServerHost != "127.0.0.1" {
		t.Errorf("ServerHost = %s, want 127.0.0.1", cfg.ServerHost)
	}
	if cfg.Environment != "production" {
		t.Errorf("Environment = %s, want production", cfg.Environment)
	}
}

func TestConfig_IsDevelopment(t *testing.T) {
	cfg := &config.Config{Environment: "development"}
	if !cfg.IsDevelopment() {
		t.Error("IsDevelopment() should return true")
	}
	if cfg.IsProduction() {
		t.Error("IsProduction() should return false")
	}
}

func TestConfig_IsProduction(t *testing.T) {
	cfg := &config.Config{Environment: "production"}
	if !cfg.IsProduction() {
		t.Error("IsProduction() should return true")
	}
	if cfg.IsDevelopment() {
		t.Error("IsDevelopment() should return false")
	}
}

func TestConfig_OtherEnvironment(t *testing.T) {
	cfg := &config.Config{Environment: "staging"}
	if cfg.IsDevelopment() {
		t.Error("IsDevelopment() should return false")
	}
	if cfg.IsProduction() {
		t.Error("IsProduction() should return false")
	}
}

func TestConfig_Validate_MissingDatabaseURL(t *testing.T) {
	cfg := &config.Config{DatabaseURL: ""}
	err := cfg.Validate()
	if err == nil {
		t.Error("Validate() should return error when DATABASE_URL is empty")
	}
	if err.Error() != "DATABASE_URL is required" {
		t.Errorf("Validate() error = %v, want 'DATABASE_URL is required'", err)
	}
}

func TestConfig_Validate_ValidDatabaseURL(t *testing.T) {
	cfg := &config.Config{DatabaseURL: "postgres://user:pass@localhost:5432/db"}
	err := cfg.Validate()
	if err != nil {
		t.Errorf("Validate() should not return error when DATABASE_URL is set, got: %v", err)
	}
}
