package sign

import (
	"bytes"
	"net/http"
)

// captureResponseWriter — это реализация http.ResponseWriter,
// основной целью которой является буферизация тела ответа для последующего использования в подписи.
type captureResponseWriter struct {
	header     http.Header
	body       bytes.Buffer
	statusCode int
}

// NewCaptureResponseWriter создаёт новый экземпляр captureResponseWriter для захвата данных ответа.
func NewCaptureResponseWriter() *captureResponseWriter {
	return &captureResponseWriter{
		header: make(http.Header),
	}
}

// Header возвращает заголовки ответа.
func (cw *captureResponseWriter) Header() http.Header {
	return cw.header
}

// WriteHeader устанавливает код статуса ответа.
func (cw *captureResponseWriter) WriteHeader(statusCode int) {
	if cw.statusCode == 0 {
		cw.statusCode = statusCode
	}
}

// Write записывает данные в тело ответа и сохраняет код статуса, если он еще не установлен.
func (cw *captureResponseWriter) Write(p []byte) (int, error) {
	if cw.statusCode == 0 {
		cw.statusCode = http.StatusOK
	}
	return cw.body.Write(p)
}

// FlushSignedResponse отправляет буферизированный ответ клиенту, добавляя подпись в заголовок.
func (cw *captureResponseWriter) FlushSignedResponse(w http.ResponseWriter, signer *Signer) {
	body := cw.body.Bytes()
	for k, values := range cw.Header() {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.Header().Set("HashSHA256", signer.Sign(body))

	if cw.statusCode == 0 {
		cw.statusCode = http.StatusOK
	}

	w.WriteHeader(cw.statusCode)
	_, _ = w.Write(body)
}
