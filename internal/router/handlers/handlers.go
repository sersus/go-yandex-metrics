package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/sersus/go-yandex-metrics/internal/harvester"
	"github.com/sersus/go-yandex-metrics/internal/storage"
)

type handler struct {
	dbAddress string
}

func New(db string) *handler {
	return &handler{
		dbAddress: db,
	}
}

func (h *handler) SaveMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if metricName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	metric := storage.Metric{
		ID:    metricName,
		MType: metricType,
	}
	switch metricType {
	case storage.Counter:
		v, err := strconv.Atoi(metricValue)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		metric.Delta = harvester.PtrInt64(int64(v))
	case storage.Gauge:
		v, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		metric.Value = &v
	}
	err := storage.MetricStorage.Collect(metric)
	if errors.Is(err, storage.ErrBadRequest) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if errors.Is(err, storage.ErrNotImplemented) {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = io.WriteString(w, fmt.Sprintf("inserted metric %q with value %q", metricName, metricValue)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.Header().Set("content-length", strconv.Itoa(len(metricName)))
}

func (h *handler) SaveMetricFromJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metric storage.Metric
	if err := json.Unmarshal(buf.Bytes(), &metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if metric.ID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := storage.MetricStorage.Collect(metric)
	if errors.Is(err, storage.ErrBadRequest) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if errors.Is(err, storage.ErrNotImplemented) {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	resultJSON, err := storage.MetricStorage.GetMetricJSON(metric.ID)
	if errors.Is(err, storage.ErrBadRequest) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if errors.Is(err, storage.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if errors.Is(err, storage.ErrNotImplemented) {
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

func (h *handler) SaveListMetricsFromJSON(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metrics []storage.Metric
	if err := json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var results []byte
	for _, metric := range metrics {
		if metric.ID == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		err := storage.MetricStorage.Collect(metric)
		if errors.Is(err, storage.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if errors.Is(err, storage.ErrNotImplemented) {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}

		resultJSON, err := storage.MetricStorage.GetMetricJSON(metric.ID)
		if errors.Is(err, storage.ErrBadRequest) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if errors.Is(err, storage.ErrNotFound) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if errors.Is(err, storage.ErrNotImplemented) {
			w.WriteHeader(http.StatusNotImplemented)
			return
		}
		results = append(results, resultJSON...)
	}
	if _, err := w.Write(results); err != nil {
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *handler) GetMetricFromJSON(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metric storage.Metric
	if err := json.Unmarshal(buf.Bytes(), &metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resultJSON, err := storage.MetricStorage.GetMetricJSON(metric.ID)
	if errors.Is(err, storage.ErrBadRequest) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if errors.Is(err, storage.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if errors.Is(err, storage.ErrNotImplemented) {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	w.Header().Set("content-type", "application/json")
	if _, err = w.Write(resultJSON); err != nil {
		return
	}
	w.Header().Set("content-length", strconv.Itoa(len(metric.ID)))
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
}

func (h *handler) GetMetric(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	if metricType != storage.Counter && metricType != storage.Gauge {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	value, err := storage.MetricStorage.GetMetric(metricName)
	if errors.Is(err, storage.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if errors.Is(err, storage.ErrNotImplemented) {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	w.WriteHeader(http.StatusOK)
	switch metricType {
	case storage.Counter:
		if _, err = io.WriteString(w, fmt.Sprintf("%d", *value.Delta)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	case storage.Gauge:
		//if _, err = io.WriteString(w, fmt.Sprintf("%.3f", *value.Value)); err != nil {
		if _, err = io.WriteString(w, fmt.Sprintf("%g", *value.Value)); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.Header().Set("content-type", "text/plain; charset=utf-8")
}

func (h *handler) ShowMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "Content-Type: text/html; charset=utf-8")
	if r.URL.Path != "/" {
		http.Error(w, fmt.Sprintf("wrong path %q", r.URL.Path), http.StatusNotFound)
		return
	}
	page := ""
	for _, n := range storage.MetricStorage.GetAvailableMetrics() {
		page += fmt.Sprintf("<h1>	%s</h1>", n)
	}
	tmpl, _ := template.New("data").Parse("<h1>AVAILABLE METRICS</h1>{{range .}}<h3>{{ .}}</h3>{{end}}")
	if err := tmpl.Execute(w, storage.MetricStorage.GetAvailableMetrics()); err != nil {
		return
	}
	w.Header().Set("content-type", "Content-Type: text/html; charset=utf-8")
}

func (h *handler) Ping(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", h.dbAddress)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	_, err = w.Write([]byte("pong"))
	if err != nil {
		return
	}
}
