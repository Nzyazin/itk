package logger

import (
	"os"
	
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)
type Logger interface {
    Info(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Debug(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    With(fields ...Field) Logger
}

type Field struct {
	Key       string
	Type      zapcore.FieldType
	Integer   int64
	String    string
	Interface interface{}
}

type ZapAdapter struct {
    zapLogger *zap.Logger
}

func (za *ZapAdapter) With(fields ...Field) Logger {
	return &ZapAdapter{zapLogger: za.zapLogger.With(convertFields(fields)...)}
}

func (za *ZapAdapter) Info(msg string, fields ...Field) {
    za.zapLogger.Info(msg, convertFields(fields)...)
}

func (za *ZapAdapter) Error(msg string, fields ...Field) {
    za.zapLogger.Error(msg, convertFields(fields)...)
}

func (za *ZapAdapter) Debug(msg string, fields ...Field) {
    za.zapLogger.Debug(msg, convertFields(fields)...)
}

func (za *ZapAdapter) Warn(msg string, fields ...Field) {
    za.zapLogger.Warn(msg, convertFields(fields)...)
}

func NewLogger() (Logger, func()) {
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

func StringField(key string, value string) Field {
	return Field{Key: key, Type: zapcore.StringType, String: value}
}

func Int64Field(key string, value int64) Field {
	return Field{Key: key, Type: zapcore.Int64Type, Integer: value}
}

func AnyField(key string, val interface{}) Field {
	return Field{Key: key, Type: zapcore.ReflectType, Interface: val}
}

func ErrorField(key string, val error) Field {
	return Field{Key: key, Type: zapcore.ErrorType, Interface: val}
}

func convertFields(fields []Field) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		switch f.Type {
		case zapcore.StringType:
			zapFields = append(zapFields, zap.String(f.Key, f.String))
		case zapcore.Int64Type:
			zapFields = append(zapFields, zap.Int64(f.Key, f.Integer))
		case zapcore.ReflectType:
			zapFields = append(zapFields, zap.Any(f.Key, f.Interface))
		case zapcore.ErrorType:
			zapFields = append(zapFields, zap.Error(f.Interface.(error)))
		default:
			zapFields = append(zapFields, zap.Skip())
		}
	}
	return zapFields
}