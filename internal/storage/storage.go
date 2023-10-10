package storage

type MetricKind int

const (
	Counter = "counter"
	Gauge   = "gauge"
)

type Metric struct {
	Value      any
	MetricType string
}

type MemStorage struct {
	Metrics map[string]Metric
}

var MetricsStorage = MemStorage{Metrics: make(map[string]Metric)}
