package admin

import (
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sskorolev/balun_microservices/lib/metrics"
)

// RegisterHandlers регистрирует все HTTP handlers для admin сервера
func RegisterHandlers(mux *http.ServeMux, cfg Config) {
	// Регистрируем эндпоинт метрик
	if cfg.Metrics.Enabled {
		registerMetricsHandler(mux, cfg.Metrics)
	}

	// Регистрируем pprof эндпоинты
	if cfg.Pprof.Enabled {
		registerPprofHandlers(mux, cfg.Pprof)
	}

	// Регистрируем health check эндпоинты
	registerHealthHandlers(mux)
}

// registerMetricsHandler регистрирует Prometheus metrics handler
func registerMetricsHandler(mux *http.ServeMux, cfg MetricsConfig) {
	path := cfg.Path
	if path == "" {
		path = DefaultMetricsPath
	}

	// Получаем registry из lib/metrics
	registry := metrics.GetRegistry()

	// Создаем Prometheus HTTP handler
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
		Registry:          registry,
	})

	mux.Handle(path, handler)
}

// registerPprofHandlers регистрирует pprof handlers
func registerPprofHandlers(mux *http.ServeMux, cfg PprofConfig) {
	basePath := cfg.Path
	if basePath == "" {
		basePath = DefaultPprofPath
	}

	// Регистрируем все pprof endpoints
	mux.HandleFunc(basePath+"/", pprof.Index)
	mux.HandleFunc(basePath+"/cmdline", pprof.Cmdline)
	mux.HandleFunc(basePath+"/profile", pprof.Profile)
	mux.HandleFunc(basePath+"/symbol", pprof.Symbol)
	mux.HandleFunc(basePath+"/trace", pprof.Trace)

	// Handler'ы для конкретных профилей
	mux.Handle(basePath+"/allocs", pprof.Handler("allocs"))
	mux.Handle(basePath+"/block", pprof.Handler("block"))
	mux.Handle(basePath+"/goroutine", pprof.Handler("goroutine"))
	mux.Handle(basePath+"/heap", pprof.Handler("heap"))
	mux.Handle(basePath+"/mutex", pprof.Handler("mutex"))
	mux.Handle(basePath+"/threadcreate", pprof.Handler("threadcreate"))
}

// registerHealthHandlers регистрирует health check endpoints
func registerHealthHandlers(mux *http.ServeMux) {
	// Liveness probe - сервис живой
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Readiness probe - сервис готов принимать запросы
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}
