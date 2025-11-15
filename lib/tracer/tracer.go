package tracer

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// Config содержит настройки трейсинга
type Config struct {
	Enabled         bool
	ServiceName     string
	JaegerHost      string
	JaegerAgentHost string
	JaegerAgentPort int
	SamplerType     string
	SamplerParam    float64
}

// Init инициализирует глобальный tracer с указанной конфигурацией (OpenTelemetry)
// Возвращает cleanup функцию для корректного завершения работы tracer'а
func Init(cfg Config) (func(), error) {
	// Если трейсинг отключен, возвращаем пустую cleanup функцию
	if !cfg.Enabled {
		return func() {}, nil
	}

	// Валидация обязательных полей
	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("service name cannot be empty when tracer is enabled")
	}

	ctx := context.Background()

	// Создаем resource с информацией о сервисе
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Определяем endpoint для OTLP exporter
	var endpoint string
	if cfg.JaegerHost != "" {
		// Используем старый формат с JAEGER_HOST (host:port)
		endpoint = cfg.JaegerHost
	} else if cfg.JaegerAgentHost != "" {
		// Используем новый формат с разделением host и port
		endpoint = fmt.Sprintf("%s:%d", cfg.JaegerAgentHost, cfg.JaegerAgentPort)
	} else {
		return nil, fmt.Errorf("either jaeger_host or jaeger_agent_host must be specified")
	}

	// Создаем OTLP gRPC exporter для Jaeger
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(), // для dev окружения
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP exporter: %w", err)
	}

	// Определяем sampler
	var sampler sdktrace.Sampler
	switch cfg.SamplerType {
	case "const":
		if cfg.SamplerParam >= 1.0 {
			sampler = sdktrace.AlwaysSample()
		} else {
			sampler = sdktrace.NeverSample()
		}
	case "probabilistic":
		sampler = sdktrace.TraceIDRatioBased(cfg.SamplerParam)
	default:
		sampler = sdktrace.AlwaysSample()
	}

	// Создаем trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	)

	// Устанавливаем глобальный trace provider
	otel.SetTracerProvider(tp)

	// Устанавливаем propagator для распространения trace context
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Cleanup функция для корректного завершения работы tracer'а
	cleanup := func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			// Логируем ошибку, но не паникуем
			_ = err
		}
	}

	return cleanup, nil
}
