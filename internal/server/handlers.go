package server

import (
	"fmt"
	"math"
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

		if metricType == common.MetricTypeGauge {
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

		if metricType == common.MetricTypeCounter {
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

func MetricGetHandler(store Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		if metricType == common.MetricTypeGauge {
			value, ok, err := store.GetGauge(metricName)
			if err != nil {
				http.Error(w, "Internal Error", http.StatusInternalServerError)
			}
			if !ok {
				http.Error(w, "Not Found", http.StatusNotFound)
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%v", value)
			return
		}

		if metricType == common.MetricTypeCounter {
			value, ok, err := store.GetCounter(metricName)
			if err != nil {
				http.Error(w, "Internal Error", http.StatusInternalServerError)
			}
			if !ok {
				http.Error(w, "Not Found", http.StatusNotFound)
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "%v", value)
			return
		}

		http.Error(w, "Not Found", http.StatusNotFound)

	}
}

func MetricListHandler(store Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gauges, err := store.ListGauges()
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}
		counters, err := store.ListCounters()
		if err != nil {
			http.Error(w, "Internal Error", http.StatusInternalServerError)
		}

		gaugesKeys := make([]string, 0, len(gauges))
		for k := range gauges {
			gaugesKeys = append(gaugesKeys, k)
		}

		sort.Strings(gaugesKeys)

		countersKeys := make([]string, 0, len(counters))
		for k := range counters {
			countersKeys = append(countersKeys, k)
		}

		sort.Strings(countersKeys)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintln(w, "<html><body><pre>")

		fmt.Fprintln(w, "<b>gauges</b>")
		for _, k := range gaugesKeys {
			v := gauges[k]
			if v == math.Trunc(v) {
				fmt.Fprintf(w, "%s=%.0f\n", k, v)
			} else {
				fmt.Fprintf(w, "%s=%.2f\n", k, v)
			}
		}

		fmt.Fprintln(w, "\n<b>counters</b>")
		for _, k := range countersKeys {
			v := counters[k]
			fmt.Fprintf(w, "%s=%d\n", k, v)
		}

		fmt.Fprintln(w, "</pre></body></html>")

	}
}
