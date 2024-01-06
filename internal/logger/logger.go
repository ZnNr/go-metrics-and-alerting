package logger

import (
	"go.uber.org/zap"
	"net/http"
	"time"
)

var SugarLogger zap.SugaredLogger

func RequestLogger(h http.Handler) http.Handler {
	logFn := func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rd := &responseData{
			status: 0,
			size:   0,
		}
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
		)
	}
	return http.HandlerFunc(logFn)
}

type responseData struct {
	status int
	size   int
}

// добавляем реализацию http.ResponseWriter
type loggingResponseWriter struct {
	http.ResponseWriter // встраиваем оригинальный http.ResponseWriter
	responseData        *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	// записываем ответ, используя оригинальный http.ResponseWriter
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	// записываем код статуса, используя оригинальный http.ResponseWriter
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}