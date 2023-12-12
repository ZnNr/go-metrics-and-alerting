package main

import (
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	"github.com/ZnNr/go-musthave-metrics.git/internal/handlers"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

func main() {
	params := flags.Init(flags.WithAddr())
	r := chi.NewRouter() // Создаем новый маршрутизатор с помощью chi.NewRouter()

	// Определяем маршрут для POST запроса на обновление метрики.
	//{name} имя метрики  {value} новое значение
	r.Post("/update/name/value", handlers.SaveMetric)

	// Определяем маршрут для GET запроса на получение значения метрики.
	//{name} имя метрики
	r.Get("/value/name", handlers.GetMetric)

	// Определяем маршрут для GET запроса на отображение всех метрик.
	// Шаблон "/" обозначает корневой путь.
	r.Get("/", handlers.ShowMetrics)

	// Запускаем сервер на порту 8080 и передаем ему созданный маршрутизатор r.
	//log.Fatal используется для логирования и завершения программы в случае возникновения критической ошибки.
	log.Fatal(http.ListenAndServe(params.FlagRunAddr, r))

}
