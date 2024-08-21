package log

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newFileCore(config *ZapConfig) zapcore.Core {
	level, err := zap.ParseAtomicLevel(config.Level)
	if err != nil {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	fileCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(&config.RotateLog),
		level,
	)
	return fileCore
}
