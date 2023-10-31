package harvester

import (
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/avast/retry-go"
	"github.com/go-resty/resty/v2"
	"github.com/sersus/go-yandex-metrics/internal/config"
	"github.com/sersus/go-yandex-metrics/internal/storage"
)

func (h *harvester) Collect(metrics *runtime.MemStats) {
	h.storage.Metrics["Alloc"] = storage.Metric{Value: float64(metrics.Alloc), MetricType: storage.Gauge}
	h.storage.Metrics["BuckHashSys"] = storage.Metric{Value: float64(metrics.BuckHashSys), MetricType: storage.Gauge}
	h.storage.Metrics["Frees"] = storage.Metric{Value: float64(metrics.Frees), MetricType: storage.Gauge}
	h.storage.Metrics["GCCPUFraction"] = storage.Metric{Value: metrics.GCCPUFraction, MetricType: storage.Gauge}
	h.storage.Metrics["GCSys"] = storage.Metric{Value: float64(metrics.GCSys), MetricType: storage.Gauge}
	h.storage.Metrics["HeapAlloc"] = storage.Metric{Value: float64(metrics.HeapAlloc), MetricType: storage.Gauge}
	h.storage.Metrics["HeapIdle"] = storage.Metric{Value: float64(metrics.HeapIdle), MetricType: storage.Gauge}
	h.storage.Metrics["HeapInuse"] = storage.Metric{Value: float64(metrics.HeapInuse), MetricType: storage.Gauge}
	h.storage.Metrics["HeapObjects"] = storage.Metric{Value: float64(metrics.HeapObjects), MetricType: storage.Gauge}
	h.storage.Metrics["HeapReleased"] = storage.Metric{Value: float64(metrics.HeapReleased), MetricType: storage.Gauge}
	h.storage.Metrics["HeapSys"] = storage.Metric{Value: float64(metrics.HeapSys), MetricType: storage.Gauge}
	h.storage.Metrics["Lookups"] = storage.Metric{Value: float64(metrics.Lookups), MetricType: storage.Gauge}
	h.storage.Metrics["MCacheInuse"] = storage.Metric{Value: float64(metrics.MCacheInuse), MetricType: storage.Gauge}
	h.storage.Metrics["MCacheSys"] = storage.Metric{Value: float64(metrics.MCacheSys), MetricType: storage.Gauge}
	h.storage.Metrics["MSpanInuse"] = storage.Metric{Value: float64(metrics.MSpanInuse), MetricType: storage.Gauge}
	h.storage.Metrics["MSpanSys"] = storage.Metric{Value: float64(metrics.MSpanSys), MetricType: storage.Gauge}
	h.storage.Metrics["Mallocs"] = storage.Metric{Value: float64(metrics.Mallocs), MetricType: storage.Gauge}
	h.storage.Metrics["NextGC"] = storage.Metric{Value: float64(metrics.NextGC), MetricType: storage.Gauge}
	h.storage.Metrics["LastGC"] = storage.Metric{Value: float64(metrics.LastGC), MetricType: storage.Gauge}
	h.storage.Metrics["NumForcedGC"] = storage.Metric{Value: float64(metrics.NumForcedGC), MetricType: storage.Gauge}
	h.storage.Metrics["NumGC"] = storage.Metric{Value: float64(metrics.NumGC), MetricType: storage.Gauge}
	h.storage.Metrics["OtherSys"] = storage.Metric{Value: float64(metrics.OtherSys), MetricType: storage.Gauge}
	h.storage.Metrics["PauseTotalNs"] = storage.Metric{Value: float64(metrics.PauseTotalNs), MetricType: storage.Gauge}
	h.storage.Metrics["StackInuse"] = storage.Metric{Value: float64(metrics.StackInuse), MetricType: storage.Gauge}
	h.storage.Metrics["StackSys"] = storage.Metric{Value: float64(metrics.StackSys), MetricType: storage.Gauge}
	h.storage.Metrics["Sys"] = storage.Metric{Value: float64(metrics.Sys), MetricType: storage.Gauge}
	h.storage.Metrics["TotalAlloc"] = storage.Metric{Value: float64(metrics.TotalAlloc), MetricType: storage.Gauge}
	h.storage.Metrics["RandomValue"] = storage.Metric{Value: float64(rand.Int()), MetricType: storage.Gauge}

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
			switch i.MetricType {
			case storage.Counter:
				req := client.R().
					SetHeader("Content-Type", "application/json").SetBody(fmt.Sprintf(`{"id":%q, "type":"counter", "delta":%d}`, n, i.Value))
				//Post(fmt.Sprintf("http://%s/update/%s/%s/%d", options.Address, i.MetricType, n, i.Value))
				if err := sendRequest(req, options.Address); err != nil {
					return err
				}
			case storage.Gauge:
				req := client.R().
					SetHeader("Content-Type", "application/json").SetBody(fmt.Sprintf(`{"id":%q, "type":"gauge", "value":%f}`, n, i.Value))
				//Post(fmt.Sprintf("http://%s/update/%s/%s/%f", options.Address, i.MetricType, n, i.Value))
				if err := sendRequest(req, options.Address); err != nil {
					return err
				}
			}
		}
		time.Sleep(time.Second * time.Duration(options.ReportInterval))
	}
}

func sendRequest(req *resty.Request, addr string) error {
	err := retry.Do(
		func() error {
			var err error
			_, err = req.Post(fmt.Sprintf("http://%s/update/", addr))
			return err
		},
		retry.Attempts(10),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Retrying request after error: %v", err)
		}),
	)
	if err != nil {
		return err
	}
	// do something with the response
	return nil
}
