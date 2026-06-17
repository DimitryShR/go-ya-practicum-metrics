package middleware

import (
	"net/http"
	"time"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/logger"
	"go.uber.org/zap"
)

// responseData содержит информацию об ответе для логирования.
type responseData struct {
	statusCode int
	size       int
}

// loggingResponseWriter оборачивает http.ResponseWriter для перехвата статуса и размера ответа.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

// Write перехватывает запись ответа, сохраняя размер.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader перехватывает установку статус-кода.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.statusCode = statusCode
}

// LogRequest — middleware для логирования HTTP-запросов.
// Логирует метод, путь, статус ответа, размер и длительность.
func LogRequest(h http.Handler) http.Handler {
	return http.Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		responseData := &responseData{}

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		h.ServeHTTP(lrw, r)

		duration := time.Since(start)

		if responseData.statusCode == 0 {
			responseData.statusCode = http.StatusOK
		}

		logger.Log.Info("Request completed",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
			zap.Int("status_code", responseData.statusCode),
			zap.Int("response_size", responseData.size),
			zap.Duration("duration", duration),
		)
	}))
}
