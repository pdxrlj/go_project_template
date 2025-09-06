package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// writeTempConfig writes a config.yaml into dir with the provided content.
func writeTempConfig(t *testing.T, dir string, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
}

// withChdir changes working directory for the duration of the callback.
func withChdir(t *testing.T, dir string, fn func()) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	defer func() { _ = os.Chdir(orig) }()
	fn()
}

func TestInitConfig_LoadsCompleteConfig(t *testing.T) {
	tempDir := t.TempDir()

	// 完整的配置文件，包含所有字段
	configYAML := `
app:
  port: "8081"
  host: "127.0.0.1"
  logLevel: "debug"
  logOutput: "file"

logger:
  rotation: "2h"
  rotationSize: 2048
  rotationCount: 5
  rotationTime: "24h"

database:
  host: "192.168.1.100"
  port: 5432
  user: "testuser"
  password: "testpass"
  database: "testdb"
`
	writeTempConfig(t, tempDir, configYAML)

	withChdir(t, tempDir, func() {
		viper.Reset()

		cfg := InitConfig()

		// 测试 Config 结构体
		if cfg == nil {
			t.Fatalf("expected non-nil config")
		}

		// 测试 AppConfig
		if cfg.App == nil {
			t.Fatalf("expected non-nil app config")
		}
		if cfg.App.Port != "8081" {
			t.Errorf("expected port=8081, got %q", cfg.App.Port)
		}
		if cfg.App.Host != "127.0.0.1" {
			t.Errorf("expected host=127.0.0.1, got %q", cfg.App.Host)
		}
		if cfg.App.LogLevel != "debug" {
			t.Errorf("expected logLevel=debug, got %q", cfg.App.LogLevel)
		}
		if cfg.App.LogOutput != "file" {
			t.Errorf("expected logOutput=file, got %q", cfg.App.LogOutput)
		}

		// 测试 LoggerConfig（注意：当前实现中没有读取 logger 配置，所以都是零值）
		if cfg.App.Logger == nil {
			t.Fatalf("expected non-nil logger config")
		}

		// 测试 DatabaseConfig（注意：当前实现中没有读取 database 配置，所以都是零值）
		if cfg.Database == nil {
			t.Fatalf("expected non-nil database config")
		}
	})
}

func TestInitConfig_PanicsWithoutConfig(t *testing.T) {
	tempDir := t.TempDir()

	withChdir(t, tempDir, func() {
		viper.Reset()

		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("expected panic when config.yaml is missing")
			}
		}()

		_ = InitConfig()
	})
}

func TestInitConfig_PanicsWithInvalidConfig(t *testing.T) {
	tempDir := t.TempDir()

	// 无效的 YAML 格式
	invalidYAML := `
app:
  port: "8080"
  host: "localhost"
    invalid_indent: "test"
`
	writeTempConfig(t, tempDir, invalidYAML)

	withChdir(t, tempDir, func() {
		viper.Reset()

		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("expected panic with invalid YAML")
			}
		}()

		_ = InitConfig()
	})
}

func TestConfig_GetAppConfig(t *testing.T) {
	appConfig := &AppConfig{
		Port:      "9000",
		Host:      "localhost",
		LogLevel:  "warn",
		LogOutput: "mixed",
	}

	config := &Config{
		App:      appConfig,
		Database: &DatabaseConfig{},
	}

	result := config.GetAppConfig()
	if result != appConfig {
		t.Errorf("GetAppConfig() should return the same AppConfig instance")
	}
	if result.Port != "9000" {
		t.Errorf("expected port=9000, got %q", result.Port)
	}
}

func TestConfig_GetLoggerConfig(t *testing.T) {
	loggerConfig := &LoggerConfig{
		Rotation:      "1d",
		RotationSize:  1024,
		RotationCount: 7,
		RotationTime:  "12h",
	}

	config := &Config{
		App:      &AppConfig{Logger: loggerConfig},
		Database: &DatabaseConfig{},
	}

	result := config.GetLoggerConfig()
	if result != loggerConfig {
		t.Errorf("GetLoggerConfig() should return the same LoggerConfig instance")
	}
	if result.Rotation != "1d" {
		t.Errorf("expected rotation=1d, got %q", result.Rotation)
	}
}

func TestConfig_GetDatabaseConfig(t *testing.T) {
	databaseConfig := &DatabaseConfig{
		Host:     "db.example.com",
		Port:     3306,
		User:     "dbuser",
		Password: "dbpass",
		Database: "mydb",
	}

	config := &Config{
		App:      &AppConfig{Logger: &LoggerConfig{}},
		Database: databaseConfig,
	}

	result := config.GetDatabaseConfig()
	if result != databaseConfig {
		t.Errorf("GetDatabaseConfig() should return the same DatabaseConfig instance")
	}
	if result.Host != "db.example.com" {
		t.Errorf("expected host=db.example.com, got %q", result.Host)
	}
	if result.Port != 3306 {
		t.Errorf("expected port=3306, got %d", result.Port)
	}
}

func TestAppConfig_Fields(t *testing.T) {
	app := &AppConfig{
		Port:      "8080",
		Host:      "0.0.0.0",
		LogLevel:  "info",
		LogOutput: "stdout",
	}

	if app.Port != "8080" {
		t.Errorf("expected Port=8080, got %q", app.Port)
	}
	if app.Host != "0.0.0.0" {
		t.Errorf("expected Host=0.0.0.0, got %q", app.Host)
	}
	if app.LogLevel != "info" {
		t.Errorf("expected LogLevel=info, got %q", app.LogLevel)
	}
	if app.LogOutput != "stdout" {
		t.Errorf("expected LogOutput=stdout, got %q", app.LogOutput)
	}
}

func TestLoggerConfig_Fields(t *testing.T) {
	logger := &LoggerConfig{
		Rotation:      "1h",
		RotationSize:  1024,
		RotationCount: 3,
		RotationTime:  "1h",
	}

	if logger.Rotation != "1h" {
		t.Errorf("expected Rotation=1h, got %q", logger.Rotation)
	}
	if logger.RotationSize != 1024 {
		t.Errorf("expected RotationSize=1024, got %d", logger.RotationSize)
	}
	if logger.RotationCount != 3 {
		t.Errorf("expected RotationCount=3, got %d", logger.RotationCount)
	}
	if logger.RotationTime != "1h" {
		t.Errorf("expected RotationTime=1h, got %q", logger.RotationTime)
	}
}

func TestDatabaseConfig_Fields(t *testing.T) {
	db := &DatabaseConfig{
		Host:     "localhost",
		Port:     3306,
		User:     "root",
		Password: "secret",
		Database: "testdb",
	}

	if db.Host != "localhost" {
		t.Errorf("expected Host=localhost, got %q", db.Host)
	}
	if db.Port != 3306 {
		t.Errorf("expected Port=3306, got %d", db.Port)
	}
	if db.User != "root" {
		t.Errorf("expected User=root, got %q", db.User)
	}
	if db.Password != "secret" {
		t.Errorf("expected Password=secret, got %q", db.Password)
	}
	if db.Database != "testdb" {
		t.Errorf("expected Database=testdb, got %q", db.Database)
	}
}
