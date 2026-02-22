package middleware

import (
	"net/http"
	"time"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/logger"
	"go.uber.org/zap"
)

type responseData struct {
	statusCode int
	size       int
}

type loggingResponseWriter struct {
	http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
	responseData        *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b) // вызываем оригинальный Write
	r.responseData.size += size            // увеличиваем размер ответа
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode) // вызываем оригинальный WriteHeader
	r.responseData.statusCode = statusCode   // сохраняем статус код
}

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
