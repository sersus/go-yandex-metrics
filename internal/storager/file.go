package storager

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"github.com/sersus/go-yandex-metrics/internal/config"
	"github.com/sersus/go-yandex-metrics/internal/middleware"
	"github.com/sersus/go-yandex-metrics/internal/storage"
)

type filesaver struct {
	fileName string
}

func (m *filesaver) Restore(ctx context.Context) ([]storage.Metric, error) {
	file, err := os.OpenFile(m.fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, err
	}

	data := scanner.Bytes()
	var metricsFromFile []storage.Metric
	if err = json.Unmarshal(data, &metricsFromFile); err != nil {
		return nil, err
	}
	return metricsFromFile, nil
}

func (m *filesaver) Save(ctx context.Context, metrics []storage.Metric) error {
	var saveError error
	file, err := os.OpenFile(m.fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			saveError = err
		}
	}()

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
	return saveError
}

func NewFilesaver(params *config.Options, ctx context.Context) *filesaver {
	fs := &filesaver{fileName: params.FileStoragePath}
	if params.Restore {
		metrics, err := fs.Restore(ctx)
		if err != nil {
			middleware.SugarLogger.Error(err.Error(), "restore from file error")
		}
		storage.Harvester.Metrics = metrics
		middleware.SugarLogger.Info("metrics restored from file")
	}

	return fs
}
