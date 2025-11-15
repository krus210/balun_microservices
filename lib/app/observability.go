package app

import (
	"fmt"

	"github.com/sskorolev/balun_microservices/lib/config"
	"github.com/sskorolev/balun_microservices/lib/logger"
	"github.com/sskorolev/balun_microservices/lib/metrics"
	"github.com/sskorolev/balun_microservices/lib/tracer"
	"go.uber.org/zap/zapcore"
)

// InitLogger инициализирует глобальный логгер с настройками из конфигурации
func (a *App) InitLogger(loggerCfg config.LoggerConfig, serviceName, environment string) error {
	// Определяем уровень логирования
	var level zapcore.Level
	var err error

	// Сначала пробуем получить из конфигурации
	if loggerCfg.Level != "" {
		level, err = logger.ParseLevel(loggerCfg.Level)
		if err != nil {
			return fmt.Errorf("failed to parse log level: %w", err)
		}
	} else {
		// Если уровень не указан, используем уровень по умолчанию для окружения
		level = logger.GetLevelByEnvironment(environment)
	}

	// Инициализируем логгер
	cleanup, err := logger.Init(serviceName, level)
	if err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Добавляем cleanup функцию
	a.cleanupFuncs = append(a.cleanupFuncs, cleanup)

	return nil
}

// InitTracer инициализирует глобальный tracer с настройками из конфигурации
func (a *App) InitTracer(tracerCfg config.TracerConfig) error {
	// Конвертируем конфигурацию из lib/config в lib/tracer
	cfg := tracer.Config{
		Enabled:         tracerCfg.Enabled,
		ServiceName:     tracerCfg.ServiceName,
		JaegerHost:      tracerCfg.JaegerHost,
		JaegerAgentHost: tracerCfg.JaegerAgentHost,
		JaegerAgentPort: tracerCfg.JaegerAgentPort,
		SamplerType:     tracerCfg.SamplerType,
		SamplerParam:    float64(tracerCfg.SamplerParam),
	}

	// Инициализируем tracer
	cleanup, err := tracer.Init(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize tracer: %w", err)
	}

	// Добавляем cleanup функцию
	a.cleanupFuncs = append(a.cleanupFuncs, cleanup)

	return nil
}

// InitMetrics инициализирует метрики с настройками из конфигурации
func (a *App) InitMetrics(metricsCfg config.MetricsConfig, serviceName string) error {
	// Конвертируем конфигурацию из lib/config в lib/metrics
	cfg := metrics.Config{
		Enabled:     metricsCfg.Enabled,
		ServiceName: serviceName,
		Namespace:   metricsCfg.Namespace,
		Subsystem:   metricsCfg.Subsystem,
	}

	// Инициализируем metrics
	cleanup, err := metrics.Init(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize metrics: %w", err)
	}

	// Добавляем cleanup функцию
	a.cleanupFuncs = append(a.cleanupFuncs, cleanup)

	return nil
}
