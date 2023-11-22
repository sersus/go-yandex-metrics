package storager

import (
	"context"
	"time"

	"github.com/sersus/go-yandex-metrics/internal/config"
	"github.com/sersus/go-yandex-metrics/internal/middleware"
	"github.com/sersus/go-yandex-metrics/internal/storage"
)

type saver interface {
	Restore(ctx context.Context) ([]storage.Metric, error)
	Save(ctx context.Context, metrics []storage.Metric) error
}

func InitSaver(params *config.Options, ctx context.Context) saver {
	var saver saver
	var err error
	if params.FileStoragePath != "" && params.DatabaseAddress == "" {
		saver = NewFilesaver(params, ctx)
	} else if params.DatabaseAddress != "" {
		saver, err = NewDBSaver(params, ctx)
		if err != nil {
			middleware.SugarLogger.Errorf(err.Error())
		}
	}
	return saver
}

type SaverHelper struct {
	saver    saver
	ctx      context.Context
	interval time.Duration
}

func InitSaverHelper(opts *config.Options) *SaverHelper {
	ctx := context.Background()
	sh := &SaverHelper{
		saver:    InitSaver(opts, ctx),
		ctx:      ctx,
		interval: time.Duration(opts.StoreInterval),
	}
	return sh
}

func (sh *SaverHelper) SaveMetrics() {
	ticker := time.NewTicker(sh.interval)
	for {
		select {
		case <-sh.ctx.Done():
			return
		case <-ticker.C:
			if err := sh.saver.Save(sh.ctx, storage.MetricStorage.Metrics); err != nil {
				middleware.SugarLogger.Error(err.Error(), "save error")
			}
		}
	}
}
