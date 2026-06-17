package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"
)

// PingHandler — обработчик для проверки соединения с базой данных.
type PingHandler struct {
	db *sql.DB
}

// pingTimeout — таймаут для проверки соединения с БД.
const pingTimeout = 1 * time.Second

// NewPingHandler создаёт новый PingHandler с указанным подключением к БД.
func NewPingHandler(db *sql.DB) *PingHandler {
	return &PingHandler{db: db}
}

// Ping — обработчик GET /ping.
// Проверяет соединение с базой данных. Возвращает 200 OK в случае успеха или 500 при ошибке.
func (ph PingHandler) Ping(w http.ResponseWriter, r *http.Request) {
	if ph.db == nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	ctxTimeout, cancel := context.WithTimeout(r.Context(), pingTimeout)
	defer cancel()

	if err := ph.db.PingContext(ctxTimeout); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)

}
