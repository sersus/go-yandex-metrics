package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/sersus/go-yandex-metrics/internal/handlers"
)

var address = flag.String("a", "localhost:8080", "Server listening address")

func main() {
	flag.Parse()
	envAddr, exists := os.LookupEnv("ADDRESS")
	if exists && envAddr != "" {
		*address = envAddr
	}
	metricsHandler := &handlers.MetricsHandler{}
	r := chi.NewRouter()
	r.Post("/update/*", metricsHandler.SendMetric)
	r.Get("/value/*", metricsHandler.GetMetric)
	r.Get("/", metricsHandler.ShowMetrics)

	//log.Fatal(http.ListenAndServe(":8080", r))
	log.Fatal(http.ListenAndServe(*address, r))
}
