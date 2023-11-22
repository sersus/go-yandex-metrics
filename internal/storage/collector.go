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

var MetricStorage = MetricCollection{
	Metrics: make([]Metric, 0),
}

func (mc *MetricCollection) Collect(metric Metric) error {
	if (metric.Delta != nil && *metric.Delta < 0) || (metric.Value != nil && *metric.Value < 0) {
		return ErrBadRequest
	}
	switch metric.MType {
	case Counter:
		v, err := mc.GetMetric(metric.ID)
		if err != nil {
			if !errors.Is(err, ErrNotFound) {
				return err
			}
		}
		if v.Delta != nil {
			*metric.Delta += *v.Delta
		}
		mc.UpsertMetric(metric)

	case Gauge:
		mc.UpsertMetric(metric)
	default:
		return ErrNotImplemented
	}
	return nil
}

func (mc *MetricCollection) GetMetric(metricName string) (Metric, error) {
	for _, m := range mc.Metrics {
		if m.ID == metricName {
			return m, nil
		}
	}
	return Metric{}, ErrNotFound
}

func (mc *MetricCollection) GetMetricJSON(metricName string) ([]byte, error) {
	for _, m := range mc.Metrics {
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

func (mc *MetricCollection) GetAvailableMetrics() []string {
	names := make([]string, 0)
	for _, m := range mc.Metrics {
		names = append(names, m.ID)
	}
	return names
}

func (mc *MetricCollection) UpsertMetric(metric Metric) {
	for i, m := range mc.Metrics {
		if m.ID == metric.ID {
			mc.Metrics[i] = metric
			return
		}
	}
	mc.Metrics = append(mc.Metrics, metric)
}
