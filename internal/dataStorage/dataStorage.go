package dataStorage

// Metric metric  структура, представляющая метрику.
type Metric struct {
	Value      interface{} // Значение метрики
	MetricType string      // Тип метрики
}

// MemStorage - структура, представляющая хранилище памяти для метрик.
type MemStorage struct {
	Metrics map[string]Metric //// Мапа метрик, где ключ - строковый идентификатор, значение - метрика
}
