package metric_handler

import (
	"runtime"
	"testing"

	"github.com/sersus/go-yandex-metrics/internal/storage"
	"github.com/stretchr/testify/assert"
)

func Test_metric_handler_Collect(t *testing.T) {
	testCases := []struct {
		name     string
		storage  storage.MemStorage
		metric   runtime.MemStats
		expected storage.MemStorage
	}{
		{
			name:    "case0",
			storage: storage.MemStorage{Metrics: map[string]storage.Metric{}},
			metric:  runtime.MemStats{Alloc: 1, Sys: 1, GCCPUFraction: 42.42},
			expected: storage.MemStorage{Metrics: map[string]storage.Metric{
				"Alloc":         {MetricType: storage.Gauge, Value: uint64(1)},
				"Sys":           {MetricType: storage.Gauge, Value: uint64(1)},
				"GCCPUFraction": {MetricType: storage.Gauge, Value: 42.42},
			}},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			metric := runtime.MemStats{Alloc: 1, Sys: 1, GCCPUFraction: 42.42}
			metricsmetric_handler := New(&tt.storage)
			metricsmetric_handler.Collect(&metric)
			assert.Equal(t, tt.expected.Metrics["Alloc"], tt.storage.Metrics["Alloc"])
			assert.Equal(t, tt.expected.Metrics["Sys"], tt.storage.Metrics["Sys"])
			assert.Equal(t, tt.expected.Metrics["GCCPUFraction"], tt.storage.Metrics["GCCPUFraction"])
		})
	}
}
