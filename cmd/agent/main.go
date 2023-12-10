package main

import (
	"context"
	"fmt"
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/storage"
	"github.com/go-resty/resty/v2"
	"golang.org/x/sync/errgroup"
	"runtime"
	"time"
)

// создает переменную m типа storage.MemStorage и инициализирует поле Metrics как пустую мапу (map[string]storage.Metric{}).
// Таким образом, переменная m представляет собой хранилище метрик в памяти.
//var m = storage.MemStorage{Metrics: map[string]storage.Metric{}}

func main() {
	//Создается объект metricsCollector типа collector, принимающий указатель на переменную storage.MetricsStorage в качестве аргумента.
	//Этот объект отвечает за сбор метрик.
	metricsCollector := collector.New(&storage.MetricsStorage)
	//Создается контекст ctx с помощью функции context.Background(). Контекст используется для управления жизненным циклом операций в приложении.
	ctx := context.Background()
	//Создается периодический таймер mtick с интервалом в 2 секунды. Этот таймер будет использоваться в функции performCollect() для регулярного сбора метрик.

	errs, _ := errgroup.WithContext(ctx)
	// Внутри группы ошибок запускается анонимная функция с использованием метода Go(). В этой функции вызывается функция performCollect(), которая принимает объект metricsCollector. Если во время выполнения функции произойдет ошибка, она будет обрабатываться с помощью функции panic().
	errs.Go(func() error {
		if err := performCollect(metricsCollector); err != nil {
			panic(err)
		}
		return nil
	})
	//Создается периодический таймер stick с интервалом в 10 секунд.
	stick := time.NewTicker(time.Second * 10)
	//Создается объект client типа resty.Client, который будет использоваться для отправки HTTP-запросов.
	client := resty.New()
	defer stick.Stop()
	//Запускается вторая анонимная функция внутри группы ошибок, которая вызывает функцию Send() с использованием  объекта client. Если произойдет ошибка во время выполнения функции, она также будет обработана с помощью функции panic().
	errs.Go(func() error {
		if err := Send(client); err != nil {
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
// - metricsCollector - объект, ответственный за сбор и хранение метрик.
func performCollect(metricsCollector IСollector) error {
	for {
		metrics := runtime.MemStats{}
		runtime.ReadMemStats(&metrics)
		metricsCollector.CollectMetrics(&metrics)
		time.Sleep(time.Second * 2)

	}
}

// Функция Send() принимает аргументы client (клиент REST API)
func Send(client *resty.Client) error {
	//В бесконечном цикле for функцция ожидает событий от двух каналов
	for {
		// Перебираем все метрики в хранилище.
		for n, i := range storage.MetricsStorage.Metrics {
			// Используем switch для определения типа значения метрики.
			switch i.Value.(type) {
			case uint, uint64, int, int64: // Если тип значения метрики является целочисленным, отправляем POST запрос на сервер.
				// Используем resty.Client для создания HTTP запроса.
				// Устанавливаем заголовок "Content-Type" со значением "text/plain".
				// Используем strconv для преобразования значения метрики в строку и форматируем URL запроса с помощью fmt.Sprintf.
				// Отправляем запрос на URL "http://localhost:8080/update/<тип_метрики>/<имя_метрики>/<значение_метрики>".
				_, err := client.R().
					SetHeader("Content-Type", "text/plain").
					Post(fmt.Sprintf("http://localhost:8080/update/%s/%s/%d", i.MetricType, n, i.Value))
				if err != nil {
					return err // Если произошла ошибка при отправке запроса, возвращаем ошибку.
				}
			case float64: // Если тип значения метрики является вещественным числом, выполняем аналогичные действия для отправки запроса,
				// но преобразуем значение метрики в строку с помощью формата "%f".
				_, err := client.R().
					SetHeader("Content-Type", "text/plain").
					Post(fmt.Sprintf("http://localhost:8080/update/%s/%s/%f", i.MetricType, n, i.Value))
				if err != nil {
					return err // Если произошла ошибка при отправке запроса, возвращаем ошибку.
				}
			}
		}
		time.Sleep(time.Second * 10) // Приостанавливаем выполнение функции на 10 секунд перед следующей итерацией цикла.
	}
}
