package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(store Storager) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Post("/update/{metricType}/{metricName}/{metricValue}", MetricUpdateHandler(store))

	return r
}
