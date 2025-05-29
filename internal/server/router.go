package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(store Storager) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", MetricListHandler(store))
	r.Post("/update/{metricType}/{metricName}/{metricValue}", MetricUpdateHandler(store))
	r.Get("/value/{metricType}/{metricName}", MetricGetHandler(store))

	return r
}
