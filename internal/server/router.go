package server

import (
	"net/http"

	"github.com/etoneja/go-metrics/internal/logger"
	"github.com/go-chi/chi/v5"
)

func NewRouter(store Storager) http.Handler {
	r := chi.NewRouter()

	r.Use(LoggerMiddleware(logger.Get()))
	r.Use(GzipMiddleware)

	r.Get("/", MetricListHandler(store))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", MetricUpdateHandler(store))
	r.Post("/update/", MetricUpdateJSONHandler(store))
	r.Post("/updates/", MetricBatchUpdateJSONHandler(store))
	r.Get("/value/{metricType}/{metricName}", MetricGetHandler(store))
	r.Post("/value/", MetricGetJSONHandler(store))
	r.Get("/ping", PingHandler(store))

	return r
}
