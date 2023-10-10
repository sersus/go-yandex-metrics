package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sersus/go-yandex-metrics/internal/handlers"
)

var address = flag.String("a", "localhost:8080", "Server listening address")

func main() {
	flag.Parse()
	r := chi.NewRouter()
	r.Post("/update/*", handlers.SendMetric)
	r.Get("/value/*", handlers.GetMetric)
	r.Get("/", handlers.ShowMetrics)

	//log.Fatal(http.ListenAndServe(":8080", r))
	log.Fatal(http.ListenAndServe(*address, r))
}
