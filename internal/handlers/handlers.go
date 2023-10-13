package handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/sersus/go-yandex-metrics/internal/storage"
)

type MetricsHandler struct{}

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
	switch url[2] {
	case storage.Counter:
		value, err := strconv.Atoi(url[4])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if storage.MetricsStorage.Metrics[url[3]].Value != nil {
			value += storage.MetricsStorage.Metrics[url[3]].Value.(int)
		}
		storage.MetricsStorage.Metrics[url[3]] = storage.Metric{Value: value, MetricType: url[2]}
	case storage.Gauge:
		_, err := strconv.ParseFloat(url[4], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		storage.MetricsStorage.Metrics[url[3]] = storage.Metric{Value: url[4], MetricType: url[2]}
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	io.WriteString(w, "")
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.Header().Set("content-length", strconv.Itoa(len(url[3])))
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

type MetricsHandlerInterface interface {
	SendMetric(w http.ResponseWriter, r *http.Request)
	GetMetric(w http.ResponseWriter, r *http.Request)
	ShowMetrics(w http.ResponseWriter, r *http.Request)
}
