package main

import (
	"flag"
	"github.com/ZnNr/go-musthave-metrics.git/internal/handlers"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"os"
)

var flagRunAddr string

func parseFlags() {
	cnvFlags := flag.NewFlagSet("cnv", flag.ContinueOnError)
	cnvFlags.StringVar(&flagRunAddr, "a", "localhost:8080", "address and port to run server")
	err := cnvFlags.Parse(os.Args[1:])
	if err != nil {
		log.Fatal(err)
	}
	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}
}
func main() {
	parseFlags()
	r := chi.NewRouter() // Создаем новый маршрутизатор с помощью chi.NewRouter()

	// Определяем маршрут для POST запроса на обновление метрики.
	// Шаблон "/update/*" обозначает, что после "/update/" может быть произвольный путь.
	r.Post("/update/{name}/{value}", handlers.SaveMetric)

	// Определяем маршрут для GET запроса на получение значения метрики.
	// Шаблон "/value/*" обозначает, что после "/value/" может быть произвольный путь.
	r.Get("/value/{name}", handlers.GetMetric)

	// Определяем маршрут для GET запроса на отображение всех метрик.
	// Шаблон "/" обозначает корневой путь.
	r.Get("/", handlers.ShowMetrics)

	// Запускаем сервер на порту 8080 и передаем ему созданный маршрутизатор r.
	//log.Fatal используется для логирования и завершения программы в случае возникновения критической ошибки.
	log.Fatal(http.ListenAndServe(flagRunAddr, r))

}
