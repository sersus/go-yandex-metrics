package storage

import (
	"encoding/json"
	"testing"
)

func ptrFloat64(f float64) *float64 {
	return &f
}

func ptrInt64(i int64) *int64 {
	return &i
}

func TestMetricCollection_Collect(t *testing.T) {
	tests := []struct {
		name     string
		metric   Metric
		expected error
	}{
		{
			name:     "ValidCounter",
			metric:   Metric{ID: "counter1", MType: Counter, Delta: ptrInt64(5)},
			expected: nil,
		},
		{
			name:     "ValidGauge",
			metric:   Metric{ID: "gauge1", MType: Gauge, Value: ptrFloat64(10.5)},
			expected: nil,
		},
		{
			name:     "NegativeDelta",
			metric:   Metric{ID: "counter2", MType: Counter, Delta: ptrInt64(-2)},
			expected: ErrBadRequest,
		},
	}

	mc := MetricCollection{
		Metrics: make([]Metric, 0),
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := mc.Collect(test.metric)
			if err != test.expected {
				t.Errorf("Expected error: %v, got: %v", test.expected, err)
			}
		})
	}
}

func TestMetricCollection_GetMetric(t *testing.T) {
	mc := MetricCollection{
		Metrics: []Metric{
			{ID: "metric1", MType: Counter, Delta: ptrInt64(5)},
			{ID: "metric2", MType: Gauge, Value: ptrFloat64(10.5)},
		},
	}

	t.Run("ExistingMetric", func(t *testing.T) {
		metric, err := mc.GetMetric("metric1")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if metric.ID != "metric1" {
			t.Errorf("Expected ID: metric1, got: %s", metric.ID)
		}
	})

	t.Run("NonExistingMetric", func(t *testing.T) {
		_, err := mc.GetMetric("nonexistent")
		if err != ErrNotFound {
			t.Errorf("Expected ErrNotFound, got: %v", err)
		}
	})
}

func TestMetricCollection_GetMetricJSON(t *testing.T) {
	mc := MetricCollection{
		Metrics: []Metric{
			{ID: "metric1", MType: Counter, Delta: ptrInt64(5)},
			{ID: "metric2", MType: Gauge, Value: ptrFloat64(10.5)},
		},
	}

	t.Run("ExistingMetric", func(t *testing.T) {
		jsonData, err := mc.GetMetricJSON("metric1")
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		var metric Metric
		err = json.Unmarshal(jsonData, &metric)
		if err != nil {
			t.Errorf("Error unmarshaling JSON: %v", err)
		}
		if metric.ID != "metric1" {
			t.Errorf("Expected ID: metric1, got: %s", metric.ID)
		}
	})

	t.Run("NonExistingMetric", func(t *testing.T) {
		_, err := mc.GetMetricJSON("nonexistent")
		if err != ErrNotFound {
			t.Errorf("Expected ErrNotFound, got: %v", err)
		}
	})
}

func TestMetricCollection_GetAvailableMetrics(t *testing.T) {
	mc := MetricCollection{
		Metrics: []Metric{
			{ID: "metric1", MType: Counter, Delta: ptrInt64(5)},
			{ID: "metric2", MType: Gauge, Value: ptrFloat64(10.5)},
		},
	}

	expected := []string{"metric1", "metric2"}
	result := mc.GetAvailableMetrics()

	if len(expected) != len(result) {
		t.Errorf("Expected length: %d, got: %d", len(expected), len(result))
	}

	for i := range expected {
		if expected[i] != result[i] {
			t.Errorf("Expected metric: %s, got: %s", expected[i], result[i])
		}
	}
}
