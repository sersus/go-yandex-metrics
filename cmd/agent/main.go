package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sersus/go-yandex-metrics/internal/metric_handler"
	"github.com/sersus/go-yandex-metrics/internal/storage"
	"golang.org/x/sync/errgroup"
)

func main() {
	metricsCollector := metric_handler.New(&storage.MetricsStorage)

	ctx := context.Background()

	errs, _ := errgroup.WithContext(ctx)
	errs.Go(func() error {
		if err := performCollect(metricsCollector); err != nil {
			panic(err)
		}
		return nil
	})

	reportTicker := time.NewTicker(time.Second * 10)
	client := resty.New()
	defer reportTicker.Stop()
	errs.Go(func() error {
		if err := sendMetricsToServer(client); err != nil {
			panic(err)
		}
		return nil
	})

	_ = errs.Wait()
}

type Imetric_handler interface {
	Collect(metrics *runtime.MemStats)
}

func performCollect(metricsCollector Imetric_handler) error {
	for {
		metrics := runtime.MemStats{}
		runtime.ReadMemStats(&metrics)
		metricsCollector.Collect(&metrics)
		time.Sleep(time.Second * 2)
	}
}

func sendMetricsToServer(client *resty.Client) error {
	for {
		for n, i := range storage.MetricsStorage.Metrics {
			switch i.Value.(type) {
			case uint, uint64, int, int64:
				_, err := client.R().
					SetHeader("Content-Type", "text/plain").
					Post(fmt.Sprintf("http://localhost:8080/update/%s/%s/%d", i.MetricType, n, i.Value))
				if err != nil {
					return err
				}
			case float64:
				_, err := client.R().
					SetHeader("Content-Type", "text/plain").
					Post(fmt.Sprintf("http://localhost:8080/update/%s/%s/%f", i.MetricType, n, i.Value))
				if err != nil {
					return err
				}
			}
		}
		time.Sleep(time.Second * 10)
	}
}
