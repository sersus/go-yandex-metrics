package main

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/sersus/go-yandex-metrics/internal/config"
	"github.com/sersus/go-yandex-metrics/internal/harvester"
	"github.com/sersus/go-yandex-metrics/internal/middleware"
	"github.com/sersus/go-yandex-metrics/internal/storage"
	"golang.org/x/sync/errgroup"
)

func main() {
	params := config.Init(
		config.WithPollInterval(),
		config.WithReportInterval(),
		config.WithAddr(),
		config.WithKey(),
		config.WithRateLimit(),
	)
	ctx := context.Background()

	errs, _ := errgroup.WithContext(ctx)

	logger, err := zap.NewDevelopment()
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync()
	middleware.SugarLogger = *logger.Sugar()

	errs.Go(func() error {
		h := harvester.New(&storage.MetricStorage)
		for {
			h.HarvestRuntime()
			h.HarvestGopsutil()
			time.Sleep(time.Duration(params.PollInterval) * time.Second)
		}
	})

	sender := harvester.InitSender(params, middleware.SugarLogger)
	errs.Go(func() error {
		if err := sender.SendMetrics(ctx); err != nil {
			middleware.SugarLogger.Fatalln(err)
		}
		return nil
	})

	_ = errs.Wait()
}
