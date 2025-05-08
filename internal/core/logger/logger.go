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
}

func NewLogger() (*zap.Logger, func()) {
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

    return logger, cleanup
}