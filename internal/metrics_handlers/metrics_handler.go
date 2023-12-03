package metrics_handlers

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Объявляем тип metric, который представляет метрику с полем value произвольного типа и полем mType типа string:
type metric struct {
	value any
	mType string
}

// Определяем тип MemStorage как структуру с полем metric типа map[string]metric, которое будет служить как коллекция для хранения метрик
type MemStorage struct {
	metrics map[string]metric
}

// Создаем переменную m типа MemStorage
var m MemStorage

// SaveMetric Функция обрабатывает POST-запросы для сохранения метрик на сервере
func SaveMetric(w http.ResponseWriter, r *http.Request) {
	// Функция проверяет метод запроса и возвращает код ошибки 405 (http.StatusMethodNotAllowed), если метод не соответствует POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	//Если массив metric пустой, инициализируем его с помощью make
	if len(m.metrics) == 0 {
		m.metrics = make(map[string]metric, 0)
	}
	// Разбиваем URL на компоненты, используя "/" в качестве разделителя
	url := strings.Split(r.URL.String(), "/")
	// Проверяем, что URL содержит ожидаемое количество компонентов (равное 5) и возвращаем код ошибки 404 (http.StatusNotFound), если это не так
	if len(url) != 5 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	// Проверка пути запроса
	if url[1] != "update" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// В зависимости от второго компонента URL выбираем тип метрики (counter или gauge)
	switch url[2] {
	// Тип counter, int64 — новое значение должно добавляться к предыдущему, если какое-то значение уже было известно серверу.
	case "counter":
		//Для метрик типа counter преобразуем строковое значение четвертого компонента URL в целочисленное значение с помощью strconv.Atoi()
		value, err := strconv.Atoi(url[4])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m.metrics[url[2]] = metric{value, url[2]}
		// Тип gauge, float64 — новое значение должно замещать предыдущее
	case "gauge":
		//Для метрик типа gauge преобразуем строковое значение четвертого компонента URL в число с плавающей запятой (тип float64) с помощью strconv.ParseFloat()
		value, err := strconv.ParseFloat(url[4], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		m.metrics[url[2]] = metric{value, url[2]}
	default:
		//Если тип метрики неизвестен, возвращаем код ошибки 501 (http.StatusNotImplemented)
		w.WriteHeader(http.StatusNotImplemented)
		return
	}
	//Записываем пустую строку в ответ (io.WriteString(w, "")),
	//устанавливаем заголовки Content-Type и Content-Length для ответа,
	//возвращаем код успешного выполнения (http.StatusOK)
	io.WriteString(w, "")
	w.Header().Set("content-type", "text/plain; charset=utf-8")
	w.Header().Set("content-lenght", strconv.Itoa(len(url[3])))
	w.WriteHeader(http.StatusOK)
	// Выводим значения метрик и URL для отладки
	fmt.Println(m.metrics)
	fmt.Println(r.URL)
}
