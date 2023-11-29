package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/sersus/go-yandex-metrics/internal/config"
	"github.com/sersus/go-yandex-metrics/internal/middleware"
	"github.com/sersus/go-yandex-metrics/internal/router/handlers"
)

func New(params config.Options) *chi.Mux {
	handler := handlers.New(params.DatabaseAddress, params.Key)

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger)
	r.Use(middleware.Compress)
	r.Post("/update/", handler.SaveMetricFromJSON)
	r.Post("/value/", handler.GetMetricFromJSON)
	r.Post("/update/{type}/{name}/{value}", handler.SaveMetric)
	r.Get("/value/{type}/{name}", handler.GetMetric)
	r.Get("/", handler.ShowMetrics)
	r.Get("/ping", handler.Ping)
	r.Post("/updates/", handler.SaveListMetricsFromJSON)

	return r
}
