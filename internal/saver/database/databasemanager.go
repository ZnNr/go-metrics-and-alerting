// Package database предоставляет функционал для работы с базой данных в контексте метрик.
package database

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"go.uber.org/zap"
	"time"
)

// SQL-запросы
const (
	// Запрос для восстановления состояния метрик
	selectMetricsQuery = `select id, mtype, delta, mvalue from metrics`

	// Запросы для сохранения состояния метрик (Gauge и Counter)
	insertGaugeQuery   = `insert into metrics (id, mtype, mvalue) values ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET mvalue = EXCLUDED.mvalue;`
	insertCounterQuery = `insert into metrics (id, mtype, delta) values ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET delta = EXCLUDED.delta;`

	// Запрос для создания таблицы метрик
	createMetricsTableQuery = `create table if not exists metrics (id text primary key, mtype text, delta bigint, mvalue double precision)`
)

var log *zap.SugaredLogger

func initLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		fmt.Println("error while creating logger, exit")
		return
	}
	defer logger.Sync()
	log = logger.Sugar()
}

// Restore восстанавливает состояние метрик из базы данных.
func (m *Manager) Restore(ctx context.Context) ([]collector.StoredMetric, error) {
	rows, err := m.db.QueryContext(ctx, selectMetricsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	var metrics []collector.StoredMetric
	for rows.Next() {
		var (
			id          string
			mtype       string
			deltaFromDB sql.NullInt64
			valueFromDB sql.NullFloat64
		)
		if err = rows.Scan(&id, &mtype, &deltaFromDB, &valueFromDB); err != nil {
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
		metric := collector.StoredMetric{
			ID:           id,
			MType:        mtype,
			CounterValue: delta,
			GaugeValue:   mvalue,
		}
		metrics = append(metrics, metric)
	}
	return metrics, nil
}

// Save — метод сохранения состояния метрики в БД.
func (m *Manager) Save(ctx context.Context, metrics []collector.StoredMetric) error {
	for _, metric := range metrics {
		switch metric.MType {
		case collector.Gauge:
			if err := m.execWithRetries(ctx, insertGaugeQuery, metric.ID, metric.MType, &metric.GaugeValue); err != nil {
				return fmt.Errorf("error while executing insert counter query: %w", err)
			}
		case collector.Counter:
			if err := m.execWithRetries(ctx, insertCounterQuery, metric.ID, metric.MType, &metric.CounterValue); err != nil {
				return fmt.Errorf("error while executing insert counter query: %w", err)
			}
		}
	}
	return nil
}

func (m *Manager) execWithRetries(ctx context.Context, query string, args ...interface{}) error {
	var err error
	initLogger()
	log.Info("Starting execWithRetries")
	retries := []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}
	for _, t := range retries {
		log.Info("Executing query with parameters", zap.String("query", query), zap.Any("args", args))
		if _, err = m.db.ExecContext(ctx, query, args...); err == nil {
			log.Info("Query executed successfully")
			break
		} else {
			log.Error("Error executing query", zap.Error(err))
			time.Sleep(t)
		}
	}
	return err
}

// init — метод подготовки новой БД для хранения метрик.
func (m *Manager) init(ctx context.Context) error {
	if _, err := m.db.ExecContext(ctx, createMetricsTableQuery); err != nil {
		return fmt.Errorf("error while trying to create table: %w", err)
	}
	return nil
}

// New создает новый экземпляр менеджера базы данных.
func New(db *sql.DB) (*Manager, error) {
	ctx := context.Background()
	m := Manager{
		db: db,
	}
	if err := m.init(ctx); err != nil {
		return nil, err
	}
	return &m, nil
}

// Manager представляет менеджер базы данных для сохранения и восстановления метрик.
type Manager struct {
	db *sql.DB
}
