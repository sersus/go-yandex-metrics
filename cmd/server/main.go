package main

import (
	"net/http"

	"github.com/sersus/go-yandex-metrics/internal/config"
	"github.com/sersus/go-yandex-metrics/internal/middleware"
	"github.com/sersus/go-yandex-metrics/internal/router/router"
	"github.com/sersus/go-yandex-metrics/internal/storager"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	middleware.SugarLogger = *logger.Sugar()

	params := config.Init(
		config.WithAddr(),
		config.WithStoreInterval(),
		config.WithFileStoragePath(),
		config.WithRestore(),
		config.WithDatabase(),
	)

	r := router.New(*params)

	middleware.SugarLogger.Infow(
		"Starting server",
		"addr", params.FlagRunAddr,
	)

	// regularly save metrics if needed
	if params.DatabaseAddress != "" || params.FileStoragePath != "" {
		sh := storager.InitSaverHelper(params)
		go sh.SaveMetrics()
	}

	// run server
	if err := http.ListenAndServe(params.FlagRunAddr, r); err != nil {
		middleware.SugarLogger.Fatalw(err.Error(), "event", "start server")
	}
}
