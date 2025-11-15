package metrics

import (
	"fmt"
	"strconv"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Config содержит настройки метрик
type Config struct {
	Enabled     bool
	ServiceName string
	Namespace   string
	Subsystem   string
}

var (
	registry    *prometheus.Registry
	serviceName string
	namespace   string
	subsystem   string
	initialized bool
)

var ms struct {
	serverMetrics         *grpc_prometheus.ServerMetrics
	responseTimeHistogram *prometheus.HistogramVec
	requestsCount         *prometheus.CounterVec
}

// Init инициализирует метрики и возвращает cleanup функцию
func Init(cfg Config) (func(), error) {
	// Если метрики отключены, возвращаем пустую cleanup функцию
	if !cfg.Enabled {
		return func() {}, nil
	}

	// Валидация обязательных полей
	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("service name cannot be empty when metrics are enabled")
	}

	// Устанавливаем значения из конфигурации
	serviceName = cfg.ServiceName
	namespace = cfg.Namespace
	if namespace == "" {
		namespace = "balun_courses"
	}
	subsystem = cfg.Subsystem
	if subsystem == "" {
		subsystem = "grpc"
	}

	// Создаем новый registry
	registry = prometheus.NewRegistry()

	// Инициализируем метрики
	ms.serverMetrics = grpc_prometheus.NewServerMetrics(
		grpc_prometheus.WithServerHandlingTimeHistogram(
			grpc_prometheus.WithHistogramBuckets([]float64{0.001, 0.01, 0.1, 0.3, 0.6, 1, 3, 6, 9, 20, 30, 60, 90, 120}),
		),
	)

	ms.responseTimeHistogram = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "histogram_response_time_seconds",
			Help:      "Время ответа от сервера",
			Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 16),
		},
		[]string{"service", "method", "is_error"},
	)

	ms.requestsCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "requests",
		Help:      "Количество запросов",
	}, []string{
		"service", "method",
	})

	// Регистрируем метрики
	registry.MustRegister(
		ms.serverMetrics,
		ms.responseTimeHistogram,
		ms.requestsCount,
	)

	initialized = true

	// Cleanup функция
	cleanup := func() {
		// Сбрасываем флаг инициализации
		initialized = false

		// Очищаем registry и сбрасываем глобальные переменные
		if registry != nil {
			registry = nil
		}

		// Сбрасываем метрики
		ms.serverMetrics = nil
		ms.responseTimeHistogram = nil
		ms.requestsCount = nil

		// Сбрасываем конфигурацию
		serviceName = ""
		namespace = ""
		subsystem = ""
	}

	return cleanup, nil
}

// GetRegistry возвращает registry для использования в HTTP handler
func GetRegistry() *prometheus.Registry {
	if !initialized {
		// Возвращаем пустой registry если не инициализированы
		return prometheus.NewRegistry()
	}
	return registry
}

// ResponseTimeHistogramObserve записывает время ответа в гистограмму
func ResponseTimeHistogramObserve(method string, err error, d time.Duration) {
	if !initialized {
		return
	}
	isError := strconv.FormatBool(err != nil)
	ms.responseTimeHistogram.WithLabelValues(serviceName, method, isError).Observe(d.Seconds())
}

// IncRequests увеличивает счетчик запросов
func IncRequests(method string) {
	if !initialized {
		return
	}
	ms.requestsCount.WithLabelValues(serviceName, method).Inc()
}
