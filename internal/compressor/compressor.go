// Package compressor предоставляет функционал для сжатия и декомпрессии данных в HTTP-запросах и ответах.
// Используется пакет compress/gzip для работы с сжатием данных.
package compressor

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// NewCompressWriter создает новый экземпляр compressWriter с инициализацией gzip.Writer.
func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header возвращает HTTP-заголовки обертки.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write записывает данные в сжатом формате с использованием gzip.Writer.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// WriteHeader устанавливает HTTP статус ответа и добавляет Content-Encoding, если код статуса меньше 300.
func (c *compressWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader создает экземпляр compressReader и инициализирует его.
func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read реализует метод Read интерфейса io.ReadCloser и выполняет чтение декомпрессированных данных.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close реализует метод Close интерфейса io.ReadCloser и закрывает все открытые ресурсы.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}

// HTTPCompressHandler является обработчиком HTTP, который сжимает данные ответа, если клиент поддерживает сжатие.
func HTTPCompressHandler(h http.Handler) http.Handler {
	zipFn := func(w http.ResponseWriter, r *http.Request) {
		ow := w
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")

		if sendsGzip {
			cw := NewCompressWriter(w)
			ow = cw
			defer cw.Close()

			cr, err := NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			r.Body = cr
			defer cr.Close()
		} else if supportsGzip && !sendsGzip {
			cw := NewCompressWriter(w)
			ow = cw
			ow.Header().Set("Content-Encoding", "gzip")
			defer cw.Close()

		}
		h.ServeHTTP(ow, r)
	}
	return http.HandlerFunc(zipFn)
}
