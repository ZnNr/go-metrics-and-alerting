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
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	"github.com/ZnNr/go-musthave-metrics.git/internal/storage"
	"github.com/avast/retry-go"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"log"
	"os"
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

// SendMetrics — метод отправки метрик на сервер по таймеру
func (a *Agent) SendMetrics(ctx context.Context) (err error) {
	numRequests := make(chan struct{}, a.params.RateLimit)
	reportTicker := time.NewTicker(time.Duration(a.params.ReportInterval) * time.Second)
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			a.log.Warn("Sending of metrics stopped")
			return nil
		// проверяем, пришло ли время отправлять метрики на сервер
		case <-reportTicker.C:
			select {
			case <-ctx.Done():
				a.log.Info("reportTicker.C triggered")
				return nil
			// проверяем, превышен ли лимит
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

// sendMetrics — метод, инкапсулирующий логику отправки http-запроса на сервер.
func (a *Agent) sendMetrics(client *resty.Client) error {
	req := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Encoding", "gzip")

	for _, v := range collector.Collector().Metrics {
		jsonInput, _ := json.Marshal(collector.MetricRequest{
			ID:    v.ID,
			MType: v.MType,
			Delta: v.CounterValue,
			Value: v.GaugeValue,
		})
		if a.params.Key != "" {

			dst := sha256.Sum256(jsonInput)
			req.SetHeader("HashSHA256", fmt.Sprintf("%x", dst))
		}
		message := string(jsonInput)
		if a.cryptoKey != nil {
			encryptedData, err := rsa.EncryptPKCS1v15(rand.Reader, a.cryptoKey, jsonInput)
			if err != nil {
				return fmt.Errorf("error encrypting message with public key: %w", err)
			}
			message = string(encryptedData)
		}
		if err := a.sendRequestsWithRetries(req, message); err != nil {
			return fmt.Errorf("error while sending agent request for counter metric: %w", err)
		}
	}
	return nil
}

// sendMetrics — метод, реализующий логику отправки запроса с повторами.
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
func New(params *flags.Params, storage *storage.Storage, log *zap.SugaredLogger) (*Agent, error) {
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
	params    *flags.Params
	storage   *storage.Storage
	cryptoKey *rsa.PublicKey
	log       *zap.SugaredLogger
	client    *resty.Client
}
