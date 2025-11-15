package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// global глобальный экземпляр логгера.
	global       *zap.SugaredLogger
	defaultLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
)

func init() {
	// Инициализируем дефолтный логгер без service name
	// Реальная инициализация должна быть через Init()
	SetLogger(New(defaultLevel, zap.AddStacktrace(zap.FatalLevel)))
}

// Init инициализирует глобальный логгер с указанным service name и уровнем логирования
// Возвращает cleanup функцию для корректного завершения работы логгера
func Init(serviceName string, level zapcore.Level) (func(), error) {
	if serviceName == "" {
		return nil, fmt.Errorf("service name cannot be empty")
	}

	// Создаем новый логгер с указанным уровнем
	atomicLevel := zap.NewAtomicLevelAt(level)
	logger := New(atomicLevel,
		zap.AddStacktrace(zap.FatalLevel),
	).With("service", serviceName)

	// Устанавливаем как глобальный
	SetLogger(logger)
	defaultLevel.SetLevel(level)

	// Cleanup функция для корректного завершения работы логгера
	cleanup := func() {
		if logger != nil {
			_ = logger.Sync()
		}
	}

	return cleanup, nil
}

// ParseLevel преобразует строковое значение уровня логирования в zapcore.Level
func ParseLevel(levelStr string) (zapcore.Level, error) {
	switch strings.ToLower(levelStr) {
	case "debug":
		return zapcore.DebugLevel, nil
	case "info":
		return zapcore.InfoLevel, nil
	case "warn", "warning":
		return zapcore.WarnLevel, nil
	case "error":
		return zapcore.ErrorLevel, nil
	case "fatal":
		return zapcore.FatalLevel, nil
	case "panic":
		return zapcore.PanicLevel, nil
	default:
		return zapcore.InfoLevel, fmt.Errorf("unknown log level: %s", levelStr)
	}
}

// GetLevelByEnvironment возвращает уровень логирования в зависимости от окружения
func GetLevelByEnvironment(environment string) zapcore.Level {
	switch strings.ToLower(environment) {
	case "dev", "development":
		return zapcore.DebugLevel
	case "stage", "staging":
		return zapcore.InfoLevel
	case "prod", "production":
		return zapcore.WarnLevel
	default:
		return zapcore.InfoLevel
	}
}

func New(level zapcore.LevelEnabler, options ...zap.Option) *zap.SugaredLogger {
	return NewWithSink(level, os.Stdout, options...)
}

func NewWithSink(level zapcore.LevelEnabler, sink io.Writer, options ...zap.Option) *zap.SugaredLogger {
	if level == nil {
		level = defaultLevel
	}

	core := newZapCore(level, sink)

	return zap.New(core, options...).Sugar()
}

func newZapCore(level zapcore.LevelEnabler, sink io.Writer) zapcore.Core {
	return zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			TimeKey:        "ts",
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
		}),
		zapcore.AddSync(sink),
		level,
	)
}

// Level возвращает текущий уровень логгирования глобального логгера.
func Level() zapcore.Level {
	return defaultLevel.Level()
}

// SetLevel устанавливает уровень логгирования глобального логгера.
func SetLevel(l zapcore.Level) {
	defaultLevel.SetLevel(l)
}

// Logger возвращает глобальный логгер.
func Logger() *zap.SugaredLogger {
	return global
}

// SetLogger устанавливает глобальный логгер. Функция непотокобезопасна.
func SetLogger(l *zap.SugaredLogger) {
	global = l
}

func Debug(ctx context.Context, args ...interface{}) {
	if logger := FromContext(ctx); logger.Level().Enabled(zapcore.DebugLevel) {
		logger.Debug(args...)
	}
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	if logger := FromContext(ctx); logger.Level().Enabled(zapcore.DebugLevel) {
		logger.Debugf(format, args...)
	}
}

func DebugKV(ctx context.Context, message string, kvs ...interface{}) {
	if logger := FromContext(ctx); logger.Level().Enabled(zapcore.DebugLevel) {
		logger.Debugw(message, kvs...)
	}
}

func Info(ctx context.Context, args ...interface{}) {
	if logger := FromContext(ctx); logger.Level().Enabled(zapcore.InfoLevel) {
		logger.Info(args...)
	}
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	if logger := FromContext(ctx); logger.Level().Enabled(zapcore.InfoLevel) {
		logger.Infof(format, args...)
	}
}

func InfoKV(ctx context.Context, message string, kvs ...interface{}) {
	if logger := FromContext(ctx); logger.Level().Enabled(zapcore.InfoLevel) {
		logger.Infow(message, kvs...)
	}
}

func Warn(ctx context.Context, args ...interface{}) {
	if logger := FromContext(ctx); logger.Level().Enabled(zapcore.WarnLevel) {
		logger.Warn(args...)
	}
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	if logger := FromContext(ctx); logger.Level().Enabled(zapcore.WarnLevel) {
		logger.Warnf(format, args...)
	}
}

func WarnKV(ctx context.Context, message string, kvs ...interface{}) {
	if logger := FromContext(ctx); logger.Level().Enabled(zapcore.WarnLevel) {
		logger.Warnw(message, kvs...)
	}
}

func Error(ctx context.Context, args ...interface{}) {
	if logger := FromContext(ctx); logger.Level().Enabled(zapcore.ErrorLevel) {
		logger.Error(args...)
	}
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	if logger := FromContext(ctx); logger.Level().Enabled(zapcore.ErrorLevel) {
		logger.Errorf(format, args...)
	}
}

func ErrorKV(ctx context.Context, message string, kvs ...interface{}) {
	if logger := FromContext(ctx); logger.Level().Enabled(zapcore.ErrorLevel) {
		logger.Errorw(message, kvs...)
	}
}

func Fatal(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Fatal(args...)
}

func Fatalf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Fatalf(format, args...)
}

func FatalKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Fatalw(message, kvs...)
}

func Panic(ctx context.Context, args ...interface{}) {
	FromContext(ctx).Panic(args...)
}

func Panicf(ctx context.Context, format string, args ...interface{}) {
	FromContext(ctx).Panicf(format, args...)
}

func PanicKV(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Panicw(message, kvs...)
}

func Audit(ctx context.Context, message string, kvs ...interface{}) {
	FromContext(ctx).Errorw(message, kvs...)
}
