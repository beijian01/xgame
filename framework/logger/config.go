package log

import "gopkg.in/natefinch/lumberjack.v2"

type ZapConfig struct {
	Level               string            `json:"level"`
	EnableConsoleWriter bool              `json:"enableConsoleWriter"`
	EnableFileWriter    bool              `json:"enableFileWriter"`
	RotateLog           lumberjack.Logger `json:"rotateLog"`
}
