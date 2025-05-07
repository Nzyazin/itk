package logger

import (
	"os"
	
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger создает новый экземпляр логгера и возвращает его вместе с функцией для очистки
func NewLogger() (*zap.Logger, func()) {
	// Открываем файл для логирования
	logFile, err := os.OpenFile("logs/service.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic("failed to open log file: " + err.Error())
	}

	// Настройка конфигурации логера
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout", "file://" + logFile.Name()} // Записываем в stdout и файл
	config.EncoderConfig = zapcore.EncoderConfig{
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

	// Создаем новый логгер с конфигурированием
	logger, err := config.Build()
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}

	// Функция для очистки логов (вызов sync для сохранения всех записанных логов)
	cleanup := func() {
		err := logger.Sync()
		if err != nil {
			panic("failed to sync logger: " + err.Error())
		}
	}

	// Возвращаем логгер и функцию очистки
	return logger, cleanup
}