package compress

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"sync"
)

var bufPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(nil)
	},
}

var gzipReaderPool = sync.Pool{
	New: func() interface{} {
		return new(gzip.Reader)
	},
}

// Реализуем интерфейс http.ResponseWriter для сжатия данных, отправляемых клиенту
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	zw := gzipWriterPool.Get().(*gzip.Writer)
	zw.Reset(w)
	return &compressWriter{
		w:  w,
		zw: zw,
	}
}

func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	err := c.zw.Close()
	gzipWriterPool.Put(c.zw)
	return err
}

// Реализуем интерфейс io.ReadCloser для декомпрессии данных, получаемых от клиента
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr := gzipReaderPool.Get().(*gzip.Reader)
	if err := zr.Reset(r); err != nil {
		gzipReaderPool.Put(zr)
		return nil, err
	}
	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	err := c.zr.Close()
	gzipReaderPool.Put(c.zr)
	if err1 := c.r.Close(); err1 != nil {
		return err1
	}
	return err
}

// GzipData сжимает входной срез данных в формате gzip.
func GzipData(data []byte) ([]byte, error) {
	// Ранее на каждый вызов создавался новый буфер, теперь переиспользуется
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()

	// Ранее на каждый вызов создавался новый gzip.Writer, теперь переиспользуется
	zw := gzipWriterPool.Get().(*gzip.Writer)
	zw.Reset(buf)

	if _, err := zw.Write(data); err != nil {
		_ = zw.Close()
		gzipWriterPool.Put(zw)
		bufPool.Put(buf)
		return nil, err
	}

	if err := zw.Close(); err != nil {
		gzipWriterPool.Put(zw)
		bufPool.Put(buf)
		return nil, err
	}

	// Копируем результат, чтобы вернуть buf в пул
	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	gzipWriterPool.Put(zw)
	bufPool.Put(buf)
	return out, nil
}
