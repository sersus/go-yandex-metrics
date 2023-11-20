package storager

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/sersus/go-yandex-metrics/internal/config"
	"github.com/sersus/go-yandex-metrics/internal/storage"
)

type dbsaver struct {
	db *sql.DB
}

func (m *dbsaver) Restore(ctx context.Context) ([]storage.Metric, error) {
	const query = `select id, mtype, delta, mvalue from metrics`
	rows, err := m.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []storage.Metric
	for rows.Next() {
		var (
			id          string
			mtype       string
			deltaFromDB sql.NullInt64
			valueFromDB sql.NullFloat64
		)
		if err := rows.Scan(&id, &mtype, &deltaFromDB, &valueFromDB); err != nil {
			return nil, err
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
	return metrics, nil
}

func (m *dbsaver) Save(ctx context.Context, metrics []storage.Metric) error {
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

func (m *dbsaver) init(ctx context.Context) error {
	const query = `create table if not exists metrics (id text primary key, mtype text, delta bigint, mvalue double precision)`
	if _, err := m.db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("error while trying to create table: %w", err)
	}
	return nil
}

func NewDBSaver(params *config.Options) (*dbsaver, error) {
	ctx := context.Background()
	db, err := sql.Open("pgx", params.DatabaseAddress)
	if err != nil {
		return nil, err
	}

	m := dbsaver{
		db: db,
	}
	if err := m.init(ctx); err != nil {
		return nil, err
	}
	return &m, nil
}
