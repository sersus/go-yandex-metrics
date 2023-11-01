package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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
			log.Printf("metric name: %s new value : %s", metricName, metricValue)
			log.Println("old value : ", storage.MetricsStorage.Metrics[metricName].Value)
			switch storage.MetricsStorage.Metrics[metricName].Value.(type) {
			case int:
				value += storage.MetricsStorage.Metrics[metricName].Value.(int)
			case float64:
				value += int(storage.MetricsStorage.Metrics[metricName].Value.(float64))
			}
		}
		storage.MetricsStorage.Metrics[metricName] = storage.Metric{Value: value, MetricType: metricType}

	case storage.Gauge:
		_, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			log.Printf("Bad request: %s", metricName)
			return errBadRequest
		}
		log.Printf("Good request: %s", metricName)
		storage.MetricsStorage.Metrics[metricName] = storage.Metric{Value: metricValue, MetricType: metricType}
	default:
		return errNotImplemented
	}
	return nil
}

func findMetricByName(metricName string, metricType string) (string, error) {
	log.Printf("metricName: %s", metricName)
	log.Printf("metricType: %s", metricType)
	switch metricType {
	case storage.Counter:
		metric, ok := storage.MetricsStorage.Metrics[metricName]
		if !ok {
			return "", errNotFound
		}
		value := 0
		switch storage.MetricsStorage.Metrics[metricName].Value.(type) {
		case int:
			value = metric.Value.(int)
		case float64:
			value = int(metric.Value.(float64))
		}
		return strconv.Itoa(value), nil
	case storage.Gauge:
		metric, ok := storage.MetricsStorage.Metrics[metricName]
		if !ok {
			return "", errNotFound
		}
		log.Printf("metric.Value.(string): %s", metric.Value.(string))
		return metric.Value.(string), nil
	default:
		return "", errNotImplemented
	}
}

func (h *MetricsHandler) SendMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("SendMetric Method not allowed: %s", r.Method)
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
	err := findMetric(url[3], url[2], url[4])
	if errors.Is(err, errNotImplemented) {
		w.WriteHeader(http.StatusNotImplemented)
	}
	if errors.Is(err, errBadRequest) {
		w.WriteHeader(http.StatusBadRequest)
	}

	w.WriteHeader(http.StatusOK)
	if _, err = io.WriteString(w, ""); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.Header().Set("content-length", strconv.Itoa(len(url[3])))
}

func (h *MetricsHandler) GetMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("GetMetric Method not allowed: %s", r.Method)
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

	metric := storage.MetricsStorage.Metrics[url[3]]

	io.WriteString(w, "")
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.Header().Set("content-length", strconv.Itoa(len(url[3])))
	w.WriteHeader(http.StatusOK)

	switch metric.MetricType {
	case storage.Counter:
		io.WriteString(w, strconv.Itoa(metric.Value.(int)))
	default:
		io.WriteString(w, metric.Value.(string))
	}
}

func (h *MetricsHandler) ShowMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "Content-Type: text/html; charset=utf-8")
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
	//w.WriteHeader(http.StatusOK)
	w.Write([]byte(page))
}

func (h *MetricsHandler) SaveMetricFromJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("SaveMetricFromJSON Method not allowed: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		log.Println("Bad request 1")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metric Metrics
	if err := json.Unmarshal(buf.Bytes(), &metric); err != nil {
		log.Println("Bad request 2")
		log.Printf("%s", buf.String())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := collectFromJSON(metric)
	if errors.Is(err, errBadRequest) {
		log.Println("Bad request 3")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if errors.Is(err, errNotImplemented) {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	resultJSON, err := getMetricJSON(metric.ID, metric.MType)
	if errors.Is(err, errBadRequest) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if errors.Is(err, errNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if errors.Is(err, errNotImplemented) {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	if _, err = w.Write(resultJSON); err != nil {
		return
	}
	w.Header().Set("content-length", strconv.Itoa(len(metric.ID)))
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *MetricsHandler) GetMetricFromJSON(w http.ResponseWriter, r *http.Request) {
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

	resultJSON, err := getMetricJSON(metric.ID, metric.MType)
	if errors.Is(err, errBadRequest) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if errors.Is(err, errNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if errors.Is(err, errNotImplemented) {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	log.Printf("%s", resultJSON)
	w.Header().Set("content-type", "application/json")
	if _, err = w.Write(resultJSON); err != nil {
		return
	}
	w.Header().Set("content-length", strconv.Itoa(len(metric.ID)))
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func collectFromJSON(metric Metrics) error {
	metricValue := ""
	switch metric.MType {
	case storage.Counter:
		metricValue = strconv.Itoa(int(*metric.Delta))
	case storage.Gauge:
		metricValue = fmt.Sprintf("%.11f", *metric.Value)
	}

	return findMetric(metric.ID, metric.MType, metricValue)
}

func getMetricJSON(metricName string, metricType string) ([]byte, error) {
	updated, err := findMetricByName(metricName, metricType)
	if err != nil {
		return nil, err
	}

	result := Metrics{
		ID:    metricName,
		MType: metricType,
	}
	switch result.MType {
	case storage.Counter:
		counter, err := strconv.Atoi(updated)
		if err != nil {
			return nil, errBadRequest
		}
		c64 := int64(counter)
		result.Delta = &c64
	case storage.Gauge:
		g, err := strconv.ParseFloat(updated, 64)
		if err != nil {
			return nil, errBadRequest
		}
		result.Value = &g
	}

	resultJSON, err := json.Marshal(result)
	if err != nil {
		return nil, errBadRequest
	}
	return resultJSON, nil
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
