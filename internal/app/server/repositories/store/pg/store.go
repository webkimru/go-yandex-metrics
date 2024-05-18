// Package pg обеспечивает хранение данных в PostgreSQL.
package pg

import (
	"context"
	"database/sql"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/config"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/logger"
	"github.com/webkimru/go-yandex-metrics/internal/app/server/models"
)

var DB *Store

// Store реализует интерфейс store.StoreRepository и позволяет взаимодействовать с СУБД PostgreSQL.
type Store struct {
	// Поле conn содержит объект соединения с СУБД.
	Conn *sql.DB
}

// NewStore возвращает новый экземпляр PostgreSQL-хранилища.
func NewStore(conn *sql.DB) *Store {
	return &Store{Conn: conn}
}

// UpdateCounter обновляет поле Counter с использованием конструкции INSERT INTO ... ON CONFLICT ... UPDATE.
func (s *Store) UpdateCounter(ctx context.Context, name string, value int64) (int64, error) {
	stmt, err := s.Conn.PrepareContext(ctx, `
		INSERT INTO metrics.counters (name, delta) VALUES($1, $2)
			ON CONFLICT (name) DO
		    	UPDATE SET delta = metrics.counters.delta + $2 RETURNING delta
	`)
	if err != nil {
		return 0, err
	}

	var res int64
	err = stmt.QueryRowContext(ctx, name, value).Scan(&res)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	return res, nil
}

// UpdateGauge обновляет поле Gauge с использованием конструкции INSERT INTO ... ON CONFLICT ... UPDATE.
func (s *Store) UpdateGauge(ctx context.Context, name string, value float64) (float64, error) {
	stmt, err := s.Conn.PrepareContext(ctx, `
		INSERT INTO metrics.gauges (name, value) VALUES($1, $2)
			ON CONFLICT (name) DO
		    	UPDATE SET value = $2 RETURNING value
	`)
	if err != nil {
		return 0, err
	}

	var res float64
	err = stmt.QueryRowContext(ctx, name, value).Scan(&res)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	return res, nil
}

// GetCounter возращает значение счетчика Counter.
func (s *Store) GetCounter(ctx context.Context, metric string) (int64, error) {
	stmt, err := s.Conn.PrepareContext(ctx, `
		SELECT delta FROM metrics.counters
		WHERE name = $1
	`)
	if err != nil {
		return 0, err
	}

	var res int64
	err = stmt.QueryRowContext(ctx, metric).Scan(&res)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	return res, nil
}

// GetGauge возращает значение счетчика Gauge.
func (s *Store) GetGauge(ctx context.Context, metric string) (float64, error) {
	stmt, err := s.Conn.PrepareContext(ctx, `
		SELECT value FROM metrics.gauges
		WHERE name = $1
	`)
	if err != nil {
		return 0, err
	}

	var res float64
	err = stmt.QueryRowContext(ctx, metric).Scan(&res)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	return res, nil
}

// GetAllMetrics возращает мапку счетчиков Counter и Gauge.
func (s *Store) GetAllMetrics(ctx context.Context) (map[string]interface{}, error) {
	all := make(map[string]interface{}, 30)

	// gauge
	gauge, err := s.GetGaugeMetrics(ctx)
	if err != nil {
		return nil, err
	}
	all["gauge"] = gauge

	// counter
	counter, err := s.GetCounterMetrics(ctx)
	if err != nil {
		return nil, err
	}
	all["counter"] = counter

	return all, nil
}

// GetGaugeMetrics возращает все метрики Gauge и их значения в виде мапки
func (s *Store) GetGaugeMetrics(ctx context.Context) (map[string]float64, error) {
	// По умолчанию до 30 метрик данного типа.
	gauges := make(map[string]float64, 30)

	stmt, err := s.Conn.PrepareContext(ctx, `SELECT name, value FROM metrics.gauges`)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	// Не забываем закрыть курсор после завершения работы с данными.
	defer rows.Close()

	// Считываем записи.
	for rows.Next() {
		var idx string
		var value float64
		err = rows.Scan(&idx, &value)
		if err != nil {
			return nil, err
		}
		gauges[idx] = value
	}

	// Необходимо проверить ошибки уровня курсора.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return gauges, nil
}

// GetCounterMetrics возращает все метрики Counter и их значения в виде мапки
func (s *Store) GetCounterMetrics(ctx context.Context) (map[string]int64, error) {
	// По умолчанию 1 метрика данного типа.
	counters := make(map[string]int64, 1)

	stmt, err := s.Conn.PrepareContext(ctx, `SELECT name, delta FROM metrics.counters`)
	if err != nil {
		return nil, err
	}

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	// Не забываем закрыть курсор после завершения работы с данными.
	defer rows.Close()

	// Считываем записи.
	for rows.Next() {
		var idx string
		var delta int64
		err = rows.Scan(&idx, &delta)
		if err != nil {
			return nil, err
		}
		counters[idx] = delta
	}

	// Необходимо проверить ошибки уровня курсора.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return counters, nil
}

// UpdateBatchMetrics обновляет значение метрик Gauge и Counter по входящему батчу.
func (s *Store) UpdateBatchMetrics(ctx context.Context, metrics []models.Metrics) error {
	tx, err := s.Conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for i := range metrics {
		switch metrics[i].MType {
		case "gauge":
			_, err = tx.ExecContext(ctx, `
				INSERT INTO metrics.gauges (name, value) VALUES($1, $2)
					ON CONFLICT (name) DO
						UPDATE SET value = $2 RETURNING value
			`, metrics[i].ID, metrics[i].Value)
			if err != nil {
				logger.Log.Errorln(err)
			}

		case "counter":
			_, err = tx.ExecContext(ctx, `
				INSERT INTO metrics.counters (name, delta) VALUES($1, $2)
					ON CONFLICT (name) DO
						UPDATE SET delta = metrics.counters.delta + $2 RETURNING delta
			`, metrics[i].ID, metrics[i].Delta)
			if err != nil {
				logger.Log.Errorln(err)
			}
		}
	}

	return tx.Commit()
}

func (s *Store) Initialize(ctx context.Context, app config.AppConfig) error {
	var err error
	if s.Conn, err = ConnectToDB(app.DatabaseDSN); err != nil {
		return err
	}

	if err = Bootstrap(ctx, s.Conn); err != nil {
		return err
	}

	// нужно для грейсфула, где требуется вызов pg.DB.Conn.Close()
	DB = s

	return nil
}
