package repository

import (
	"context"
	"database/sql"
)

type PgStorage struct {
	db *sql.DB
}

func NewPgStorage(db *sql.DB) *PgStorage {
	return &PgStorage{db: db}
}

func (ps *PgStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	query := `
		INSERT INTO metrics.counters (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET
		value = counters.value + EXCLUDED.value,
		updated_at = now()
	`
	res, err := ps.db.ExecContext(ctx, query, name, value)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (ps *PgStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	query := `
		INSERT INTO metrics.gauges (name, value)
		VALUES ($1, $2)
		ON CONFLICT (name)
		DO UPDATE SET value = EXCLUDED.value,
		updated_at = now()
	`
	res, err := ps.db.ExecContext(ctx, query, name, value)
	if err != nil {
		return err
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (ps *PgStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	query := `
		SELECT value
		FROM metrics.counters
		WHERE name=$1
	`
	var value int64
	err := ps.db.QueryRowContext(ctx, query, name).Scan(&value)
	if err != nil {
		return 0, err
	}
	return value, nil

}

func (ps *PgStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	query := `
		SELECT value
		FROM metrics.gauges
		WHERE name=$1
	`
	var value float64
	err := ps.db.QueryRowContext(ctx, query, name).Scan(&value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (ps *PgStorage) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	query := `
		SELECT name, value
		FROM metrics.gauges
	`
	rows, err := ps.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	gauges := make(map[string]float64)
	for rows.Next() {
		var name string
		var value float64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		gauges[name] = value
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return gauges, nil
}

func (ps *PgStorage) GetAllCounters(ctx context.Context) (map[string]int64, error) {
	query := `
		SELECT name, value
		FROM metrics.counters
	`
	rows, err := ps.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counters := make(map[string]int64)
	for rows.Next() {
		var name string
		var value int64
		if err := rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		counters[name] = value
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return counters, nil
}
