package compress_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DimitryShR/go-ya-practicum-metrics/internal/compress"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- GzipData ----------

func TestGzipData_Success(t *testing.T) {
	data := []byte(`{"id":"Alloc","type":"gauge","value":1234567.89}`)

	compressed, err := compress.GzipData(data)
	require.NoError(t, err)
	require.NotEmpty(t, compressed)

	// compressed data must be different from original
	assert.NotEqual(t, data, compressed)

	// decompress and verify
	decompressed := decompressGzip(t, compressed)
	assert.Equal(t, data, decompressed)
}

func TestGzipData_EmptyInput(t *testing.T) {
	compressed, err := compress.GzipData([]byte{})
	require.NoError(t, err)

	// empty input after gzip should produce non-empty gzip stream
	assert.NotEmpty(t, compressed)

	decompressed := decompressGzip(t, compressed)
	assert.Equal(t, []byte{}, decompressed)
}

func TestGzipData_NilInput(t *testing.T) {
	compressed, err := compress.GzipData(nil)
	require.NoError(t, err)
	assert.NotEmpty(t, compressed)

	decompressed := decompressGzip(t, compressed)
	assert.Empty(t, decompressed)
}

func TestGzipData_RoundTrip(t *testing.T) {
	payloads := [][]byte{
		[]byte(`{"id":"Alloc","type":"gauge","value":1234567.89}`),
		[]byte(`{"id":"PollCount","type":"counter","delta":1}`),
		[]byte(`[{"id":"A","type":"gauge","value":1},{"id":"B","type":"counter","delta":2}]`),
	}

	for _, original := range payloads {
		compressed, err := compress.GzipData(original)
		require.NoError(t, err)

		decompressed := decompressGzip(t, compressed)
		assert.Equal(t, original, decompressed)
	}
}

// ---------- compressReader ----------

func TestNewCompressReader_Success(t *testing.T) {
	original := []byte("hello, world")
	compressed := compressForTest(t, original)

	reader, err := compress.NewCompressReader(io.NopCloser(bytes.NewReader(compressed)))
	require.NoError(t, err)
	defer reader.Close()

	got, err := io.ReadAll(reader)
	require.NoError(t, err)
	assert.Equal(t, original, got)
}

func TestNewCompressReader_InvalidData(t *testing.T) {
	_, err := compress.NewCompressReader(io.NopCloser(bytes.NewReader([]byte("not gzip data"))))
	assert.Error(t, err)
}

func TestCompressReader_Read(t *testing.T) {
	original := []byte("some test data for reading")
	compressed := compressForTest(t, original)

	reader, err := compress.NewCompressReader(io.NopCloser(bytes.NewReader(compressed)))
	require.NoError(t, err)
	defer reader.Close()

	buf := make([]byte, 4)
	var result []byte
	for {
		n, err := reader.Read(buf)
		result = append(result, buf[:n]...)
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
	}

	assert.Equal(t, original, result)
}

func TestCompressReader_Close(t *testing.T) {
	original := []byte("data for close test")
	compressed := compressForTest(t, original)

	reader, err := compress.NewCompressReader(io.NopCloser(bytes.NewReader(compressed)))
	require.NoError(t, err)

	// read some data
	buf := make([]byte, 4)
	_, err = reader.Read(buf)
	require.NoError(t, err)

	// close should succeed
	assert.NoError(t, reader.Close())
}

// ---------- compressWriter ----------

func TestNewCompressWriter(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := compress.NewCompressWriter(recorder)
	require.NotNil(t, writer)
}

func TestCompressWriter_Write(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := compress.NewCompressWriter(recorder)
	defer writer.Close()

	data := []byte("some response data")
	n, err := writer.Write(data)
	require.NoError(t, err)
	assert.Equal(t, len(data), n)

	// flush gzip data
	require.NoError(t, writer.Close())

	// the body must be valid gzip
	compressed := recorder.Body.Bytes()
	decompressed := decompressGzip(t, compressed)
	assert.Equal(t, data, decompressed)
}

func TestCompressWriter_WriteHeader_SuccessStatus(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := compress.NewCompressWriter(recorder)
	defer writer.Close()

	writer.WriteHeader(http.StatusOK)

	// for status < 300 Content-Encoding header should be set
	assert.Equal(t, "gzip", recorder.Header().Get("Content-Encoding"))
	assert.Equal(t, http.StatusOK, recorder.Code)
}

func TestCompressWriter_WriteHeader_ErrorStatus(t *testing.T) {
	statusCodes := []int{
		http.StatusBadRequest,
		http.StatusNotFound,
		http.StatusInternalServerError,
		http.StatusServiceUnavailable,
	}

	for _, status := range statusCodes {
		t.Run(http.StatusText(status), func(t *testing.T) {
			recorder := httptest.NewRecorder()
			writer := compress.NewCompressWriter(recorder)
			defer writer.Close()

			writer.WriteHeader(status)

			// for status >= 300 Content-Encoding should NOT be set
			assert.Empty(t, recorder.Header().Get("Content-Encoding"))
			assert.Equal(t, status, recorder.Code)
		})
	}
}

func TestCompressWriter_Close(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := compress.NewCompressWriter(recorder)

	_, err := writer.Write([]byte("data"))
	require.NoError(t, err)

	// close should succeed and flush data
	assert.NoError(t, writer.Close())

	// after close, the body should contain gzip-compressed data
	compressed := recorder.Body.Bytes()
	assert.NotEmpty(t, compressed)

	decompressed := decompressGzip(t, compressed)
	assert.Equal(t, []byte("data"), decompressed)
}

// ---------- helpers ----------

// compressForTest compresses data with standard gzip.Writer for test setup
func compressForTest(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err := w.Write(data)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return buf.Bytes()
}

// decompressGzip decompresses gzip-compressed data
func decompressGzip(t *testing.T, data []byte) []byte {
	t.Helper()
	r, err := gzip.NewReader(bytes.NewReader(data))
	require.NoError(t, err)
	defer r.Close()
	got, err := io.ReadAll(r)
	require.NoError(t, err)
	return got
}
