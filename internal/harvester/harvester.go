package harvester

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"runtime"
	"time"

	"github.com/avast/retry-go"
	"github.com/go-resty/resty/v2"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"

	"github.com/sersus/go-yandex-metrics/internal/config"
	"github.com/sersus/go-yandex-metrics/internal/storage"
)

type Harvest struct {
	h Harvester
}

type Harvester interface {
	Collect(json storage.Metric) error
}

func New(harvester Harvester) *Harvest {
	return &Harvest{
		h: harvester,
	}
}

func (a *Harvest) HarvestRuntime() {
	metrics := runtime.MemStats{}
	runtime.ReadMemStats(&metrics)

	a.h.Collect(storage.Metric{ID: "Alloc", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.Alloc))})
	a.h.Collect(storage.Metric{ID: "BuckHashSys", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.BuckHashSys))})
	a.h.Collect(storage.Metric{ID: "Frees", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.Frees))})
	a.h.Collect(storage.Metric{ID: "GCCPUFraction", MType: storage.Gauge, Value: &metrics.GCCPUFraction})
	a.h.Collect(storage.Metric{ID: "GCSys", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.GCSys))})
	a.h.Collect(storage.Metric{ID: "HeapAlloc", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.HeapAlloc))})
	a.h.Collect(storage.Metric{ID: "HeapIdle", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.HeapIdle))})
	a.h.Collect(storage.Metric{ID: "HeapInuse", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.HeapInuse))})
	a.h.Collect(storage.Metric{ID: "HeapObjects", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.HeapObjects))})
	a.h.Collect(storage.Metric{ID: "HeapReleased", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.HeapReleased))})
	a.h.Collect(storage.Metric{ID: "HeapSys", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.HeapSys))})
	a.h.Collect(storage.Metric{ID: "Lookups", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.Lookups))})
	a.h.Collect(storage.Metric{ID: "MCacheInuse", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.MCacheInuse))})
	a.h.Collect(storage.Metric{ID: "MCacheSys", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.MCacheSys))})
	a.h.Collect(storage.Metric{ID: "MSpanInuse", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.MSpanInuse))})
	a.h.Collect(storage.Metric{ID: "MSpanSys", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.MSpanSys))})
	a.h.Collect(storage.Metric{ID: "Mallocs", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.Mallocs))})
	a.h.Collect(storage.Metric{ID: "NextGC", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.NextGC))})
	a.h.Collect(storage.Metric{ID: "NumForcedGC", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.NumForcedGC))})
	a.h.Collect(storage.Metric{ID: "NumGC", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.NumGC))})
	a.h.Collect(storage.Metric{ID: "OtherSys", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.OtherSys))})
	a.h.Collect(storage.Metric{ID: "PauseTotalNs", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.PauseTotalNs))})
	a.h.Collect(storage.Metric{ID: "StackInuse", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.StackInuse))})
	a.h.Collect(storage.Metric{ID: "StackSys", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.StackSys))})
	a.h.Collect(storage.Metric{ID: "Sys", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.Sys))})
	a.h.Collect(storage.Metric{ID: "TotalAlloc", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.TotalAlloc))})
	a.h.Collect(storage.Metric{ID: "RandomValue", MType: storage.Gauge, Value: PtrFloat64(float64(rand.Int()))})
	a.h.Collect(storage.Metric{ID: "LastGC", MType: storage.Gauge, Value: PtrFloat64(float64(metrics.LastGC))})

	cnt, _ := storage.MetricStorage.GetMetric("PollCount")
	counter := int64(0)
	if cnt.Delta != nil {
		counter = *cnt.Delta + 1
	}
	storage.MetricStorage.Collect(storage.Metric{ID: "PollCount", MType: storage.Counter, Delta: PtrInt64(counter)})
}

func (a *Harvest) HarvestGopsutil() {
	v, _ := mem.VirtualMemory()
	cp, _ := cpu.Percent(0, false)
	a.h.Collect(storage.Metric{ID: "FreeMemory", MType: "gauge", Value: PtrFloat64(float64(v.Free))})
	a.h.Collect(storage.Metric{ID: "TotalMemory", MType: "gauge", Value: PtrFloat64(float64(v.Total))})
	a.h.Collect(storage.Metric{ID: "CPUutilization1", MType: "gauge", Value: PtrFloat64(cp[0])})
}

func PtrFloat64(f float64) *float64 {
	return &f
}

func PtrInt64(i int64) *int64 {
	return &i
}

type Sender struct {
	client        *resty.Client
	reportTimeout time.Duration
	addr          string
	log           zap.SugaredLogger
	rateLimit     int
}

func InitSender(opts *config.Options, log zap.SugaredLogger) *Sender {
	s := &Sender{
		client:        resty.New(),
		reportTimeout: time.Duration(opts.PollInterval),
		addr:          opts.FlagRunAddr,
		log:           log,
		rateLimit:     opts.RateLimit,
	}
	return s
}

func (s *Sender) SendMetrics(ctx context.Context) error {
	numRequests := make(chan struct{}, s.rateLimit)
	reportTicker := time.NewTicker(time.Duration(s.reportTimeout) * time.Second)
	for {
		select {
		case <-ctx.Done():
			return nil
		// check if time to send metrics on server
		case <-reportTicker.C:
			select {
			case <-ctx.Done():
				return nil
			// check if the rate limit is exceeded
			case numRequests <- struct{}{}:
				go func() {
					defer func() { <-numRequests }()
					s.log.Info("metrics sent on server")
					if err := s.sendMetricsToServer(); err != nil {
						log.Printf("error while sending metrics: %v", err)
					}
				}()
			default:
				s.log.Info("rate limit is exceeded")
			}
		}
	}
}

func (s *Sender) sendMetricsToServer() error {
	req := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Encoding", "gzip")

	for {
		for _, v := range storage.MetricStorage.Metrics {
			jsonInput, _ := json.Marshal(v)
			if err := s.sendRequest(req, string(jsonInput)); err != nil {
				return fmt.Errorf("error while sending agent request for counter metric: %w", err)
			}
		}
		time.Sleep(s.reportTimeout * time.Second)
	}
}

func (s *Sender) sendRequest(req *resty.Request, jsonInput string) error {
	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	if _, err := zb.Write([]byte(jsonInput)); err != nil {
		return fmt.Errorf("error while write json input: %w", err)
	}
	if err := zb.Close(); err != nil {
		return fmt.Errorf("error while trying to close writer: %w", err)
	}

	err := retry.Do(
		func() error {
			var err error
			if _, err = req.SetBody(buf).Post(fmt.Sprintf("http://%s/update/", s.addr)); err != nil {
				return fmt.Errorf("error while trying to create post request: %w", err)
			}
			return nil
		},
		retry.Attempts(10),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Retrying request after error: %v", err)
		}),
	)
	if err != nil {
		return fmt.Errorf("error while trying to connect to server: %w", err)
	}
	return nil
}
