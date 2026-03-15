package repository

import (
	"context"
	"database/sql"
	"embed"
)

//go:embed sql/*.sql
var sqlFS embed.FS

func mustSQL(path string) string {
	b, err := sqlFS.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return string(b)
}

var (
	updateCounterSQL  = mustSQL("sql/updatecounter.sql")
	updateGaugeSQL    = mustSQL("sql/updategauge.sql")
	getCounterSQL     = mustSQL("sql/getcounter.sql")
	getGaugeSQL       = mustSQL("sql/getgauge.sql")
	getAllCountersSQL = mustSQL("sql/getallcounters.sql")
	getAllGaugesSQL   = mustSQL("sql/getallgauges.sql")
)

type PgStorage struct {
	db *sql.DB
}

func NewPgStorage(db *sql.DB) *PgStorage {
	return &PgStorage{db: db}
}

func (ps *PgStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	stmt, err := ps.db.PrepareContext(ctx, updateCounterSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, name, value)
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
	stmt, err := ps.db.PrepareContext(ctx, updateGaugeSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	res, err := stmt.ExecContext(ctx, name, value)
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

func (ps *PgStorage) updateCounters(ctx context.Context, tx *sql.Tx, counters map[string]int64) error {
	stmt, err := tx.PrepareContext(ctx, updateCounterSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for counter, value := range counters {
		_, err := stmt.ExecContext(ctx, counter, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ps *PgStorage) updateGauges(ctx context.Context, tx *sql.Tx, gauges map[string]float64) error {
	stmt, err := tx.PrepareContext(ctx, updateGaugeSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for gauge, value := range gauges {
		_, err := stmt.ExecContext(ctx, gauge, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func (ps *PgStorage) UpdateMetrics(ctx context.Context, counters map[string]int64, gauges map[string]float64) error {
	tx, err := ps.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if len(counters) > 0 {
		err := ps.updateCounters(ctx, tx, counters)
		if err != nil {
			return err
		}
	}

	if len(gauges) > 0 {
		err := ps.updateGauges(ctx, tx, gauges)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (ps *PgStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	var value int64
	err := ps.db.QueryRowContext(ctx, getCounterSQL, name).Scan(&value)
	if err != nil {
		return 0, err
	}
	return value, nil

}

func (ps *PgStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	var value float64
	err := ps.db.QueryRowContext(ctx, getGaugeSQL, name).Scan(&value)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func (ps *PgStorage) GetAllGauges(ctx context.Context) (map[string]float64, error) {
	rows, err := ps.db.QueryContext(ctx, getAllGaugesSQL)
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
	rows, err := ps.db.QueryContext(ctx, getAllCountersSQL)
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
