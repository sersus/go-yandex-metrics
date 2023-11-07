package main

import (
	"flag"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sersus/go-yandex-metrics/internal/config"
	"github.com/sersus/go-yandex-metrics/internal/handlers"
	"github.com/sersus/go-yandex-metrics/internal/middleware"

	"github.com/sersus/go-yandex-metrics/internal/storage"
	"go.uber.org/zap"
)

var options config.ServerOptions

var metricsFromFile = storage.MetricsStorage

func init() {
	flag.StringVar(&options.Address, "a", "localhost:8080", "Server listening address")
	flag.IntVar(&options.StoreInterval, "i", 300, "store interval")
	flag.StringVar(&options.FileStoragePath, "f", "/tmp/metrics-db.json", "file path")
	flag.BoolVar(&options.Restore, "r", true, "restore metrics from file on start")
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	middleware.SugarLogger = *logger.Sugar()

	config.ParceServerFlags(&options)
	metricsHandler := &handlers.MetricsHandler{}
	r := chi.NewRouter()
	r.Use(middleware.RequestLogger)
	r.Use(middleware.Compress)
	r.Post("/update/", metricsHandler.SaveMetricFromJSON)
	r.Post("/value/", metricsHandler.GetMetricFromJSON)
	r.Post("/update/*", metricsHandler.SendMetric)
	r.Get("/value/*", metricsHandler.GetMetric)
	r.Get("/", metricsHandler.ShowMetrics)

	middleware.SugarLogger.Infow(
		"Starting server",
		"addr", options.Address,
	)
	if options.Restore {
		if err := metricsFromFile.Restore(options.FileStoragePath); err != nil {
			middleware.SugarLogger.Error(err.Error(), "restore error")
		}
	}

	go storage.SaveByInterval(&metricsFromFile, options.FileStoragePath, options.StoreInterval)
	if err := http.ListenAndServe(options.Address, r); err != nil {
		middleware.SugarLogger.Fatalw(err.Error(), "event", "start server")
	}
}
