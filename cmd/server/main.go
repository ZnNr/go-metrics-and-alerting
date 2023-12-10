package main

import (
	"github.com/ZnNr/go-musthave-metrics.git/internal/metricshandlers"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	r := chi.NewRouter() // Создаем новый маршрутизатор с помощью chi.NewRouter()

	// Определяем маршрут для POST запроса на обновление метрики.
	// Шаблон "/update/*" обозначает, что после "/update/" может быть произвольный путь.
	r.Post("/update/*", metricshandlers.SaveMetric)

	// Определяем маршрут для GET запроса на получение значения метрики.
	// Шаблон "/value/*" обозначает, что после "/value/" может быть произвольный путь.
	r.Get("/value/*", metricshandlers.GetMetric)

	// Определяем маршрут для GET запроса на отображение всех метрик.
	// Шаблон "/" обозначает корневой путь.
	r.Get("/", metricshandlers.ShowMetrics)

	// Запускаем сервер на порту 8080 и передаем ему созданный маршрутизатор r.
	//log.Fatal используется для логирования и завершения программы в случае возникновения критической ошибки.
	log.Fatal(http.ListenAndServe(":8080", r))

}
