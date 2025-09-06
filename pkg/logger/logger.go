package logger

import (
	"io"
	"log/slog"
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Logger struct {
	Level         string
	Output        string
	Rotation      string
	RotationSize  int
	RotationCount int
	RotationTime  string
}

func NewLogger(level, output, rotation,
	rotationTime string,
	rotationSize int,
	rotationCount int) *Logger {
	return &Logger{
		Level:         level,
		Output:        output,
		Rotation:      rotation,
		RotationSize:  rotationSize,
		RotationCount: rotationCount,
		RotationTime:  rotationTime,
	}
}

func (l *Logger) Init() {
	dsetWriter := io.MultiWriter(os.Stdout)

	if l.Output == "file" {
		rotater := lumberjack.Logger{
			Filename:   l.Output,
			MaxSize:    l.RotationSize,
			MaxAge:     l.RotationCount,
			MaxBackups: l.RotationCount,
			Compress:   true,
		}
		dsetWriter = io.MultiWriter(dsetWriter, &rotater)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(dsetWriter,
		&slog.HandlerOptions{
			Level: l.GetLevel(),
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if t, ok := a.Value.Any().(time.Time); ok {
					a.Value = slog.StringValue(t.Format(time.DateTime))
				}

				// 去除包名称
				if a.Key == slog.SourceKey {
					// TODO: 去除包名称
				}

				return a
			},
		})))

}

func (l *Logger) GetLevel() slog.Level {
	level := slog.Level(0)
	level.UnmarshalText([]byte(l.Level))
	return level
}
