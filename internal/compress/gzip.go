// Package compress предоставляет утилиты для сжатия/декомпрессии данных в формате gzip.
// Включает обёртки над http.ResponseWriter и io.ReadCloser для прозрачного сжатия трафика.
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

// compressWriter реализует интерфейс http.ResponseWriter для сжатия данных gzip.
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// NewCompressWriter создаёт новый compressWriter, оборачивающий http.ResponseWriter.
func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	zw := gzipWriterPool.Get().(*gzip.Writer)
	zw.Reset(w)
	return &compressWriter{
		w:  w,
		zw: zw,
	}
}

// Header возвращает http.Header для последующей записи.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write записывает данные, сжимая их gzip.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader устанавливает HTTP-статус заголовка и Content-Encoding: gzip для успешных ответов.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и возвращает его в пул.
func (c *compressWriter) Close() error {
	err := c.zw.Close()
	gzipWriterPool.Put(c.zw)
	return err
}

// compressReader реализует интерфейс io.ReadCloser для декомпрессии gzip-данных.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader создаёт новый compressReader из io.ReadCloser.
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

// Read читает декомпрессированные данные.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close закрывает gzip.Reader и возвращает его в пул, затем закрывает исходный io.ReadCloser.
func (c *compressReader) Close() error {
	err := c.zr.Close()
	gzipReaderPool.Put(c.zr)
	if err1 := c.r.Close(); err1 != nil {
		return err1
	}
	return err
}

// GzipData сжимает входной срез данных в формате gzip с использованием sync.Pool.
func GzipData(data []byte) ([]byte, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()

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

	out := make([]byte, buf.Len())
	copy(out, buf.Bytes())
	gzipWriterPool.Put(zw)
	bufPool.Put(buf)
	return out, nil
}
