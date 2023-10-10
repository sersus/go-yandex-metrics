package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/sersus/go-yandex-metrics/internal/handlers"
)

func main() {
	r := chi.NewRouter()
	r.Post("/update/*", handlers.SendMetric)
	r.Get("/value/*", handlers.GetMetric)
	r.Get("/", handlers.ShowMetrics)

	log.Fatal(http.ListenAndServe(":8080", r))

	//http.HandleFunc("/update/", handlers.SaveMetric)
	//http.HandleFunc("/value/", handlers.GetMetric)
	//http.HandleFunc("/", handlers.ShowMetrics)
	//
	//if err := http.ListenAndServe(`:8080`, nil); err != nil {
	//	panic(err)
	//}
}
