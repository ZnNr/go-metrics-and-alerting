package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/json"
	"encoding/pem"
	"fmt"
	collector2 "github.com/ZnNr/go-musthave-metrics.git/internal/agent/collector"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"html/template"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

// SaveMetricHandler - a method for saving metric from url.
func (h *Handler) SaveMetricHandler(w http.ResponseWriter, r *http.Request) {

	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if err := collector2.Collector().Collect(
		collector2.MetricRequest{
			ID:    metricName,
			MType: metricType,
		}, metricValue); err != nil {
		w.WriteHeader(h.getStatusOnError(err))
		return
	}

	if _, err := io.WriteString(w, fmt.Sprintf("saved metric %q with value %q", metricName, metricValue)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.Header().Set("content-length", strconv.Itoa(len(metricName)))
}

// SaveMetricFromJSONHandler - a method for saving metric from JSON body of http request.
func (h *Handler) SaveMetricFromJSONHandler(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// decrypt message if crypto key was specified
	message := buf.Bytes()
	if h.cryptoKey != nil {
		encryptedData, err := rsa.DecryptPKCS1v15(rand.Reader, h.cryptoKey, message)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		message = encryptedData
	}

	// unmarshall request body and get metric
	var metric collector2.MetricRequest
	if err := json.Unmarshal(message, &metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// save metric
	resultJSON, err := h.collectMetric(metric)
	if err != nil {
		w.WriteHeader(h.getStatusOnError(err))
		return
	}

	if _, err = w.Write(resultJSON); err != nil {
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-length", strconv.Itoa(len(metric.ID)))
	w.Header().Set("content-type", "application/json")
}

// SaveListMetricsFromJSONHandler - a method for saving a list of metrics from JSON body of http request.
func (h *Handler) SaveListMetricsFromJSONHandler(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// unmarshall request body and get metric
	var metrics []collector2.MetricRequest
	if err := json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var results []byte
	// save all metrics from request
	for _, metric := range metrics {
		resultJSON, err := h.collectMetric(metric)
		if err != nil {
			w.WriteHeader(h.getStatusOnError(err))
			return
		}
		results = append(results, resultJSON...)
	}
	if _, err := w.Write(results); err != nil {
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// GetMetricFromJSONHandler - a method for getting metrics by JSON from http request.
func (h *Handler) GetMetricFromJSONHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "application/json")
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// unmarshall body and get requested metric name
	var metric collector2.MetricRequest
	if err := json.Unmarshal(buf.Bytes(), &metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// get metric from collector
	resultJSON, err := collector2.Collector().GetMetric(metric.ID)
	if err != nil {
		w.WriteHeader(h.getStatusOnError(err))
		return
	}
	// get metric value
	switch metric.MType {
	case collector2.Counter:
		metric.Delta = resultJSON.CounterValue
	case collector2.Gauge:
		metric.Value = resultJSON.GaugeValue
	}
	answer, err := json.Marshal(metric)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if _, err = w.Write(answer); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-length", strconv.Itoa(len(metric.ID)))
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
}

// GetMetricHandler - a metric for getting metric from url.
func (h *Handler) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")

	if metricType != collector2.Counter && metricType != collector2.Gauge {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
	// get requested metric from collector
	value, err := collector2.Collector().GetMetric(metricName)
	if err != nil {
		w.WriteHeader(h.getStatusOnError(err))
		return
	}

	if _, err = io.WriteString(w, *value.TextValue); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("content-type", "text/plain; charset=utf-8")
}

// ShowMetricsHandler - a method for getting all available metrics from server.
func (h *Handler) ShowMetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("content-type", "Content-Type: text/html; charset=utf-8")
	if r.URL.Path != "/" {
		http.Error(w, fmt.Sprintf("wrong path %q", r.URL.Path), http.StatusNotFound)
		return
	}
	var page string
	for _, n := range collector2.Collector().GetAvailableMetrics() {
		page += fmt.Sprintf("<h1>	%s</h1>", n)
	}
	tmpl, _ := template.New("data").Parse("<h1>AVAILABLE METRICS</h1>{{range .}}<h3>{{ .}}</h3>{{end}}")
	if err := tmpl.Execute(w, collector2.Collector().GetAvailableMetrics()); err != nil {
		return
	}
	w.Header().Set("content-type", "Content-Type: text/html; charset=utf-8")
}

// CheckDatabaseAvailability выполняет проверку доступности базы данных (Ping).
func (h *Handler) CheckDatabaseAvailability(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", h.dbAddress)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err = w.Write([]byte("pong")); err != nil {
		return
	}
}

func (h *Handler) CheckSubscriptionHandler(hh http.Handler) http.Handler {
	checkFn := func(w http.ResponseWriter, r *http.Request) {
		bodyBytes, _ := io.ReadAll(r.Body)
		r.Body.Close()

		buf := bytes.NewBuffer(bodyBytes)

		gotHash := r.Header.Get("HashSHA256")
		want := h.getHash(buf.Bytes())
		if gotHash != "" {
			w.Header().Set("HashSHA256", want)
		}
		if !h.checkSubscription(w, *buf, r.Header.Get("HashSHA256")) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		hh.ServeHTTP(w, r)
	}
	return http.HandlerFunc(checkFn)
}

// CheckSubnetHandler возвращает обработчик, который проверяет, принадлежит ли входящий IP-адрес доверенной подсети.
func (h *Handler) CheckSubnetHandler(hh http.Handler) http.Handler {
	// Функция checkSubnetFn выполняет проверку подсети перед обработкой запроса.
	checkSubnetFn := func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, установлена ли доверенная подсеть.
		if h.trustedIPNet != nil {
			realIP := r.Header.Get("X-Real-IP") // Получаем реальный IP-адрес из заголовка запроса.
			clientIP := net.ParseIP(realIP)
			// Проверяем, принадлежит ли реальный IP-адрес доверенной подсети.
			if !h.trustedIPNet.Contains(clientIP) {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}
		hh.ServeHTTP(w, r) // Если IP принадлежит доверенной подсети, передаем управление следующему обработчику.
	}
	// Возвращаем обработчик, который будет выполнять проверку доверенной подсети перед вызовом переданного обработчика.
	return http.HandlerFunc(checkSubnetFn)
}

// collectMetric - метод для сохранения метрики.
func (h *Handler) collectMetric(metric collector2.MetricRequest) ([]byte, error) {
	c := collector2.Collector()

	// get metric value
	var metricValue string
	switch metric.MType {
	case collector2.Counter:
		metricValue = strconv.Itoa(int(*metric.Delta))
	case collector2.Gauge:
		metricValue = strconv.FormatFloat(*metric.Value, 'f', 11, 64)
	default:
		return nil, collector2.ErrNotImplemented
	}

	// save metric
	if err := c.Collect(metric, metricValue); err != nil {
		return nil, err
	}

	// get saved metric in JSON format for response
	resultJSON, err := c.GetMetricJSON(metric.ID)
	if err != nil {
		return nil, err
	}
	return resultJSON, err
}

// checkSubscription - метод для проверки подписки и хеша.
func (h *Handler) checkSubscription(w http.ResponseWriter, buf bytes.Buffer, header string) bool {
	want := h.getHash(buf.Bytes())
	if header != "" {
		w.Header().Set("HashSHA256", want)
	}
	if h.key != "" && len(want) != 0 && header != "" {

		return header == want
	}
	return true
}

// getStatusOnError - метод для получения статусного кода на основе ошибки.
func (h *Handler) getStatusOnError(err error) int {
	statusCodes := map[error]int{
		collector2.ErrBadRequest:     http.StatusBadRequest,
		collector2.ErrNotImplemented: http.StatusNotImplemented,
		collector2.ErrNotFound:       http.StatusNotFound,
	}

	if statusCode, ok := statusCodes[err]; ok {
		return statusCode
	}

	return http.StatusInternalServerError
}

// getHash - метод для получения хеша из тела запроса.
func (h *Handler) getHash(body []byte) string {
	want := sha256.Sum256(body)
	wantDecoded := fmt.Sprintf("%x", want)
	return wantDecoded
}

// New - функция создания нового экземпляра Handler.
func New(db string, key string, cryptoKey string, trustedSubnet string) (*Handler, error) {
	handler := &Handler{
		dbAddress:     db,
		key:           key,
		trustedSubnet: trustedSubnet,
	}
	if trustedSubnet != "" {
		_, ipnet, err := net.ParseCIDR(trustedSubnet)
		if err != nil {
			return nil, fmt.Errorf("error parsing trusted subnet: %v", err)
		}
		handler.trustedIPNet = ipnet
	}
	if cryptoKey != "" {
		b, err := os.ReadFile(cryptoKey)
		if err != nil {
			return nil, fmt.Errorf("error while reading file with crypto private key: %w", err)
		}
		block, _ := pem.Decode(b)
		privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing private key: %w", err)
		}
		handler.cryptoKey = privateKey.(*rsa.PrivateKey)
	}
	return handler, nil
}

// Handler - структура, представляющая обработчик запросов.
// Она содержит методы для сохранения метрик, получения метрик, проверки доступности базы данных и другие.
type Handler struct {
	dbAddress     string
	trustedSubnet string
	trustedIPNet  *net.IPNet // Добавьте новое поле для хранения IP-подсети
	key           string
	cryptoKey     *rsa.PrivateKey
}
