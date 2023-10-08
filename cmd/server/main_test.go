package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricHandler(t *testing.T) {
	storage := NewMemStorage()
	metricHandler := NewMetricHandler(storage)

	req := httptest.NewRequest("POST", "/update/gauge/TestMetric/42.0", nil)
	w := httptest.NewRecorder()

	metricHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, w.Code)
	}

	value := storage.metrics["TestMetric"].(float64)
	if value != 42.0 {
		t.Errorf("Expected metric value to be 42.0, but got %f", value)
	}
}

func TestInvalidMetricHandler(t *testing.T) {
	storage := NewMemStorage()
	metricHandler := NewMetricHandler(storage)

	req := httptest.NewRequest("POST", "/update/unknown/TestMetric/42.0", nil)
	w := httptest.NewRecorder()

	metricHandler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, but got %d", http.StatusBadRequest, w.Code)
	}

	if _, ok := storage.metrics["TestMetric"]; ok {
		t.Errorf("Expected metric TestMetric to not be added to the storage")
	}
}
