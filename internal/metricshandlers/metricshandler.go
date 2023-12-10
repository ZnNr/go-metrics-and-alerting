package metricshandlers

import (
	"fmt"
	"github.com/ZnNr/go-musthave-metrics.git/internal/storage"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// SaveMetric Функция обрабатывает POST-запросы для сохранения метрик на сервере
func SaveMetric(w http.ResponseWriter, r *http.Request) {
	// Функция проверяет метод запроса и возвращает код ошибки 405 (http.StatusMethodNotAllowed), если метод не соответствует POST
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	//Если массив metric пустой, инициализируем его с помощью make
	if len(storage.MetricsStorage.Metrics) == 0 {
		storage.MetricsStorage.Metrics = make(map[string]storage.Metric, 0)
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
		if storage.MetricsStorage.Metrics[url[3]].Value != nil {
			value += storage.MetricsStorage.Metrics[url[3]].Value.(int)
		}
		storage.MetricsStorage.Metrics[url[3]] = storage.Metric{Value: value, MetricType: url[2]}
		// Тип gauge, float64 — новое значение должно замещать предыдущее
	case "gauge":
		//Для метрик типа gauge преобразуем строковое значение четвертого компонента URL в число с плавающей запятой (тип float64) с помощью strconv.ParseFloat()
		_, err := strconv.ParseFloat(url[4], 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		storage.MetricsStorage.Metrics[url[3]] = storage.Metric{Value: url[4], MetricType: url[2]}
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
	//fmt.Println(m.metrics)
	//fmt.Println(r.URL)
}

// GetMetric - функция, обрабатывающая GET запрос для получения значения метрики.
func GetMetric(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { // Проверяем метод запроса, если это не GET, возвращаем ошибку "Метод не разрешен"
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	url := strings.Split(r.URL.String(), "/") // Разделяем URL запроса по символу "/"
	if len(url) != 4 {                        // Если количество полученных частей URL не равно 4, значит запрос некорректный. Возвращаем ошибку "Не найдено"
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if url[1] != "value" { // Если вторая часть URL не равна "value", значит запрос некорректный. Возвращаем ошибку "Не найдено"
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if _, ok := storage.MetricsStorage.Metrics[url[3]]; !ok { // Проверяем, есть ли метрика с указанным идентификатором. Если нет, возвращаем ошибку "Не найдено"
		w.WriteHeader(http.StatusNotFound)
		return
	}
	value := storage.MetricsStorage.Metrics[url[3]].Value // Получаем значение метрики по указанному идентификатору

	io.WriteString(w, "")                                       // Пишем пустую строку в ответ
	w.Header().Set("content-type", "text/plain; charset=utf-8") // Устанавливаем заголовок "content-type" с типом "text/plain; charset=utf-8"
	w.Header().Set("content-length", strconv.Itoa(len(url[3]))) // Устанавливаем заголовок "content-length" с длиной идентификатора метрики
	w.WriteHeader(http.StatusOK)                                // Устанавливаем статус "200 OK"
	// В зависимости от типа значения метрики записываем его в ответ с помощью функции WriteString
	switch value.(type) {
	case uint, uint64, int, int64: // Если тип значения является числовым типом, записываем его как строку
		io.WriteString(w, strconv.Itoa(value.(int)))
	default: // Если тип значения не числовой, записываем его как строку
		io.WriteString(w, value.(string))

	}
}

// ShowMetrics - функция, обрабатывающая GET запрос для отображения всех метрик.
func ShowMetrics(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" { // Проверяем путь URL запроса, если это не корневой путь, возвращаем ошибку "Не найдено"
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	page := `
<html> 
   <head> 
   </head> 
   <body> 
`
	for n := range storage.MetricsStorage.Metrics {
		// Перебираем все метрики из хранилища page
		page += fmt.Sprintf(`<h3>%s   </h3>`, n) // Добавляем имя метрики в HTML-страницу
	}
	page += `
   </body> 
</html>
`
	w.Header().Set("content-type", "Content-Type: text/html; charset=utf-8") /// Устанавливаем заголовок "content-type" с типом "text/html; charset=utf-8"
	w.WriteHeader(http.StatusOK)                                             // Устанавливаем статус "200 OK"
	w.Write([]byte(page))                                                    //// Записываем HTML-страницу в ответ
}
