package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/sign"
)

// SignMiddleware — middleware для проверки HMAC-SHA256 подписи запросов и подписи ответов.
// Если signer равен nil, middleware пропускает запрос без проверки.
// Проверяет заголовок HashSHA256 в запросе и добавляет подпись к ответу.
func SignMiddleware(signer *sign.Signer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if signer == nil {
				next.ServeHTTP(w, r)
				return
			}

			cw := sign.NewCaptureResponseWriter()

			// Если заголовок HashSHA256 есть, то проверяем подпись.
			if gotHash := r.Header.Get("HashSHA256"); gotHash != "" {
				if r.Body == nil {
					r.Body = http.NoBody
				}
				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(cw, "cannot read request body", http.StatusBadRequest)
					// Добавляем подпись к ответу и отправляем клиенту
					cw.FlushSignedResponse(w, signer)
					return
				}
				// Верифицируем подпись
				verify, err := signer.Verify(gotHash, body)
				if err != nil || !verify {
					http.Error(cw, "invalid HashSHA256 header", http.StatusBadRequest)
					// Добавляем подпись к ответу и отправляем клиенту
					cw.FlushSignedResponse(w, signer)
					return
				}
				// Восстанавливаем тело запроса для дальнейшей обработки
				r.Body = io.NopCloser(bytes.NewReader(body))
				r.ContentLength = int64(len(body))
			}
			// Вызываем следующий обработчик с перехватывающим ResponseWriter
			next.ServeHTTP(cw, r)
			// Добавляем подпись к ответу и отправляем клиенту
			cw.FlushSignedResponse(w, signer)
		})
	}
}
