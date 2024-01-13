package main

import (
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"github.com/ZnNr/go-musthave-metrics.git/internal/compressor"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	"github.com/ZnNr/go-musthave-metrics.git/internal/handlers"
	log "github.com/ZnNr/go-musthave-metrics.git/internal/logger"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"time"
)

func main() {
	logger, err := zap.NewDevelopment() // добавляем предустановленный логер NewDevelopment
	if err != nil {                     // вызываем панику, если ошибка
		panic(err)
	}
	defer logger.Sync()

	log.SugarLogger = *logger.Sugar()

	params := flags.Init(
		flags.WithAddr(),
		flags.WithStoreInterval(),
		flags.WithFileStoragePath(),
		flags.WithRestore(),
	)

	r := chi.NewRouter()
	r.Use(log.RequestLogger)
	r.Use(compressor.Compress)
	r.Post("/update/", handlers.SaveMetricFromJSON)
	r.Post("/value/", handlers.GetMetricFromJSON)
	r.Post("/update/{type}/{name}/{value}", handlers.SaveMetric)
	r.Get("/value/{type}/{name}", handlers.GetMetric)
	r.Get("/", handlers.ShowMetrics)
	log.SugarLogger.Infow(
		"Starting server",
		"addr", params.FlagRunAddr,
	)

	if params.Restore {
		if err := collector.Collector.Restore(params.FileStoragePath); err != nil {
			log.SugarLogger.Error(err.Error(), "restore error")
		}
	}
	if params.FileStoragePath != "" {
		go saveMetrics(params.FileStoragePath, params.StoreInterval)
	}

	if err := http.ListenAndServe(params.FlagRunAddr, r); err != nil {
		// записываем в лог ошибку, если сервер не запустился
		log.SugarLogger.Fatalw(err.Error(), "event", "start server")
	}
}
func saveMetrics(path string, interval int) {
	for {
		if err := collector.Collector.Save(path); err != nil {
			log.SugarLogger.Error(err.Error(), "save error")
		} else {
			log.SugarLogger.Info("successfully saved metrics")
		}
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
