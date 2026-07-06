package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/crypto/rsacrypto"
)

// DecryptMiddleware создаёт middleware для расшифровки тела запроса с использованием RSA-ключа.
// Если ключ не задан (nil), middleware пропускает запрос без изменений.
// Если заголовок X-Encrypted отсутствует, тело передаётся как есть (обратная совместимость).
// При ошибке расшифровки возвращается HTTP 400.
func DecryptMiddleware(privateKey *rsa.PrivateKey) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Если приватный ключ не задан
			// И получено не зашифрованное сообщение — пропускаем без изменений
			// Иначе возвращаем ошибку об отсутствии поддержки шифрования
			if privateKey == nil {
				if r.Header.Get("X-Encrypted") == "true" {
					http.Error(w, "server does not support encryption", http.StatusBadRequest)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Если заголовок X-Encrypted отсутствует — пропускаем без изменений
			if r.Header.Get("X-Encrypted") != "true" {
				next.ServeHTTP(w, r)
				return
			}

			// Читаем зашифрованное тело
			encryptedBody, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "failed to read request body", http.StatusBadRequest)
				return
			}
			r.Body.Close()

			// Расшифровываем
			decryptedBody, err := rsacrypto.Decrypt(privateKey, encryptedBody)
			if err != nil {
				http.Error(w, "failed to decrypt request body", http.StatusBadRequest)
				return
			}

			// Заменяем тело запроса на расшифрованное
			r.Body = io.NopCloser(bytes.NewReader(decryptedBody))
			r.ContentLength = int64(len(decryptedBody))

			next.ServeHTTP(w, r)
		})
	}
}
