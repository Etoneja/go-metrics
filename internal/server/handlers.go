package server

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/go-chi/chi/v5"
)

func MetricUpdateHandler(store Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		helper := &storageHelper{store: store}

		err := helper.addMetric(metricType, metricName, metricValue)
		if err != nil {
			if errors.Is(err, errValidation) {
				http.Error(w, err.Error(), http.StatusBadRequest)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	}

}

func MetricGetHandler(store Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		helper := &storageHelper{store: store}

		metricValue, err := helper.getMetric(metricType, metricName)
		if err != nil {
			if errors.Is(err, errValidation) {
				http.Error(w, err.Error(), http.StatusBadRequest)
			} else if errors.Is(err, errNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
			} else {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%v", metricValue)

	}
}

func MetricListHandler(store Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		helper := &storageHelper{store: store}

		metrics, err := helper.listMetrics()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		keys := make([]string, 0, len(metrics))
		for k := range metrics {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintln(w, "<html><body><pre>")

		for _, k := range keys {
			v := metrics[k]
			fmt.Fprintf(w, "%s=%s\n", k, common.AnyToString(v))
		}

		fmt.Fprintln(w, "</pre></body></html>")

	}
}
