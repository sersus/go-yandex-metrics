package main

import (
	"context"
	"flag"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sersus/go-yandex-metrics/internal/config"
	"github.com/sersus/go-yandex-metrics/internal/harvester"
	"github.com/sersus/go-yandex-metrics/internal/storage"
	"golang.org/x/sync/errgroup"
)

var options config.Options

func init() {
	flag.StringVar(&options.Address, "a", "localhost:8080", "Server listening address")
	flag.IntVar(&options.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&options.PollInterval, "p", 2, "poll interval")
}

func main() {
	options = *config.ParceFlags()
	metricsCollector := harvester.NewHarvester(&storage.MetricsStorage)

	ctx := context.Background()

	errs, _ := errgroup.WithContext(ctx)
	errs.Go(func() error {
		//if err := performCollect(metricsCollector); err != nil {
		if err := harvester.PerformCollect(metricsCollector, time.Duration(options.PollInterval)); err != nil {
			panic(err)
		}
		return nil
	})

	reportTicker := time.NewTicker(time.Second * 10)
	client := resty.New()
	defer reportTicker.Stop()
	errs.Go(func() error {
		if err := harvester.SendMetricsToServer(client, options); err != nil {
			panic(err)
		}
		return nil
	})

	_ = errs.Wait()
}
