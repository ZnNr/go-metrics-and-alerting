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
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"html/template"
	"io"
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

	if err := collector.Collector.Collect(
		collector.MetricRequest{
			ID:    metricName,
			MType: metricType,
		}, metricValue); err != nil {
		w.WriteHeader(h.getStatusOnError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := io.WriteString(w, fmt.Sprintf("inserted metric %q with value %q", metricName, metricValue)); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.Header().Set("content-length", strconv.Itoa(len(metricName)))
}

// SaveMetricFromJSONHandler - a method for saving metric from JSON body of http request.
func (h *Handler) SaveMetricFromJSONHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !h.checkSubscription(w, buf, r.Header.Get("HashSHA256")) {
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
	var metric collector.MetricRequest
	if err := json.Unmarshal(message, &metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	var metricValue string
	switch metric.MType {
	case collector.Counter:
		metricValue = strconv.Itoa(int(*metric.Delta))
	case collector.Gauge:
		metricValue = strconv.FormatFloat(*metric.Value, 'f', 11, 64)
	default:
		w.WriteHeader(http.StatusNotImplemented)
		return
	}

	if err := collector.Collector.Collect(metric, metricValue); err != nil {
		w.WriteHeader(h.getStatusOnError(err))
		return
	}

	resultJSON, err := collector.Collector.GetMetricJSON(metric.ID)
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
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r.Body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !h.checkSubscription(w, buf, r.Header.Get("HashSHA256")) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metrics []collector.MetricRequest
	if err := json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var results []byte
	for _, metric := range metrics {
		var metricValue string
		switch metric.MType {
		case collector.Counter:
			metricValue = strconv.Itoa(int(*metric.Delta))
		case collector.Gauge:
			metricValue = strconv.FormatFloat(*metric.Value, 'f', 11, 64)
		default:
			w.WriteHeader(http.StatusNotImplemented)
			return
		}

		if err := collector.Collector.Collect(metric, metricValue); err != nil {
			w.WriteHeader(h.getStatusOnError(err))
			return
		}

		resultJSON, err := collector.Collector.GetMetricJSON(metric.ID)
		if err != nil {
			w.WriteHeader(h.getStatusOnError(err))
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

	gotHash := r.Header.Get("HashSHA256")
	want := h.getHash(buf.Bytes())
	if gotHash != "" {
		w.Header().Set("HashSHA256", want)
	}
	if !h.checkSubscription(w, buf, r.Header.Get("HashSHA256")) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var metric collector.MetricRequest
	if err := json.Unmarshal(buf.Bytes(), &metric); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resultJSON, err := collector.Collector.GetMetric(metric.ID)
	if err != nil {
		w.WriteHeader(h.getStatusOnError(err))
		return
	}
	switch metric.MType {
	case collector.Counter:
		metric.Delta = resultJSON.CounterValue
	case collector.Gauge:
		metric.Value = resultJSON.GaugeValue
	}
	answer, _ := json.Marshal(metric)

	if _, err = w.Write(answer); err != nil {
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

	if metricType != collector.Counter && metricType != collector.Gauge {
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
	value, err := collector.Collector.GetMetric(metricName)
	if err != nil {
		w.WriteHeader(h.getStatusOnError(err))
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err = io.WriteString(w, *value.TextValue); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
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
	for _, n := range collector.Collector.GetAvailableMetrics() {
		page += fmt.Sprintf("<h1>	%s</h1>", n)
	}
	tmpl, _ := template.New("data").Parse("<h1>AVAILABLE METRICS</h1>{{range .}}<h3>{{ .}}</h3>{{end}}")
	if err := tmpl.Execute(w, collector.Collector.GetAvailableMetrics()); err != nil {
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

func (h *Handler) getStatusOnError(err error) int {
	statusCodes := map[error]int{
		collector.ErrBadRequest:     http.StatusBadRequest,
		collector.ErrNotImplemented: http.StatusNotImplemented,
		collector.ErrNotFound:       http.StatusNotFound,
	}

	if statusCode, ok := statusCodes[err]; ok {
		return statusCode
	}

	return http.StatusInternalServerError
}

// getHash - a method for getting hash from request body.
func (h *Handler) getHash(body []byte) string {
	want := sha256.Sum256(body)
	wantDecoded := fmt.Sprintf("%x", want)
	return wantDecoded
}

func New(db string, key string, cryptoKey string) (*Handler, error) {
	handler := &Handler{
		dbAddress: db,
		key:       key,
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

type Handler struct {
	dbAddress string
	key       string
	cryptoKey *rsa.PrivateKey
}
