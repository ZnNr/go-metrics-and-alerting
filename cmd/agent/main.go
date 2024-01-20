package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	"github.com/ZnNr/go-musthave-metrics.git/internal/storage"
	"github.com/avast/retry-go"
	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"
	"log"
	"time"
)

func main() {
	//Инициализируются параметры программы, используя пакет flags.
	params := flags.Init(
		flags.WithPollInterval(),
		flags.WithReportInterval(),
		flags.WithAddr())

	ctx := context.Background()

	errs, _ := errgroup.WithContext(ctx)
	errs.Go(func() error {
		st := storage.New(&collector.Collector)
		for {
			st.Store()
			time.Sleep(time.Duration(params.PollInterval) * time.Second)
		}
	})

	client := resty.New()
	errs.Go(func() error {
		if err := send(client, params.ReportInterval, params.FlagRunAddr); err != nil {
			log.Fatalln(err)
		}
		return nil
	})

	_ = errs.Wait()
}

// Функция send() отправляет метрики на удаленный сервер.
func send(client *resty.Client, reportTimeout int, addr string) error {
	req := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Accept-Encoding", "gzip").
		SetHeader("Content-Encoding", "gzip")

	for {
		for _, v := range collector.Collector.Metrics {
			jsonInput, _ := json.Marshal(v)
			if err := sendRequest(req, string(jsonInput), addr); err != nil {
				return fmt.Errorf("error while sending agent request for counter metric: %w", err)
			}
		}
		time.Sleep(time.Duration(reportTimeout) * time.Second)
	}
}

func sendRequest(req *resty.Request, jsonInput string, addr string) error {
	buf := bytes.NewBuffer(nil)
	zb := gzip.NewWriter(buf)
	if _, err := zb.Write([]byte(jsonInput)); err != nil {
		return fmt.Errorf("error while write json input: %w", err)
	}
	if err := zb.Close(); err != nil {
		return fmt.Errorf("error while trying to close writer: %w", err)
	}

	err := retry.Do(
		func() error {
			var err error
			if _, err = req.SetBody(buf).Post(fmt.Sprintf("http://%s/update/", addr)); err != nil {
				return fmt.Errorf("error while trying to create post request: %w", err)
			}
			return nil
		},
		retry.Attempts(10),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Retrying request after error: %v", err)
		}),
	)
	if err != nil {
		return fmt.Errorf("error while trying to connect to server: %w", err)
	}
	return nil
}
