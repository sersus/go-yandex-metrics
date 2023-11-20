package main

import (
	"context"
	"net/http"
	"time"

	"github.com/sersus/go-yandex-metrics/internal/config"
	log "github.com/sersus/go-yandex-metrics/internal/middleware"
	"github.com/sersus/go-yandex-metrics/internal/router/router"
	"github.com/sersus/go-yandex-metrics/internal/storage"
	"github.com/sersus/go-yandex-metrics/internal/storager"
	"go.uber.org/zap"
)

type saver interface {
	Restore(ctx context.Context) ([]storage.Metric, error)
	Save(ctx context.Context, metrics []storage.Metric) error
}

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	log.SugarLogger = *logger.Sugar()

	params := config.Init(
		config.WithAddr(),
		config.WithStoreInterval(),
		config.WithFileStoragePath(),
		config.WithRestore(),
		config.WithDatabase(),
	)

	r := router.New(*params)

	log.SugarLogger.Infow(
		"Starting server",
		"addr", params.FlagRunAddr,
	)

	// init restorer
	var saver saver
	if params.FileStoragePath != "" && params.DatabaseAddress == "" {
		saver = storager.NewFilesaver(params)
	} else if params.DatabaseAddress != "" {
		saver, err = storager.NewDBSaver(params)
		if err != nil {
			log.SugarLogger.Errorf(err.Error())
		}
	}

	// restore previous metrics if needed
	ctx := context.Background()
	if params.Restore && (params.FileStoragePath != "" || params.DatabaseAddress != "") {
		metrics, err := saver.Restore(ctx)
		if err != nil {
			log.SugarLogger.Error(err.Error(), "restore error")
		}
		storage.Collector.Metrics = metrics
		log.SugarLogger.Info("metrics restored")
	}

	// regularly save metrics if needed
	if params.DatabaseAddress != "" || params.FileStoragePath != "" {
		go saveMetrics(ctx, saver, params.StoreInterval)
	}

	// run server
	if err := http.ListenAndServe(params.FlagRunAddr, r); err != nil {
		log.SugarLogger.Fatalw(err.Error(), "event", "start server")
	}
}

func saveMetrics(ctx context.Context, saver saver, interval int) {
	ticker := time.NewTicker(time.Duration(interval))
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := saver.Save(ctx, storage.Collector.Metrics); err != nil {
				log.SugarLogger.Error(err.Error(), "save error")
			}
		}
	}
}
