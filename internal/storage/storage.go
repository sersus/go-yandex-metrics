package storage

import (
	"bufio"
	"encoding/json"
	"os"
	"time"

	"github.com/sersus/go-yandex-metrics/internal/middleware"
)

type MetricKind int

const (
	Counter = "counter"
	Gauge   = "gauge"
)

type Metric struct {
	Value      any
	MetricType string
}

type MemStorage struct {
	Metrics map[string]Metric
}

var MetricsStorage = MemStorage{Metrics: make(map[string]Metric)}

func SaveByInterval(metrics *MemStorage, filePath string, interval int) {
	if filePath != "" {
		for {
			err := metrics.save(filePath)
			if err != nil {
				middleware.SugarLogger.Error(err.Error(), "save error")
			} else {
				middleware.SugarLogger.Info("successfully saved metrics")
			}
			time.Sleep(time.Duration(interval) * time.Second)
		}
	}
}

func (metrics MemStorage) save(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	data, err := json.Marshal(&metrics)
	if err != nil {
		return err
	}
	if _, err := writer.Write(data); err != nil {
		return err
	}
	if err := writer.WriteByte('\n'); err != nil {
		return err
	}
	if err := writer.Flush(); err != nil {
		return err
	}
	return nil
}

func (metrics MemStorage) Restore(filePath string) error {
	file, err := os.OpenFile(filePath, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return err
	}
	data := scanner.Bytes()

	if err = json.Unmarshal(data, &metrics); err != nil {
		return err
	}
	return nil
}
