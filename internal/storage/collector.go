package storage

import (
	"encoding/json"
	"errors"
)

var (
	ErrBadRequest     = errors.New("bad request")
	ErrNotImplemented = errors.New("not implemented")
	ErrNotFound       = errors.New("not found")
)

var Harvester = harvester{
	Metrics: make([]Metric, 0),
}

func (h *harvester) Collect(metric Metric) error {
	if (metric.Delta != nil && *metric.Delta < 0) || (metric.Value != nil && *metric.Value < 0) {
		return ErrBadRequest
	}
	switch metric.MType {
	case Counter:
		v, err := h.GetMetric(metric.ID)
		if err != nil {
			if !errors.Is(err, ErrNotFound) {
				return err
			}
		}
		if v.Delta != nil {
			*metric.Delta += *v.Delta
		}
		h.UpsertMetric(metric)

	case Gauge:
		h.UpsertMetric(metric)
	default:
		return ErrNotImplemented
	}
	return nil
}

func (h *harvester) GetMetric(metricName string) (Metric, error) {
	for _, m := range h.Metrics {
		if m.ID == metricName {
			return m, nil
		}
	}
	return Metric{}, ErrNotFound
}

func (h *harvester) GetMetricJSON(metricName string) ([]byte, error) {
	for _, m := range h.Metrics {
		if m.ID == metricName {
			resultJSON, err := json.Marshal(m)
			if err != nil {
				return nil, ErrBadRequest
			}
			return resultJSON, nil
		}
	}
	return nil, ErrNotFound
}

func (h *harvester) GetAvailableMetrics() []string {
	names := make([]string, 0)
	for _, m := range h.Metrics {
		names = append(names, m.ID)
	}
	return names
}

func (h *harvester) UpsertMetric(metric Metric) {
	for i, m := range h.Metrics {
		if m.ID == metric.ID {
			h.Metrics[i] = metric
			return
		}
	}
	h.Metrics = append(h.Metrics, metric)
}
