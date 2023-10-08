package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

// MetricRepository представляет интерфейс для хранения метрик.
type MetricRepository interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
}

// MemStorage представляет хранилище метрик.
type MemStorage struct {
	mu      sync.Mutex
	metrics map[string]interface{}
}

// NewMemStorage создает новое хранилище метрик.
func NewMemStorage() *MemStorage {
	return &MemStorage{
		metrics: make(map[string]interface{}),
	}
}

// UpdateGauge обновляет метрику типа gauge.
func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metrics[name] = value
}

// UpdateCounter обновляет метрику типа counter.
func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existingValue, ok := s.metrics[name]; ok {
		if oldValue, isInt := existingValue.(int64); isInt {
			s.metrics[name] = oldValue + value
		} else {
			// Возвращаем ошибку, если существует метрика другого типа с таким же именем.
			panic(fmt.Sprintf("Metric %s is not a counter", name))
		}
	} else {
		s.metrics[name] = value
	}
}

type MetricHandler struct {
	repo MetricRepository
}

func NewMetricHandler(repo MetricRepository) *MetricHandler {
	return &MetricHandler{repo}
}

func (h *MetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 5 {
		http.Error(w, "Invalid request", http.StatusNotFound)
		return
	}

	metricType := parts[2]
	metricName := parts[3]
	metricValueStr := parts[4]

	if metricName == "" {
		http.Error(w, "Metric name can't be empty", http.StatusNotFound)
		return
	}

	switch metricType {
	case "gauge":
		metricValue, err := strconv.ParseFloat(metricValueStr, 64)
		if err != nil {
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}
		h.repo.UpdateGauge(metricName, metricValue)
	case "counter":
		metricValue, err := strconv.ParseInt(metricValueStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid metric value", http.StatusBadRequest)
			return
		}
		h.repo.UpdateCounter(metricName, metricValue)
	default:
		http.Error(w, "Invalid metric type", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func main() {
	storage := NewMemStorage()
	metricHandler := NewMetricHandler(storage)

	http.Handle("/update/", metricHandler)

	http.ListenAndServe(":8080", nil)
}
