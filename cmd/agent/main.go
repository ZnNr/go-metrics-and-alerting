package main

import (
	"context"
	"fmt"
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	"github.com/ZnNr/go-musthave-metrics.git/internal/storage"
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
// В бесконечном цикле происходит отправка POST-запросов для обновления значений счетчиков и метрик на удаленном адресе.
// В каждой итерации цикла происходит обращение к пакету collector для получения текущих значений счетчиков и метрик.
// Затем выполняется отправка POST-запросов с использованием клиента resty.
// После отправки всех метрик горутина "спит" на определенное время, заданное в параметрах (report interval).
func send(client *resty.Client, reportTimeout int, addr string) error {
	for {
		for n, v := range collector.Collector.GetCounters() {
			if _, err := client.R().
				SetHeader("Content-Type", "text/plain").
				Post(fmt.Sprintf("http://%s/update/counter/%s/%s", addr, n, v)); err != nil {
				return err
			}
		}
		for n, v := range collector.Collector.GetGauges() {
			if _, err := client.R().
				SetHeader("Content-Type", "text/plain").
				Post(fmt.Sprintf("http://%s/update/gauge/%s/%s", addr, n, v)); err != nil {
				return err
			}
		}
		time.Sleep(time.Duration(reportTimeout) * time.Second)
	}
}
