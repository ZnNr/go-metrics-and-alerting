// Package router предоставляет функционал для создания и настройки маршрутов с использованием библиотеки go-chi.
// Включает обработку HTTP запросов, сжатие данных и логирование.
package router

import (
	"fmt"
	"github.com/ZnNr/go-musthave-metrics.git/internal/flags"
	"github.com/ZnNr/go-musthave-metrics.git/internal/server/handlers"
	"github.com/ZnNr/go-musthave-metrics.git/internal/server/middlewares/compressor"
	log "github.com/ZnNr/go-musthave-metrics.git/internal/server/middlewares/logger"
	"github.com/go-chi/chi/v5"
)

// New возвращает новый экземпляр маршрутизатора с настроенными обработчиками для обработки HTTP запросов.
func New(params flags.Params) (*chi.Mux, error) {
	handler, err := handlers.New(
		params.DatabaseAddress,
		params.Key,
		params.CryptoKeyPath,
		params.TrustedSubnet,
	)
	if err != nil {
		return nil, fmt.Errorf("error while creating handler: %w", err)
	}
	r := chi.NewRouter()
	r.Use(log.RequestLogger)
	r.Use(compressor.HTTPCompressHandler)
	r.Use(handler.CheckSubscriptionHandler)
	r.Use(handler.CheckSubnetHandler)
	r.Post("/update/", handler.SaveMetricFromJSONHandler)
	r.Post("/value/", handler.GetMetricFromJSONHandler)
	r.Post("/update/{type}/{name}/{value}", handler.SaveMetricHandler)
	r.Get("/value/{type}/{name}", handler.GetMetricHandler)
	r.Get("/", handler.ShowMetricsHandler)
	r.Get("/ping", handler.CheckDatabaseAvailability)
	r.Post("/updates/", handler.SaveListMetricsFromJSONHandler)

	return r, nil
}
