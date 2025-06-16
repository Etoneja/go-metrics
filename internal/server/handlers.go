package server

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/go-chi/chi/v5"
)

func MetricUpdateHandler(store Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")
		metricValue := chi.URLParam(r, "metricValue")

		if metricName == "" {
			http.Error(w, "Bad Request: bad metric name", http.StatusBadRequest)
			return
		}

		if metricType == common.MetricTypeGauge {
			num, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = store.SetGauge(metricName, num)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		} else if metricType == common.MetricTypeCounter {
			num, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			err = store.IncrementCounter(metricName, num)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		} else {
			http.Error(w, "Bad Request: bad metric type", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

	}

}

func MetricGetHandler(store Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if metricType == common.MetricTypeGauge {
			value, ok, err := store.GetGauge(metricName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !ok {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%v", value)
			return

		} else if metricType == common.MetricTypeCounter {
			value, ok, err := store.GetCounter(metricName)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !ok {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%v", value)
			return
		}

		http.Error(w, "Bad Request: bad metric type", http.StatusBadRequest)

	}
}

func MetricListHandler(store Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		helper := &storageHelper{store: store}

		metrics, err := helper.listMetrics()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
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
