package harvester

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sersus/go-yandex-metrics/internal/config"
	"github.com/sersus/go-yandex-metrics/internal/storage"
)

func (h *harvester) Collect(metrics *runtime.MemStats) {
	h.storage.Metrics["Alloc"] = storage.Metric{Value: metrics.Alloc, MetricType: storage.Gauge}
	h.storage.Metrics["BuckHashSys"] = storage.Metric{Value: metrics.BuckHashSys, MetricType: storage.Gauge}
	h.storage.Metrics["Frees"] = storage.Metric{Value: metrics.Frees, MetricType: storage.Gauge}
	h.storage.Metrics["GCCPUFraction"] = storage.Metric{Value: metrics.GCCPUFraction, MetricType: storage.Gauge}
	h.storage.Metrics["GCSys"] = storage.Metric{Value: metrics.GCSys, MetricType: storage.Gauge}
	h.storage.Metrics["HeapAlloc"] = storage.Metric{Value: metrics.HeapAlloc, MetricType: storage.Gauge}
	h.storage.Metrics["HeapIdle"] = storage.Metric{Value: metrics.HeapIdle, MetricType: storage.Gauge}
	h.storage.Metrics["HeapInuse"] = storage.Metric{Value: metrics.HeapInuse, MetricType: storage.Gauge}
	h.storage.Metrics["HeapObjects"] = storage.Metric{Value: metrics.HeapObjects, MetricType: storage.Gauge}
	h.storage.Metrics["HeapReleased"] = storage.Metric{Value: metrics.HeapReleased, MetricType: storage.Gauge}
	h.storage.Metrics["HeapSys"] = storage.Metric{Value: metrics.HeapSys, MetricType: storage.Gauge}
	h.storage.Metrics["Lookups"] = storage.Metric{Value: metrics.Lookups, MetricType: storage.Gauge}
	h.storage.Metrics["MCacheInuse"] = storage.Metric{Value: metrics.MCacheInuse, MetricType: storage.Gauge}
	h.storage.Metrics["MCacheSys"] = storage.Metric{Value: metrics.MCacheSys, MetricType: storage.Gauge}
	h.storage.Metrics["MSpanInuse"] = storage.Metric{Value: metrics.MSpanInuse, MetricType: storage.Gauge}
	h.storage.Metrics["MSpanSys"] = storage.Metric{Value: metrics.MSpanSys, MetricType: storage.Gauge}
	h.storage.Metrics["Mallocs"] = storage.Metric{Value: metrics.Mallocs, MetricType: storage.Gauge}
	h.storage.Metrics["NextGC"] = storage.Metric{Value: metrics.NextGC, MetricType: storage.Gauge}
	h.storage.Metrics["NumForcedGC"] = storage.Metric{Value: metrics.NumForcedGC, MetricType: storage.Gauge}
	h.storage.Metrics["NumGC"] = storage.Metric{Value: metrics.NumGC, MetricType: storage.Gauge}
	h.storage.Metrics["OtherSys"] = storage.Metric{Value: metrics.OtherSys, MetricType: storage.Gauge}
	h.storage.Metrics["PauseTotalNs"] = storage.Metric{Value: metrics.PauseTotalNs, MetricType: storage.Gauge}
	h.storage.Metrics["StackInuse"] = storage.Metric{Value: metrics.StackInuse, MetricType: storage.Gauge}
	h.storage.Metrics["StackSys"] = storage.Metric{Value: metrics.StackSys, MetricType: storage.Gauge}
	h.storage.Metrics["Sys"] = storage.Metric{Value: metrics.Sys, MetricType: storage.Gauge}
	h.storage.Metrics["TotalAlloc"] = storage.Metric{Value: metrics.TotalAlloc, MetricType: storage.Gauge}
	h.storage.Metrics["RandomValue"] = storage.Metric{Value: rand.Int(), MetricType: storage.Gauge}

	var cnt int64
	if h.storage.Metrics["PollCount"].Value != nil {
		cnt = h.storage.Metrics["PollCount"].Value.(int64) + 1
	}
	h.storage.Metrics["PollCount"] = storage.Metric{Value: cnt, MetricType: storage.Counter}
}

func NewHarvester(ms *storage.MemStorage) *harvester {
	return &harvester{ms}
}

type harvester struct {
	storage *storage.MemStorage
}

type Iharvester interface {
	Collect(metrics *runtime.MemStats)
}

func PerformCollect(h Iharvester, pollInterval time.Duration) error {
	for {
		metrics := runtime.MemStats{}
		runtime.ReadMemStats(&metrics)
		h.Collect(&metrics)
		time.Sleep(time.Second * pollInterval)
	}
}

func SendMetricsToServer(client *resty.Client, options *config.Options) error {
	for {
		for n, i := range storage.MetricsStorage.Metrics {
			switch i.Value.(type) {
			case uint, uint64, int, int64:
				_, err := client.R().
					SetHeader("Content-Type", "text/plain").
					Post(fmt.Sprintf("http://%s/update/%s/%s/%d", options.Address, i.MetricType, n, i.Value))
				if err != nil {
					return err
				}
			case float64:
				_, err := client.R().
					SetHeader("Content-Type", "text/plain").
					Post(fmt.Sprintf("http://%s/update/%s/%s/%f", options.Address, i.MetricType, n, i.Value))
				if err != nil {
					return err
				}
			}
		}
		time.Sleep(time.Second * time.Duration(options.ReportInterval))
	}
}
