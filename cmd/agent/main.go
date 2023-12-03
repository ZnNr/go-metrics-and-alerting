package main

import (
	"context"
	"fmt"
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/dataStorage"
	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"
	"runtime"
	"time"
)

// создает переменную m типа dataStorage.MemStorage и инициализирует поле Metrics как пустую мапу (map[string]dataStorage.Metric{}).
// Таким образом, переменная m представляет собой хранилище метрик в памяти.
var m = dataStorage.MemStorage{Metrics: map[string]dataStorage.Metric{}}

func main() {
	//Создается объект metricsCollector типа collector, принимающий указатель на переменную m в качестве аргумента. Этот объект отвечает за сбор метрик.
	metricsCollector := collector.New(&m)
	//Создается контекст ctx с помощью функции context.Background(). Контекст используется для управления жизненным циклом операций в приложении.
	ctx := context.Background()
	//Создается периодический таймер mtick с интервалом в 2 секунды. Этот таймер будет использоваться в функции performCollect() для регулярного сбора метрик.
	mtick := time.NewTicker(time.Second * 2)
	defer mtick.Stop()
	//Запускается группа ошибок errs с использованием функции errgroup.WithContext(). Данная группа позволяет координировать выполнение нескольких операций одновременно и обрабатывать ошибки.
	errs, ctx := errgroup.WithContext(ctx)
	// Внутри группы ошибок запускается анонимная функция с использованием метода Go(). В этой функции вызывается функция performCollect(), которая принимает контекст ctx, таймер mtick и объект metricsCollector. Если во время выполнения функции произойдет ошибка, она будет обрабатываться с помощью функции panic().
	errs.Go(func() error {
		if err := performCollect(ctx, mtick, metricsCollector); err != nil {
			panic(err)
		}
		return nil
	})
	//Создается периодический таймер stick с интервалом в 10 секунд. Этот таймер будет использоваться в функции Send() для отправки данных.
	stick := time.NewTicker(time.Second * 10)
	//Создается объект client типа resty.Client, который будет использоваться для отправки HTTP-запросов.
	client := resty.New()
	defer stick.Stop()
	//Запускается вторая анонимная функция внутри группы ошибок, которая вызывает функцию Send() с использованием контекста ctx, таймера stick и объекта client. Если произойдет ошибка во время выполнения функции, она также будет обработана с помощью функции panic().
	errs.Go(func() error {
		if err := Send(ctx, stick, client); err != nil {
			panic(err)
		}
		return nil
	})
	//Вызывается метод Wait() для группы ошибок. Этот метод блокирует выполнение программы до тех пор, пока все операции не завершатся.
	_ = errs.Wait()
}

// ICollector представляет собой интерфейс для сборщика метрик
type IСollector interface {
	CollectMetrics(metrics *runtime.MemStats)
}

// performCollect() - это функция, которая выполняет сбор метрик в регулярных интервалах с использованием контекста, таймера и объекта metricsCollector.
// функция находится в постоянном цикле
// - ctx - контекст, управляющий жизненным циклом выполнения функции.
// - ticker - периодический таймер, определяющий интервалы сбора метрик.
// - metricsCollector - объект, ответственный за сбор и хранение метрик.
func performCollect(ctx context.Context, ticker *time.Ticker, metricsCollector IСollector) error {
	for {
		select {
		//Если контекст ctx завершается (например, по таймауту или при получении сигнала завершения), функция завершает свою работу и возвращает ошибку, полученную из контекста.
		case <-ctx.Done():
			return ctx.Err()
			//Если происходит срабатывание таймера ticker.C
		case <-ticker.C:
			//Обновляет метрики памяти с помощью функции runtime.ReadMemStats(), записывая их в объект metrics.
			metrics := runtime.MemStats{}
			runtime.ReadMemStats(&metrics)
			//Передает объект metrics в метод CollectMetrics() объекта metricsCollector, который выполняет сбор и хранение метрик
			metricsCollector.CollectMetrics(&metrics)
		}
	}
}

// Функция Send() принимает аргументы ctx (контекст выполнения), ticker (периодический таймер) и client (клиент REST API)
func Send(ctx context.Context, ticker *time.Ticker, client *resty.Client) error {
	//В бесконечном цикле for функцция ожидает событий от двух каналов
	for {
		select {
		//ctx.Done(): Если выполнение контекста ctx завершено (например, по истечению времени или при получении сигнала завершения), функция возвращает ошибку из контекста и завершает свою работу
		case <-ctx.Done():
			return ctx.Err()
			//ticker.C: Если происходит срабатывание таймера ticker.C то
		case <-ticker.C:
			//Функция проходит по каждому элементу в структуре m.Metrics
			for n, i := range m.Metrics {
				//Функция проверяет тип значения (value) метрики.
				switch i.Value.(type) {
				//Если это uint или uint64,
				case uint, uint64:
					resp, err := client.R().
						SetHeader("Content-Type", "text/plain").
						//Функция выполняет POST-запрос на указанный URL, передавая значение метрики в формате числа целого типа.
						Post(fmt.Sprintf("http://localhost:8080/update/%s/%s/%d", i.MetricType, n, i.Value))
					if err != nil {
						return err
					}
					fmt.Println(resp.Status())
					//Если значение метрики является float64
				case float64:
					resp, err := client.R().
						SetHeader("Content-Type", "text/plain").
						//функция также выполняет POST-запрос на указанный URL, передавая значение метрики в формате числа с плавающей точкой.
						Post(fmt.Sprintf("http://localhost:8080/update/%s/%s/%f", i.MetricType, n, i.Value))
					if err != nil {
						return err
					}
					//После каждого POST-запроса, функция выводит статус ответа (resp.Status()) на стандартный вывод
					fmt.Println(resp.Status())
				}
			}
		}
	}
}
