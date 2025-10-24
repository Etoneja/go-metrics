package server

import (
	"net/http"

	"github.com/etoneja/go-metrics/internal/logger"
	"github.com/go-chi/chi/v5"
)

func NewRouter(store Storager, cfg *config) http.Handler {
	r := chi.NewRouter()

	lg := logger.Get()

	bmw := BaseMiddleware{logger: lg}

	r.Use(bmw.LoggerMiddleware())
	r.Use(bmw.TrustedIPMiddleware(cfg.TrustedSubnet))
	r.Use(bmw.DecryptMiddleware(cfg.GetPrivateKey()))
	r.Use(bmw.HashMiddleware(cfg.HashKey))
	r.Use(bmw.GzipMiddleware())

	bh := BaseHandler{store: store, logger: lg}

	r.Get("/", bh.MetricListHandler())
	r.Post("/update/{metricType}/{metricName}/{metricValue}", bh.MetricUpdateHandler())
	r.Post("/update/", bh.MetricUpdateJSONHandler())
	r.Post("/updates/", bh.MetricBatchUpdateJSONHandler())
	r.Get("/value/{metricType}/{metricName}", bh.MetricGetHandler())
	r.Post("/value/", bh.MetricGetJSONHandler())
	r.Get("/ping", bh.PingHandler())

	return r
}
