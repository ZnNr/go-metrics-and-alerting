// Package agent
// Модуль agent собирает определенный набор runtime и gopsutil
// метрик и отправляет их на сервер.
package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/ZnNr/go-musthave-metrics.git/internal/agent/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/agent/metrics"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	pb "github.com/ZnNr/go-musthave-metrics.git/proto"
	"github.com/avast/retry-go"
	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"sync"
	"time"
)

// CollectMetrics — метод для сбора метрик времени выполнения и gopsutil.
func (a *Agent) CollectMetrics(ctx context.Context) {
	storeTicker := time.NewTicker(time.Second * time.Duration(a.params.PollInterval))
	defer storeTicker.Stop()
	go func() {
		for {
			select {
			case <-ctx.Done():
				a.log.Warn("Collection of metrics stopped")
				return
			case <-storeTicker.C:
				a.storage.RuntimeMetricStore()
				a.storage.GopsutilMetricStore()
			}
		}
	}()
}

// SendMetricsLoop — метод отправки метрик на сервер по таймеру
func (a *Agent) SendMetricsLoop(ctx context.Context) (err error) {
	numRequests := make(chan struct{}, a.params.RateLimit)
	reportTicker := time.NewTicker(time.Duration(a.params.ReportInterval) * time.Second)
	defer reportTicker.Stop()

	if a.params.GrpcRunAddr != "" {
		// Устанавливаем соединение с сервером
		conn, err := grpc.NewClient(a.params.GrpcRunAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		defer conn.Close()
		a.grpcMetricsClient = pb.NewMetricsClient(conn)
	}

	for {
		select {
		case <-ctx.Done():
			a.log.Warn("Sending of metrics stopped")
			return nil
		case <-reportTicker.C:
			select {
			case <-ctx.Done():
				a.log.Info("reportTicker.C triggered")
				return nil
			case numRequests <- struct{}{}:
				if err = a.SendMetrics(ctx); err != nil {
					return err
				}
			default:
				a.log.Info("rate limit is exceeded")
			}
		}
	}
}

// SendMetrics - метод для отправки метрик
func (a *Agent) SendMetrics(ctx context.Context) error {
	if err := a.sendHTTP(ctx); err != nil {
		return err
	}
	a.log.Info("metrics were successfully sent to HTTP server")
	if a.params.GrpcRunAddr != "" {
		if err := a.sendGrpc(ctx); err != nil {
			return err
		}
		a.log.Info("metrics were successfully sent to gRPC server")
	}
	return nil
}

func (a *Agent) sendGrpc(ctx context.Context) error {
	for _, v := range collector.Collector().Metrics {
		request := pb.MetricRequest{
			ID:    v.ID,
			MType: v.MType,
		}
		switch request.MType {
		case collector.Gauge:
			request.Value = *v.GaugeValue
		case collector.Counter:
			request.Delta = *v.CounterValue
		}

		if _, err := a.grpcMetricsClient.SaveMetricFromJSON(ctx, &request); err != nil {
			return errors.Errorf("error while sending metric to grpc server: %s", err.Error())
		}
	}
	return nil
}

// sendHTTP — метод, инкапсулирующий логику отправки http-запроса на сервер.
func (a *Agent) sendHTTP(ctx context.Context) error {
	req := a.client.SetRetryCount(3).R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Encoding", "gzip").
		SetContext(ctx)
	a.SetRealIPFromRequest(req) // Вызываем метод SetRealIPFromRequest для сохранения реального IP-адреса клиента

	var wg sync.WaitGroup

	for _, v := range collector.Collector().Metrics {
		wg.Add(1)
		go func(metric collector.StoredMetric) {
			defer wg.Done()

			jsonInput, err := json.Marshal(collector.MetricRequest{
				ID:    metric.ID,
				MType: metric.MType,
				Delta: metric.CounterValue,
				Value: metric.GaugeValue,
			})
			if err != nil {
				a.log.Errorf("Error marshaling MetricRequest: %v", err)
				return
			}
			if a.params.Key != "" {
				dst := sha256.Sum256(jsonInput)
				req.SetHeader("HashSHA256", fmt.Sprintf("%x", dst))
			}

			message := string(jsonInput)
			if a.cryptoKey != nil {
				encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, a.cryptoKey, jsonInput)
				if err != nil {
					a.log.Errorf("Error encrypting message with public key: %v", err)
					return
				}
				message = string(encryptedData)
			}

			if err := a.sendRequestsWithRetries(req, message); err != nil {
				a.log.Errorf("Error sending agent request for counter metric: %v", err)
			}
		}(v)
	}

	wg.Wait()
	return nil
}

// SetRealIPFromRequest - метод для извлечения реального IP-адреса из заголовка запроса и сохранения его.
func (a *Agent) SetRealIPFromRequest(req *resty.Request) {
	realIP := req.Header.Get("X-Real-IP")
	a.RealIP = realIP
}

// sendRequestsWithRetries — метод, реализующий логику отправки запроса с повторами.
func (a *Agent) sendRequestsWithRetries(req *resty.Request, jsonInput string) error {
	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	if _, err := zb.Write([]byte(jsonInput)); err != nil {
		return fmt.Errorf("error while write json input: %w", err)
	}
	if err := zb.Close(); err != nil {
		return fmt.Errorf("error while trying to close writer: %w", err)
	}

	if err := retry.Do(func() error {
		if _, err := req.SetBody(buf).Post(fmt.Sprintf("http://%s/update/", a.params.FlagRunAddr)); err != nil {
			return fmt.Errorf("error while trying to create post request: %w", err)
		}
		return nil
	}, retry.Attempts(10), retry.OnRetry(func(n uint, err error) {
		log.Printf("Retrying request after error: %v", err)
	})); err != nil {
		return fmt.Errorf("error while trying to connect to server: %w", err)
	}
	return nil
}

// New - функция для создания нового экземпляра Agent.
func New(params *flags.Params, storage *metrics.Storage, log *zap.SugaredLogger) (*Agent, error) {
	agent := &Agent{
		params:  params,
		storage: storage,
		log:     log,
		client:  resty.New(),
	}
	if params.CryptoKeyPath != "" {
		b, err := os.ReadFile(params.CryptoKeyPath)
		if err != nil {
			return nil, fmt.Errorf("error while reading file with crypto public key: %w", err)
		}
		block, _ := pem.Decode(b)
		if block == nil {
			return nil, fmt.Errorf("failed to decode PEM block")
		}
		publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("error parsing public key: %w", err)
		}
		agent.cryptoKey = publicKey.(*rsa.PublicKey)
	}
	return agent, nil
}

// Agent - структура, представляющая агента.
type Agent struct {
	params            *flags.Params
	storage           *metrics.Storage
	cryptoKey         *rsa.PublicKey
	log               *zap.SugaredLogger
	client            *resty.Client
	grpcMetricsClient pb.MetricsClient
	RealIP            string // Добавляем поле для хранения реального IP-адреса клиента
}
