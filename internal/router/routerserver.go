package router

import (
	"github.com/ZnNr/go-musthave-metrics.git/internal/compressor"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	"github.com/ZnNr/go-musthave-metrics.git/internal/handlers"
	log "github.com/ZnNr/go-musthave-metrics.git/internal/logger"
	"github.com/go-chi/chi/v5"
)

func New(params flags.Params) *chi.Mux {
	handler := handlers.New(params.DatabaseAddress)

	r := chi.NewRouter()
	r.Use(log.RequestLogger)
	r.Use(compressor.Compress)
	r.Post("/update/", handler.SaveMetricFromJSON)
	r.Post("/value/", handler.GetMetricFromJSON)
	r.Post("/update/{type}/{name}/{value}", handler.SaveMetric)
	r.Get("/value/{type}/{name}", handler.GetMetric)
	r.Get("/", handler.ShowMetrics)
	r.Get("/ping", handler.Ping)
	r.Post("/updates/", handler.SaveListMetricsFromJSON)

	return r
}
