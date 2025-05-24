package main

import (
	"net/http"
	"strconv"
	"strings"
)

func MetricUpdateHandler(store Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		url := strings.TrimPrefix(r.URL.Path, "/update/")
		stringSlice := strings.Split(url, "/")
		if len(stringSlice) != 3 {
			http.Error(w, "Not Found", http.StatusNotFound)
			return
		}

		metricType := stringSlice[0]
		metricName := stringSlice[1]
		metricValue := stringSlice[2]

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
