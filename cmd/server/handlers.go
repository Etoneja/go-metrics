package main

import (
	"github.com/go-chi/chi/v5"
	"net/http"
	"strconv"
)

func MetricUpdateHandler(store Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		if metricType == metricTypeGauge {
			num, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, "Bad Metric Value", http.StatusBadRequest)
				return
			}
			err = store.SetGauge(metricName, num)
			if err != nil {
				http.Error(w, "Internal Error", http.StatusInternalServerError)
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		if metricType == metricTypeCounter {
			num, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, "Bad Metric Value", http.StatusBadRequest)
				return
			}
			err = store.IncrementCounter(metricName, num)
			if err != nil {
				http.Error(w, "Internal Error", http.StatusInternalServerError)
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		http.Error(w, "Bad Metric Type", http.StatusBadRequest)

	}

}
