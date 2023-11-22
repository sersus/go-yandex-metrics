package storager

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/avast/retry-go"
	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"

	"github.com/sersus/go-yandex-metrics/internal/config"
	"github.com/sersus/go-yandex-metrics/internal/middleware"
	"github.com/sersus/go-yandex-metrics/internal/storage"
)

type dbsaver struct {
	db *sql.DB
}

func isRetriableError(err error) bool {
	pqErr, ok := err.(*pq.Error)
	if !ok {
		return false
	}
	return pgerrcode.IsConnectionException(string(pqErr.Code))
}

func (m *dbsaver) Restore(ctx context.Context) ([]storage.Metric, error) {

	var metrics []storage.Metric
	restoreOperation := func() error {
		const query = `select id, mtype, delta, mvalue from metrics`
		rows, err := m.db.QueryContext(ctx, query)
		if err != nil {
			return err
		}
		defer rows.Close()

		//var metrics []storage.Metric
		for rows.Next() {
			var (
				id          string
				mtype       string
				deltaFromDB sql.NullInt64
				valueFromDB sql.NullFloat64
			)
			if err := rows.Scan(&id, &mtype, &deltaFromDB, &valueFromDB); err != nil {
				return err
			}
			var delta *int64
			if deltaFromDB.Valid {
				delta = &deltaFromDB.Int64
			}
			var mvalue *float64
			if valueFromDB.Valid {
				mvalue = &valueFromDB.Float64
			}
			metric := storage.Metric{
				ID:    id,
				MType: mtype,
				Delta: delta,
				Value: mvalue,
			}
			metrics = append(metrics, metric)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		return nil
	}
	opts := []retry.Option{
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(1 * time.Second),
		retry.MaxDelay(5 * time.Second),
		retry.Attempts(3),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Retry #%d, error: %s\n", n, err)
		}),
	}

	err := retry.Do(
		restoreOperation,
		append(opts, retry.RetryIf(func(err error) bool {
			return isRetriableError(err) // Проверяем, является ли ошибка retriable
		}))...)
	if err != nil {
		log.Fatal("Failed to restore metrics with retries:", err)
	}
	return metrics, nil
}

func (m *dbsaver) Save(ctx context.Context, metrics []storage.Metric) error {
	saveOperation := func() error {
		for _, metric := range metrics {
			switch metric.MType {
			case storage.Gauge:
				query := `insert into metrics (id, mtype, mvalue) values ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET mvalue = EXCLUDED.mvalue;`
				if _, err := m.db.ExecContext(ctx, query, metric.ID, metric.MType, metric.Value); err != nil {
					return fmt.Errorf("error while trying to save gauge metric %q: %w", metric.ID, err)
				}
			case storage.Counter:
				query := `insert into metrics (id, mtype, delta) values ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET delta = EXCLUDED.delta;`
				if _, err := m.db.ExecContext(ctx, query, metric.ID, metric.MType, metric.Delta); err != nil {
					return fmt.Errorf("error while trying to save counter metric %q: %w", metric.ID, err)
				}
			}
		}
		return nil
	}
	opts := []retry.Option{
		retry.DelayType(retry.BackOffDelay),
		retry.Delay(1 * time.Second),
		retry.MaxDelay(5 * time.Second),
		retry.Attempts(3),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Retry #%d, error: %s\n", n, err)
		}),
	}

	err := retry.Do(
		saveOperation,
		append(opts, retry.RetryIf(func(err error) bool {
			return isRetriableError(err) // Проверяем, является ли ошибка retriable
		}))...)
	if err != nil {
		log.Fatal("Failed to save metrics with retries:", err)
	}
	return nil
}

func (m *dbsaver) init(ctx context.Context) error {
	const query = `create table if not exists metrics (id text primary key, mtype text, delta bigint, mvalue double precision)`
	if _, err := m.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("error while trying to create table: %w", err)
	}
	return nil
}

func NewDBSaver(params *config.Options, ctx context.Context) (*dbsaver, error) {
	//ctx := context.Background()
	db, err := sql.Open("pgx", params.DatabaseAddress)
	if err != nil {
		return nil, err
	}

	dbs := dbsaver{
		db: db,
	}
	if err := dbs.init(ctx); err != nil {
		return nil, err
	}
	if params.Restore {
		metrics, err := dbs.Restore(ctx)
		if err != nil {
			middleware.SugarLogger.Error(err.Error(), "restore from database error")
		}
		storage.Harvester.Metrics = metrics
		middleware.SugarLogger.Info("metrics restored from database")
	}

	return &dbs, nil
}
