package handler

import (
	"context"
	"database/sql"
	"net/http"
	"time"
)

type PingHandler struct {
	db *sql.DB
}

const pingTimeout = 1 * time.Second

func NewPingHandler(db *sql.DB) *PingHandler {
	return &PingHandler{db: db}
}

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
