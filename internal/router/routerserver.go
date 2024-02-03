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
	r.Use(compressor.HTTPCompressHandler)
	r.Post("/update/", handler.SaveMetricFromJSONHandler)
	r.Post("/value/", handler.GetMetricFromJSONHandler)
	r.Post("/update/{type}/{name}/{value}", handler.SaveMetricHandler)
	r.Get("/value/{type}/{name}", handler.GetMetricHandler)
	r.Get("/", handler.ShowMetricsHandler)
	r.Get("/ping", handler.CheckDatabaseAvailability)
	r.Post("/updates/", handler.SaveListMetricsFromJSONHandler)

	return r
}
