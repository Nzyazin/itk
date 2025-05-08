package logger

import (
	"os"
	
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)
type Logger interface {
    Info(msg string, fields ...zap.Field)
    Error(msg string, fields ...zap.Field)
    Debug(msg string, fields ...zap.Field)
    Warn(msg string, fields ...zap.Field)
    With(fields ...zap.Field) Logger
}

type ZapAdapter struct {
    zapLogger *zap.Logger
}

func (za *ZapAdapter) With(fields ...zap.Field) Logger {
    return &ZapAdapter{zapLogger: za.zapLogger.With(fields...)}
}

func (za *ZapAdapter) Info(msg string, fields ...zap.Field) {
    za.zapLogger.Info(msg, fields...)
}

func (za *ZapAdapter) Error(msg string, fields ...zap.Field) {
    za.zapLogger.Error(msg, fields...)
}

func (za *ZapAdapter) Debug(msg string, fields ...zap.Field) {
    za.zapLogger.Debug(msg, fields...)
}

func (za *ZapAdapter) Warn(msg string, fields ...zap.Field) {
    za.zapLogger.Warn(msg, fields...)
}

func NewLogger() (Logger, func()) {
	// Ensure log directory exists
	_ = os.MkdirAll("logs", os.ModePerm)

	infoFile, err := os.OpenFile("logs/info.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("failed to open info log file: " + err.Error())
	}

	errorFile, err := os.OpenFile("logs/error.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("failed to open error log file: " + err.Error())
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	infoCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(infoFile),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl <= zapcore.InfoLevel
		}),
	)

	errorCore := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(errorFile),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.WarnLevel
		}),
	)

	core := zapcore.NewTee(infoCore, errorCore)
	logger := zap.New(core, zap.AddCaller())

	cleanup := func() {
		logger.Sync()
		infoFile.Close()
		errorFile.Close()
	}

	return &ZapAdapter{zapLogger: logger}, cleanup
}