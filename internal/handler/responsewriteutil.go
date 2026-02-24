package handler

import (
	"encoding/json"
	"net/http"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/logger"
	"go.uber.org/zap"
)

type errorResponse struct {
	Error string `json:"error"`
}

// writeJSON - вспомогательная функция для отправки JSON ответа с заданным статусом и данными
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if v == nil {
		return
	}

	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Log.Info("failed to encode JSON response", zap.Error(err))
	}
}

// writeJSONError - вспомогательная функция для отправки JSON ответа с ошибкой
func writeJSONError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}
