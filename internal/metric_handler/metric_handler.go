package metric_handler

import (
	"math/rand"
	"runtime"

	"github.com/sersus/go-yandex-metrics/internal/storage"
)

func (mh *metric_handler) Collect(metrics *runtime.MemStats) {
	mh.storage.Metrics["Alloc"] = storage.Metric{Value: metrics.Alloc, MetricType: storage.Gauge}
	mh.storage.Metrics["BuckHashSys"] = storage.Metric{Value: metrics.BuckHashSys, MetricType: storage.Gauge}
	mh.storage.Metrics["Frees"] = storage.Metric{Value: metrics.Frees, MetricType: storage.Gauge}
	mh.storage.Metrics["GCCPUFraction"] = storage.Metric{Value: metrics.GCCPUFraction, MetricType: storage.Gauge}
	mh.storage.Metrics["GCSys"] = storage.Metric{Value: metrics.GCSys, MetricType: storage.Gauge}
	mh.storage.Metrics["HeapAlloc"] = storage.Metric{Value: metrics.HeapAlloc, MetricType: storage.Gauge}
	mh.storage.Metrics["HeapIdle"] = storage.Metric{Value: metrics.HeapIdle, MetricType: storage.Gauge}
	mh.storage.Metrics["HeapInuse"] = storage.Metric{Value: metrics.HeapInuse, MetricType: storage.Gauge}
	mh.storage.Metrics["HeapObjects"] = storage.Metric{Value: metrics.HeapObjects, MetricType: storage.Gauge}
	mh.storage.Metrics["HeapReleased"] = storage.Metric{Value: metrics.HeapReleased, MetricType: storage.Gauge}
	mh.storage.Metrics["HeapSys"] = storage.Metric{Value: metrics.HeapSys, MetricType: storage.Gauge}
	mh.storage.Metrics["Lookups"] = storage.Metric{Value: metrics.Lookups, MetricType: storage.Gauge}
	mh.storage.Metrics["MCacheInuse"] = storage.Metric{Value: metrics.MCacheInuse, MetricType: storage.Gauge}
	mh.storage.Metrics["MCacheSys"] = storage.Metric{Value: metrics.MCacheSys, MetricType: storage.Gauge}
	mh.storage.Metrics["MSpanInuse"] = storage.Metric{Value: metrics.MSpanInuse, MetricType: storage.Gauge}
	mh.storage.Metrics["MSpanSys"] = storage.Metric{Value: metrics.MSpanSys, MetricType: storage.Gauge}
	mh.storage.Metrics["Mallocs"] = storage.Metric{Value: metrics.Mallocs, MetricType: storage.Gauge}
	mh.storage.Metrics["NextGC"] = storage.Metric{Value: metrics.NextGC, MetricType: storage.Gauge}
	mh.storage.Metrics["NumForcedGC"] = storage.Metric{Value: metrics.NumForcedGC, MetricType: storage.Gauge}
	mh.storage.Metrics["NumGC"] = storage.Metric{Value: metrics.NumGC, MetricType: storage.Gauge}
	mh.storage.Metrics["OtherSys"] = storage.Metric{Value: metrics.OtherSys, MetricType: storage.Gauge}
	mh.storage.Metrics["PauseTotalNs"] = storage.Metric{Value: metrics.PauseTotalNs, MetricType: storage.Gauge}
	mh.storage.Metrics["StackInuse"] = storage.Metric{Value: metrics.StackInuse, MetricType: storage.Gauge}
	mh.storage.Metrics["StackSys"] = storage.Metric{Value: metrics.StackSys, MetricType: storage.Gauge}
	mh.storage.Metrics["Sys"] = storage.Metric{Value: metrics.Sys, MetricType: storage.Gauge}
	mh.storage.Metrics["TotalAlloc"] = storage.Metric{Value: metrics.TotalAlloc, MetricType: storage.Gauge}
	mh.storage.Metrics["RandomValue"] = storage.Metric{Value: rand.Int(), MetricType: storage.Gauge}

	var cnt int64
	if mh.storage.Metrics["PollCount"].Value != nil {
		cnt = mh.storage.Metrics["PollCount"].Value.(int64) + 1
	}
	mh.storage.Metrics["PollCount"] = storage.Metric{Value: cnt, MetricType: storage.Counter}
}

func New(ms *storage.MemStorage) *metric_handler {
	return &metric_handler{ms}
}

type metric_handler struct {
	storage *storage.MemStorage
}
