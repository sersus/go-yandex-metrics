package main

import (
	"context"
	"log"
	"time"

	"github.com/sersus/go-yandex-metrics/internal/config"
	"github.com/sersus/go-yandex-metrics/internal/harvester"
	"github.com/sersus/go-yandex-metrics/internal/storage"
	"golang.org/x/sync/errgroup"
)

func main() {
	params := config.Init(
		config.WithPollInterval(),
		config.WithReportInterval(),
		config.WithAddr(),
		config.WithKey(),
	)
	ctx := context.Background()

	errs, _ := errgroup.WithContext(ctx)
	errs.Go(func() error {
		h := harvester.New(&storage.MetricStorage)
		for {
			h.Harvest()
			time.Sleep(time.Duration(params.PollInterval) * time.Second)
		}
	})

	sender := harvester.InitSender(params)
	errs.Go(func() error {
		if err := sender.SendMetricsToServer(); err != nil {
			log.Fatalln(err)
		}
		return nil
	})

	_ = errs.Wait()
}
