package main

import (
	"context"
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
	//Задаются интервалы опроса (poll interval) и отчетности (report interval), а также адрес удаленного сервера.
	params := flags.Init(flags.WithPollInterval(), flags.WithReportInterval(), flags.WithAddr())
	//Создается контекст для координации выполнения горутин
	ctx := context.Background()
	//Создается группа ошибок, которая позволяет координировать работу нескольких горутин и обрабатывать ошибки, произошедшие внутри них.
	errs, _ := errgroup.WithContext(ctx)
	errs.Go(func() error {
		agg := storage.New(&collector.Collector)
		for { //// Цикл для периодического сохранения метрик
			//Запускается горутина, которая периодически сохраняет метрики.
			//В каждой итерации цикла вызывается функция Store() из пакета storage,
			//чтобы сохранить текущие метрики.
			//Затем горутина "спит" на определенное время, заданное в параметрах (poll interval).
			agg.Store()
			time.Sleep(time.Duration(params.PollInterval) * time.Second)
		}
	})
	//Создается клиент resty для выполнения HTTP-запросов.
	//Затем запускается горутина, которая периодически отправляет метрики на удаленный сервер.
	//В функции send() отправляются POST-запросы счетчиков и метрик на удаленный адрес.
	client := resty.New()
	errs.Go(func() error {
		if err := send(client, params.ReportInterval, params.FlagRunAddr); err != nil {
			log.Fatalln(err)
		}
		return nil
	})

	_ = errs.Wait() //Ожидание завершения всех горутин и обработка ошибок, возникших внутри них.
}

// Функция send() отправляет метрики на удаленный сервер.
func send(client *resty.Client, reportTimeout int, addr string) error {
	for {
		for n, v := range collector.Collector.GetCounters() {
			req := client.R().
				SetHeader("Content-Type", "text/plain").
				SetBody(fmt.Sprintf(`{"id":%q, "type":"counter", "delta": %s}`, n, v))
			if err := sendRequest(req, addr); err != nil {
				return err
			}
		}
		for n, v := range collector.Collector.GetGauges() {
			req := client.R().
				SetHeader("Content-Type", "text/plain").
				SetBody(fmt.Sprintf(`{"id":%q, "type":"gauge", "value": %s}`, n, v))
			if err := sendRequest(req, addr); err != nil {
				return err
			}
		}
		time.Sleep(time.Duration(reportTimeout) * time.Second)
	}
}

func sendRequest(req *resty.Request, addr string) error {
	err := retry.Do(
		func() error {
			var err error
			_, err = req.Post(fmt.Sprintf("https://%s/update/", addr))
			return err
		},
		retry.Attempts(10),
		retry.OnRetry(func(n uint, err error) {
			log.Printf("Retrying request after error: %v", err)
		}),
	)
	if err != nil {
		return err
	}

	return nil
}
