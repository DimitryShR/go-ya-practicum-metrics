package agent

type HTTPStatusError struct {
	Err  error
	Code int
}

func (e HTTPStatusError) Error() string {
	if e.Err == nil {
		return "http status error"
	}
	return e.Err.Error()
}

func (e HTTPStatusError) Unwrap() error {
	return e.Err
}

func NewHTTPStatusError(err error, code int) *HTTPStatusError {
	return &HTTPStatusError{
		Err:  err,
		Code: code,
	}
}
