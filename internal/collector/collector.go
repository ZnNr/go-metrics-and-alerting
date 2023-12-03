package collector

import (
	"github.com/ZnNr/go-musthave-metrics.git/internal/dataStorage"
	"math/rand"
	"runtime"
)

// CollectMetrics функция собирает метрики и сохраняет их в хранилище
// Каждая метрика имеет числовое значение (Value) и тип (MetricType), который в данном случае определен как "gauge".
// Тип "gauge" означает, что значение метрики может возрастать или убывать со временем.
func (c *collector) CollectMetrics(metrics *runtime.MemStats) {
	//Alloc: Представляет количество выделенных объектов кучи в байтах.
	c.storage.Metrics["Alloc"] = dataStorage.Metric{Value: metrics.Alloc, MetricType: "gauge"}
	//BuckHashSys: Представляет количество выделенной памяти для таблицы хеширования передачи.
	c.storage.Metrics["BuckHashSys"] = dataStorage.Metric{Value: metrics.BuckHashSys, MetricType: "gauge"}
	//GCCPUFraction: Представляет долю времени CPU, используемую сборщиком мусора.
	c.storage.Metrics["GCCPUFraction"] = dataStorage.Metric{Value: metrics.GCCPUFraction, MetricType: "gauge"}
	//GCSys: Представляет количество выделенной памяти для системных структур сборки мусора.
	c.storage.Metrics["GCSys"] = dataStorage.Metric{Value: metrics.GCSys, MetricType: "gauge"}
	//HeapAlloc: Представляет количество выделенных объектов кучи в байтах.
	c.storage.Metrics["HeapAlloc"] = dataStorage.Metric{Value: metrics.HeapAlloc, MetricType: "gauge"}
	//HeapIdle: Представляет количество неиспользуемых спанов в байтах.
	c.storage.Metrics["HeapIdle"] = dataStorage.Metric{Value: metrics.HeapIdle, MetricType: "gauge"}
	//HeapInuse: Представляет количество используемых спанов в байтах.
	c.storage.Metrics["HeapInuse"] = dataStorage.Metric{Value: metrics.HeapInuse, MetricType: "gauge"}
	//HeapObjects: Представляет количество выделенных объектов.
	c.storage.Metrics["HeapObjects"] = dataStorage.Metric{Value: metrics.HeapObjects, MetricType: "gauge"}
	//HeapReleased: Представляет количество физической памяти, возвращенной ОС.
	c.storage.Metrics["HeapReleased"] = dataStorage.Metric{Value: metrics.HeapReleased, MetricType: "gauge"}
	//HeapSys: Представляет количество памяти, полученной ОС.
	c.storage.Metrics["HeapSys"] = dataStorage.Metric{Value: metrics.HeapSys, MetricType: "gauge"}
	//Lookups: Представляет количество операций поиска указателей, выполненных runtime'ом.
	c.storage.Metrics["Lookups"] = dataStorage.Metric{Value: metrics.Lookups, MetricType: "gauge"}
	//MCacheInuse: Представляет количество выделенной памяти для структур mcache.
	c.storage.Metrics["MCacheInuse"] = dataStorage.Metric{Value: metrics.MCacheInuse, MetricType: "gauge"}
	//MCacheSys: Представляет количество используемой памяти для структур mcache.
	c.storage.Metrics["MCacheSys"] = dataStorage.Metric{Value: metrics.MCacheSys, MetricType: "gauge"}
	//MSpanInuse: Представляет количество выделенной памяти для структур mspan.
	c.storage.Metrics["MSpanInuse"] = dataStorage.Metric{Value: metrics.MSpanInuse, MetricType: "gauge"}
	//MSpanSys: Представляет количество используемой памяти для структур mspan.
	c.storage.Metrics["MSpanSys"] = dataStorage.Metric{Value: metrics.MSpanSys, MetricType: "gauge"}
	//Mallocs: Представляет накопительный счетчик выделенных объектов кучи.
	c.storage.Metrics["Mallocs"] = dataStorage.Metric{Value: metrics.Mallocs, MetricType: "gauge"}
	//NextGC: Представляет целевой размер кучи для следующего цикла GC.
	c.storage.Metrics["NextGC"] = dataStorage.Metric{Value: metrics.NextGC, MetricType: "gauge"}
	//NumForcedGC: Представляет количество принудительных циклов GC.
	c.storage.Metrics["NumForcedGC"] = dataStorage.Metric{Value: metrics.NumForcedGC, MetricType: "gauge"}
	//NumGC: Представляет количество завершенных циклов GC.
	c.storage.Metrics["NumGC"] = dataStorage.Metric{Value: metrics.NumGC, MetricType: "gauge"}
	//OtherSys: Представляет количество выделенной памяти для других системных структур.
	c.storage.Metrics["OtherSys"] = dataStorage.Metric{Value: metrics.OtherSys, MetricType: "gauge"}
	//PauseTotalNs: Представляет общее время приостановки процесса сборки мусора в наносекундах.
	c.storage.Metrics["PauseTotalNs"] = dataStorage.Metric{Value: metrics.PauseTotalNs, MetricType: "gauge"}
	//StackInuse: Представляет количество используемой памяти стека в байтах.
	c.storage.Metrics["StackInuse"] = dataStorage.Metric{Value: metrics.StackInuse, MetricType: "gauge"}
	//StackSys: Представляет количество выделенной памяти для стека горутин в байтах.
	c.storage.Metrics["StackSys"] = dataStorage.Metric{Value: metrics.StackSys, MetricType: "gauge"}
	//Sys: Представляет общее количество выделенной памяти для системных структур.
	c.storage.Metrics["Sys"] = dataStorage.Metric{Value: metrics.Sys, MetricType: "gauge"}
	//TotalAlloc: Представляет общее количество выделенной памяти за время работы программы.
	c.storage.Metrics["TotalAlloc"] = dataStorage.Metric{Value: metrics.TotalAlloc, MetricType: "gauge"}
	//RandomValue: Создает случайное значение с помощью функции rand.Int() и сохраняет его в виде метрики типа "gauge". Для тестирования или отладки.
	c.storage.Metrics["RandomValue"] = dataStorage.Metric{Value: rand.Int(), MetricType: "gauge"}
	var cnt int64
	// Проверяем, есть ли уже значение для метрики "PollCount" в хранилище
	if c.storage.Metrics["PollCount"].Value != nil {
		// Если значение уже существует, то увеличиваем его на 1
		cnt = c.storage.Metrics["PollCount"].Value.(int64) + 1
	}
	//"PollCount" используется для отслеживания количества опросов или запросов в системе, где значение метрики увеличивается на 1 каждый раз, когда выполняется определенное действие.
	// Обновляем метрику "PollCount" в хранилище с новым значением
	c.storage.Metrics["PollCount"] = dataStorage.Metric{Value: cnt, MetricType: "counter"}
}

// New  функция создания нового экземпляра коллектора с указанным хранилищем памяти.
func New(ms *dataStorage.MemStorage) *collector {
	return &collector{ms}
}

// collector - структура, представляющая коллектор.
type collector struct {
	storage *dataStorage.MemStorage
}
