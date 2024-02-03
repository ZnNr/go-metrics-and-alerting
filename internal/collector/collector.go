package collector

import (
	"encoding/json"
	"errors"
	"strconv"
)

var (
	// ErrBadRequest представляет ошибку для некорректного запроса
	ErrBadRequest = errors.New("bad request")
	// ErrNotImplemented представляет ошибку для не реализованной функциональности
	ErrNotImplemented = errors.New("not implemented")
	// ErrNotFound  представляет ошибку для не найденных данных.
	ErrNotFound = errors.New("not found")
)

// Collector Определен экземпляр структуры collector с именем Collector
var Collector = collector{
	Metrics: make([]StoredMetric, 0),
}

// Collect добавляет собранную метрику в коллектор
func (c *collector) Collect(metric MetricRequest, metricValue string) error {
	if (metric.Delta != nil && *metric.Delta < 0) || (metric.Value != nil && *metric.Value < 0) || metric.ID == "" {
		return ErrBadRequest
	}

	switch metric.MType {
	case Counter:
		v, err := c.GetMetric(metric.ID)
		if err != nil {
			if !errors.Is(err, ErrNotFound) {
				return err
			}
		}
		value, err := strconv.Atoi(metricValue)
		if err != nil {
			return ErrBadRequest
		}
		if v.CounterValue != nil {
			value = value + int(*v.CounterValue)
		}
		metricToStore := StoredMetric{
			ID:           metric.ID,
			MType:        metric.MType,
			CounterValue: PtrInt64(int64(value)),
			TextValue:    PtrString(strconv.Itoa(value)),
		}
		c.UpsertMetric(metricToStore)
	case Gauge:
		value, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return ErrBadRequest
		}
		metricToStore := StoredMetric{
			ID:         metric.ID,
			MType:      metric.MType,
			GaugeValue: &value,
			TextValue:  &metricValue,
		}
		c.UpsertMetric(metricToStore)
	default:
		return ErrNotImplemented
	}
	return nil
}

func (c *collector) GetMetricJSON(metricName string) ([]byte, error) {
	for _, m := range c.Metrics {
		if m.ID == metricName {
			resultJSON, err := json.Marshal(m)
			if err != nil {
				return nil, ErrBadRequest
			}
			return resultJSON, nil
		}
	}
	return nil, ErrNotFound
}

// GetMetric возвращает значение заданной метрики по имени метрики
func (c *collector) GetMetric(metricName string) (StoredMetric, error) {
	for _, m := range c.Metrics {
		if m.ID == metricName {
			return m, nil
		}
	}
	return StoredMetric{}, ErrNotFound
}

// GetAvailableMetrics Метод возвращает слайс с доступными метриками.
// Внутри метода перебираются элементы счетчиков и показателей в объекте "storage" и добавляются в срез.
func (c *collector) GetAvailableMetrics() []string {
	names := make([]string, 0)
	for _, m := range c.Metrics {
		names = append(names, m.ID)
	}
	return names
}

// UpsertMetric добавляет или обновляет метрику в коллекторе.
func (c *collector) UpsertMetric(metric StoredMetric) {
	for i, m := range c.Metrics {
		if m.ID == metric.ID {
			c.Metrics[i] = metric
			return
		}
	}
	c.Metrics = append(c.Metrics, metric)
}

// PtrFloat64 создает указатель на float64 с заданным значением.
func PtrFloat64(f float64) *float64 {
	return &f
}

// PtrInt64 создает указатель на int64 с заданным значением.
func PtrInt64(i int64) *int64 {
	return &i
}

// PtrString создает указатель на строку с заданным значением.
func PtrString(s string) *string {
	return &s
}
