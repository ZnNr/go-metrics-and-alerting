package agent

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	"github.com/ZnNr/go-musthave-metrics.git/internal/storage"
	"github.com/avast/retry-go"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
	"log"
	"time"
)

func (a *Agent) CollectMetrics(ctx context.Context) (err error) {
	storeTicker := time.NewTicker(time.Duration(a.params.PollInterval) * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				err = ctx.Err()
				return
			case <-storeTicker.C:
				a.storage.RuntimeMetricStore()
				a.storage.GopsutilMetricStore()
			}
		}
	}()
	return err
}

func (a *Agent) SendMetrics(ctx context.Context) error {
	numRequests := make(chan struct{}, a.params.RateLimit)
	reportTicker := time.NewTicker(time.Duration(a.params.ReportInterval) * time.Second)
	client := resty.New()
	for {
		select {
		case <-ctx.Done():
			return nil
		// проверяем, пришло ли время отправлять метрики на сервер
		case <-reportTicker.C:
			select {
			case <-ctx.Done():
				return nil
			// проверяем, превышен ли лимит
			case numRequests <- struct{}{}:
				go func() {
					defer func() { <-numRequests }()
					a.log.Info("metrics sent on server")
					if err := a.sendMetrics(client); err != nil {
						log.Printf("error while sending metrics: %v", err)
					}
				}()
			default:
				a.log.Info("rate limit is exceeded")
			}
		}
	}
}

func (a *Agent) sendMetrics(client *resty.Client) error {
	req := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Encoding", "gzip")

	for _, v := range collector.Collector.Metrics {
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
		if err := a.sendRequestsWithRetries(req, string(jsonInput)); err != nil {
			return fmt.Errorf("error while sending agent request for counter metric: %w", err)
		}
	}
	return nil
}

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

func New(params *flags.Params, storage *storage.Storage, log zap.SugaredLogger) *Agent {
	return &Agent{
		params:  params,
		storage: storage,
		log:     log,
	}
}

type Agent struct {
	params  *flags.Params
	storage *storage.Storage
	log     zap.SugaredLogger
}
