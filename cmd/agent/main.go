package main

import (
	"context"
	"flag"
	"fmt"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sersus/go-yandex-metrics/internal/harvester"
	"github.com/sersus/go-yandex-metrics/internal/storage"
	"golang.org/x/sync/errgroup"
)

var options struct {
	address        string
	reportInterval int
	pollInterval   int
}

func init() {
	flag.StringVar(&options.address, "a", "localhost:8080", "Server listening address")
	flag.IntVar(&options.reportInterval, "r", 10, "report interval")
	flag.IntVar(&options.pollInterval, "p", 2, "poll interval")
}

func main() {
	flag.Parse()
	metricsCollector := harvester.New(&storage.MetricsStorage)

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

type Iharvester interface {
	Collect(metrics *runtime.MemStats)
}

func performCollect(h Iharvester) error {
	for {
		metrics := runtime.MemStats{}
		runtime.ReadMemStats(&metrics)
		h.Collect(&metrics)
		time.Sleep(time.Second * time.Duration(options.pollInterval))
	}
}

func sendMetricsToServer(client *resty.Client) error {
	for {
		for n, i := range storage.MetricsStorage.Metrics {
			switch i.Value.(type) {
			case uint, uint64, int, int64:
				_, err := client.R().
					SetHeader("Content-Type", "text/plain").
					Post(fmt.Sprintf("http://%s/update/%s/%s/%d", options.address, i.MetricType, n, i.Value))
				if err != nil {
					return err
				}
			case float64:
				_, err := client.R().
					SetHeader("Content-Type", "text/plain").
					Post(fmt.Sprintf("http://%s/update/%s/%s/%f", options.address, i.MetricType, n, i.Value))
				if err != nil {
					return err
				}
			}
		}
		time.Sleep(time.Second * time.Duration(options.reportInterval))
	}
}
