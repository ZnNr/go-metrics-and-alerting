// Package logger предоставляет функционал для логирования HTTP запросов.
// Используется пакет go.uber.org/zap для логирования.
package logger

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

var SugarLogger zap.SugaredLogger

// RequestLogger возвращает обработчик HTTP запросов, который выполняет логирование.
func RequestLogger(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rd := &responseData{}
		lw := loggingResponseWriter{
			ResponseWriter: w, // встраиваем оригинальный http.ResponseWriter
			responseData:   rd,
		}
		h.ServeHTTP(&lw, r)

		duration := time.Since(start)
		SugarLogger.Infoln(
			"uri", r.RequestURI,
			"method", r.Method,
			"status", rd.status, // получаем перехваченный код статуса ответа
			"duration", duration,
			"size", rd.size, // получаем перехваченный размер ответа
			"request headers", r.Header,
			"response headers", w.Header(),
		)
		w.Header().Set("content-type", "Content-Type: application/json")
	}
	return http.HandlerFunc(logFn)
}

// Write записывает данные в http.ResponseWriter и обновляет данные о размере ответа.
func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size
	return size, err
}

// WriteHeader устанавливает код статуса ответа и обновляет данные о коде статуса.
func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

type (
	// responseData содержит информацию о ответе на запрос.
	responseData struct {
		status int
		size   int
	}

	// loggingResponseWriter обертка над http.ResponseWriter для логирования.
	loggingResponseWriter struct {
		http.ResponseWriter
		responseData *responseData
	}
)
