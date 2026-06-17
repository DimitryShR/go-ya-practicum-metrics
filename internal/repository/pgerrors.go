package repository

import (
	"database/sql"
	"database/sql/driver"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

// isRetryablePostgresError проверяет, является ли ошибка PostgreSQL временной (connection exception).
// Возвращает true для ошибок соединения, которые стоит повторить.
func isRetryablePostgresError(err error) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, driver.ErrBadConn) || errors.Is(err, sql.ErrConnDone) {
		return true
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return isConnectionExceptionCode(pgErr.Code)
}

// isConnectionExceptionCode проверяет, является ли код ошибки PostgreSQL кодом исключения соединения (класс 08).
func isConnectionExceptionCode(code string) bool {
	if len(code) < 2 {
		return false
	}
	return code[:2] == pgerrcode.ConnectionException[:2]
}
