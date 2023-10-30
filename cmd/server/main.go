package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/sersus/go-yandex-metrics/internal/handlers"
	log "github.com/sersus/go-yandex-metrics/internal/logger"
	"go.uber.org/zap"
)

var address = flag.String("a", "localhost:8080", "Server listening address")

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	log.SugarLogger = *logger.Sugar()

	flag.Parse()
	envAddr, exists := os.LookupEnv("ADDRESS")
	if exists && envAddr != "" {
		*address = envAddr
	}
	metricsHandler := &handlers.MetricsHandler{}
	r := chi.NewRouter()
	r.Use(log.RequestLogger)
	r.Post("/update/", metricsHandler.SaveMetricFromJSON)
	r.Post("/value/", metricsHandler.GetMetricFromJSON)
	r.Post("/update/*", metricsHandler.SendMetric)
	r.Get("/value/*", metricsHandler.GetMetric)
	r.Get("/", metricsHandler.ShowMetrics)

	log.SugarLogger.Infow(
		"Starting server",
		"addr", *address,
	)
	if err := http.ListenAndServe(*address, r); err != nil {
		// записываем в лог ошибку, если сервер не запустился
		log.SugarLogger.Fatalw(err.Error(), "event", "start server")
	}
}
