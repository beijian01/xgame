package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

// encodeLevel 自定义级别编码器，添加 ANSI 色码
func encodeLevel(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	color := getColorByLevel(level)
	enc.AppendString(fmt.Sprintf("%s%s%s", color, level.CapitalString(), resetColor))
}

// ANSI 色码
const (
	red        = "\x1b[31m"
	yellow     = "\x1b[33m"
	blue       = "\x1b[36m"
	green      = "\x1b[32m"
	resetColor = "\x1b[0m"
)

// getColorByLevel 根据日志级别返回对应的 ANSI 色码
func getColorByLevel(level zapcore.Level) string {
	switch level {
	case zapcore.DebugLevel:
		return blue
	case zapcore.InfoLevel:
		return green
	case zapcore.WarnLevel:
		return yellow
	case zapcore.ErrorLevel, zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return red
	default:
		return resetColor
	}
}

func newConsoleCore() zapcore.Core {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    encodeLevel,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	consoleCore := zapcore.NewCore(
		consoleEncoder,
		zapcore.AddSync(os.Stdout),
		zap.NewAtomicLevelAt(zap.InfoLevel),
	)
	return consoleCore
}
