package sign

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCaptureResponseWriter_WriteSetsDefaultStatusAndBody(t *testing.T) {
	cw := NewCaptureResponseWriter()

	n, err := cw.Write([]byte("test body"))

	assert.NoError(t, err)
	assert.Equal(t, len("test body"), n)
	assert.Equal(t, http.StatusOK, cw.statusCode)
	assert.Equal(t, "test body", cw.body.String())
}

func TestCaptureResponseWriter_WriteHeaderSetsStatusOnlyOnce(t *testing.T) {
	cw := NewCaptureResponseWriter()

	cw.WriteHeader(http.StatusCreated)
	cw.WriteHeader(http.StatusInternalServerError)

	assert.Equal(t, http.StatusCreated, cw.statusCode)
}

func TestCaptureResponseWriter_HeaderReturnsMutableHeader(t *testing.T) {
	cw := NewCaptureResponseWriter()

	cw.Header().Set("Content-Type", "application/json")

	assert.Equal(t, "application/json", cw.header.Get("Content-Type"))
}
