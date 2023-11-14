package storage

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"os"
	"time"

	"github.com/sersus/go-yandex-metrics/internal/config"
	"github.com/sersus/go-yandex-metrics/internal/middleware"
)

type storeKind int

const (
	file storeKind = iota
	db
	mem
)

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
	kind    storeKind
}

func NewMetricsStorage(options *config.ServerOptions) (*MemStorage, error) {
	if options.FileStoragePath != "" && options.ConnectDB == "" {

	}
	return nil, nil
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

type Storekeeper interface {
	Save() error
	Restore() (*MemStorage, error)
}

type filestorage struct {
	path string
}

func newfilestorage(path string) *filestorage {
	return &filestorage{path: path}
}

func (s *filestorage) Restore() (*MemStorage, error) {
	file, err := os.OpenFile(s.path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, err
	}
	data := scanner.Bytes()
	var metrics MemStorage
	if err = json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}
	return &metrics, nil

}

type dbstorage struct {
	db *sql.DB
}

func newdbstorage(path string) *dbstorage {
	return nil
}

func (s *dbstorage) Restore() (*MemStorage, error) {
	file, err := os.OpenFile(s.path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, err
	}
	data := scanner.Bytes()
	var metrics MemStorage
	if err = json.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}
	return &metrics, nil

}
