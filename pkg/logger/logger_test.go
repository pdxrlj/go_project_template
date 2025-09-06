package logger

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogger_Init(t *testing.T) {
	logger := NewLogger("info", "stdout", "1h", "1h", 1024, 3)
	logger.Init()
}

func TestLogger_GetLevel(t *testing.T) {
	logger := NewLogger("info", "stdout", "1h", "1h", 1024, 3)
	level := logger.GetLevel()
	t.Log("level", level)
	assert.Equal(t, slog.LevelInfo, level)
}
