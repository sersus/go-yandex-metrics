package main

import (
	"flag"
	"net/http"

	"database/sql"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
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
	//flag.StringVar(&options.ConnectDB, "d", "host=localhost port=5432 user=postgres password=mysecretpassword dbname=metrics sslmode=disable", "database connection")
	flag.StringVar(&options.ConnectDB, "d", "postgres://postgres:mysecretpassword@localhost:5432/metrics?sslmode=disable", "database connection")
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	middleware.SugarLogger = *logger.Sugar()

	db, err := sql.Open("pgx", options.ConnectDB)
	if err != nil {
		middleware.SugarLogger.Error(err.Error(), "Failed to connect to the database:")
		return
	}
	defer db.Close()

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
	r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		if err := db.Ping(); err != nil {
			//fmt.Println("Database ping failed:", err)
			middleware.SugarLogger.Error(err.Error(), "Database ping failed")
			middleware.SugarLogger.Infof("Connection string  %s", options.ConnectDB)
			http.Error(w, "Database connection error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Database connection is OK"))
	})

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
