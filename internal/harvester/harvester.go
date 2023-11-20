package harvester

import (
	"math/rand"
	"runtime"

	"github.com/sersus/go-yandex-metrics/internal/storage"
)

type Harvest struct {
	h Harvester
}

type Harvester interface {
	Collect(json storage.Metric) error
}

func (a *Harvest) Harvest() {
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

	cnt, _ := storage.Collector.GetMetric("PollCount")
	counter := int64(0)
	if cnt.Delta != nil {
		counter = *cnt.Delta + 1
	}
	storage.Collector.Collect(storage.Metric{ID: "PollCount", MType: storage.Counter, Delta: PtrInt64(counter)})
}

func New(harvester Harvester) *Harvest {
	return &Harvest{
		h: harvester,
	}
}

func PtrFloat64(f float64) *float64 {
	return &f
}

func PtrInt64(i int64) *int64 {
	return &i
}
