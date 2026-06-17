package agent

// HTTPStatusError представляет ошибку с HTTP-статусом.
type HTTPStatusError struct {
	Err  error
	Code int
}

// Error возвращает строковое представление ошибки.
func (e HTTPStatusError) Error() string {
	if e.Err == nil {
		return "http status error"
	}
	return e.Err.Error()
}

// Unwrap возвращает обёрнутую ошибку для errors.Is/As.
func (e HTTPStatusError) Unwrap() error {
	return e.Err
}

// NewHTTPStatusError создаёт новую HTTPStatusError с заданной ошибкой и кодом статуса.
func NewHTTPStatusError(err error, code int) *HTTPStatusError {
	return &HTTPStatusError{
		Err:  err,
		Code: code,
	}
}
