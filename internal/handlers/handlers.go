package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/sersus/go-yandex-metrics/internal/storage"
)

var (
	errBadRequest     = errors.New("bad request")
	errNotImplemented = errors.New("not implemented")
	errNotFound       = errors.New("not found")
)

type MetricsHandler struct{}

func findMetric(metricName string, metricType string, metricValue string) error {
	switch metricType {
	case storage.Counter:
		value, err := strconv.Atoi(metricValue)
		if err != nil {
			return errBadRequest
		}

		if storage.MetricsStorage.Metrics[metricName].Value != nil {
			value += storage.MetricsStorage.Metrics[metricName].Value.(int)
		}
		storage.MetricsStorage.Metrics[metricName] = storage.Metric{Value: value, MetricType: metricType}
	case storage.Gauge:
		_, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return errBadRequest
		}
		storage.MetricsStorage.Metrics[metricName] = storage.Metric{Value: metricValue, MetricType: metricType}
	default:
		return errNotImplemented
	}
	return nil
}

func findMetricByName(metricName string, metricType string) (string, error) {
	switch metricType {
	case storage.Counter:
		metric, ok := storage.MetricsStorage.Metrics[metricName]
		if !ok {
			return "", errNotFound
		}
		return strconv.Itoa(metric.Value.(int)), nil
	case storage.Gauge:
		metric, ok := storage.MetricsStorage.Metrics[metricName]
		if !ok {
			return "", errNotFound
		}
		return metric.Value.(string), nil
	default:
		return "", errNotImplemented
	}
}

func (h *MetricsHandler) SendMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if len(storage.MetricsStorage.Metrics) == 0 {
		storage.MetricsStorage.Metrics = make(map[string]storage.Metric, 0)
	}

	url := strings.Split(r.URL.String(), "/")
	if len(url) != 5 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if url[1] != "update" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	metricType := url[2]
	metricName := url[3]
	err := findMetric(url[3], url[2], url[4])
	if errors.Is(err, errNotImplemented) {
		w.WriteHeader(http.StatusNotImplemented)
	}
	if errors.Is(err, errBadRequest) {
		w.WriteHeader(http.StatusBadRequest)
	}
	switch metricType {
	case storage.Counter:
		value, err := strconv.Atoi(url[4])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if storage.MetricsStorage.Metrics[metricName].Value != nil {
			value += storage.MetricsStorage.Metrics[metricName].Value.(int)
		}
		storage.MetricsStorage.Metrics[metricName] = storage.Metric{Value: value, MetricType: metricType}
	case storage.Gauge:
		_, err := strconv.ParseFloat(url[4], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		storage.MetricsStorage.Metrics[metricName] = storage.Metric{Value: url[4], MetricType: metricType}
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	io.WriteString(w, "")
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.Header().Set("content-length", strconv.Itoa(len(metricName)))
	w.WriteHeader(http.StatusOK)
}

func (h *MetricsHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	url := strings.Split(r.URL.String(), "/")
	if len(url) != 4 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if url[1] != "value" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if _, ok := storage.MetricsStorage.Metrics[url[3]]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	value := storage.MetricsStorage.Metrics[url[3]].Value

	io.WriteString(w, "")
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.Header().Set("content-length", strconv.Itoa(len(url[3])))
	w.WriteHeader(http.StatusOK)

	switch value.(type) {
	case uint, uint64, int, int64:
		io.WriteString(w, strconv.Itoa(value.(int)))
	default:
		io.WriteString(w, value.(string))
	}
}

func (h *MetricsHandler) ShowMetrics(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	page := `
<html> 
   <head> 
   </head> 
   <body> 
`
	for n := range storage.MetricsStorage.Metrics {
		page += fmt.Sprintf(`<h3>%s   </h3>`, n)
	}
	page += `
   </body> 
</html>
`
	w.Header().Set("content-type", "Content-Type: text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(page))
}

func (h *MetricsHandler) SaveMetricFromJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metric Metrics
	if err := json.Unmarshal(buf.Bytes(), &metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	metricValue := ""
	switch metric.MType {
	case "counter":
		metricValue = strconv.Itoa(int(*metric.Delta))
	case "gauge":
		metricValue = fmt.Sprintf("%.11f", *metric.Value)
	}

	err := findMetric(metric.ID, metric.MType, metricValue)
	if errors.Is(err, errBadRequest) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if errors.Is(err, errNotImplemented) {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	updated, err := findMetricByName(metric.ID, metric.MType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := Metrics{
		ID:    metric.ID,
		MType: metric.MType,
	}
	switch result.MType {
	case "counter":
		c, err := strconv.Atoi(updated)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		c64 := int64(c)
		result.Delta = &c64
	case "gauge":
		g, err := strconv.ParseFloat(updated, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		result.Value = &g
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "application/json")
	if _, err = w.Write(resultJSON); err != nil {
		return
	}
	w.Header().Set("content-length", strconv.Itoa(len(metric.ID)))
	w.WriteHeader(http.StatusOK)
}

func (h *MetricsHandler) GetMetricFromJSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metric Metrics
	if err := json.Unmarshal(buf.Bytes(), &metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	value, err := findMetricByName(metric.ID, metric.MType)
	if errors.Is(err, errNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if errors.Is(err, errNotImplemented) {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	if _, err = io.WriteString(w, ""); err != nil {
		return
	}
	switch metric.MType {
	case "counter":
		c, err := strconv.Atoi(value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		c64 := int64(c)
		metric.Delta = &c64
	case "gauge":
		g, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		metric.Value = &g
	}
	resultJSON, err := json.Marshal(metric)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "application/json")
	if _, err = w.Write(resultJSON); err != nil {
		return
	}
	w.Header().Set("content-length", strconv.Itoa(len(value)))
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
}

type MetricsHandlerInterface interface {
	SendMetric(w http.ResponseWriter, r *http.Request)
	GetMetric(w http.ResponseWriter, r *http.Request)
	GetMetricFromJSON(w http.ResponseWriter, r *http.Request)
	ShowMetrics(w http.ResponseWriter, r *http.Request)
	SaveMetricFromJSON(w http.ResponseWriter, r *http.Request)
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}
