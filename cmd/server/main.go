package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sersus/go-yandex-metrics/internal/compressor"
	"github.com/sersus/go-yandex-metrics/internal/config"
	"github.com/sersus/go-yandex-metrics/internal/handlers"
	log "github.com/sersus/go-yandex-metrics/internal/logger"
	"github.com/sersus/go-yandex-metrics/internal/storage"
	"go.uber.org/zap"
)

//var address = flag.String("a", "localhost:8080", "Server listening address")

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

	log.SugarLogger = *logger.Sugar()

	config.ParceServerFlags(&options)
	//flag.Parse()
	//envAddr, exists := os.LookupEnv("ADDRESS")
	//if exists && envAddr != "" {
	//	*address = envAddr
	//}
	metricsHandler := &handlers.MetricsHandler{}
	r := chi.NewRouter()
	r.Use(log.RequestLogger)
	r.Use(compressor.Compress)
	r.Post("/update/", metricsHandler.SaveMetricFromJSON)
	r.Post("/value/", metricsHandler.GetMetricFromJSON)
	r.Post("/update/*", metricsHandler.SendMetric)
	r.Get("/value/*", metricsHandler.GetMetric)
	r.Get("/", metricsHandler.ShowMetrics)

	log.SugarLogger.Infow(
		"Starting server",
		"addr", options.Address,
	)
	if options.Restore {
		if err := restore(options.FileStoragePath); err != nil {
			log.SugarLogger.Error(err.Error(), "restore error")
		}
	}

	go func() {
		if options.FileStoragePath != "" {
			for {
				err = save(options.FileStoragePath)
				if err != nil {
					log.SugarLogger.Error(err.Error(), "save error")
				} else {
					log.SugarLogger.Info("successfully saved metrics")
				}
				time.Sleep(time.Duration(options.StoreInterval) * time.Second)
			}
		}
	}()
	//if err := http.ListenAndServe(*address, r); err != nil {
	if err := http.ListenAndServe(options.Address, r); err != nil {
		// записываем в лог ошибку, если сервер не запустился
		log.SugarLogger.Fatalw(err.Error(), "event", "start server")
	}
}

func restore(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return err
	}
	data := scanner.Bytes()

	if err = json.Unmarshal(data, &metricsFromFile); err != nil {
		return err
	}
	return nil
}

func save(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	data, err := json.Marshal(&metricsFromFile)
	if err != nil {
		return err
	}
	if _, err := writer.Write(data); err != nil {
		return err
	}
	if err := writer.WriteByte('\n'); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}
