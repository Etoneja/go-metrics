package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
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
			_, err = store.SetGauge(metricName, num)
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
			_, err = store.IncrementCounter(metricName, num)
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

		metrics := store.GetAll()

		sort.Slice(*metrics, func(i, j int) bool {
			return (*metrics)[i].ID < (*metrics)[j].ID
		})

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintln(w, "<html><body><pre>")

		for _, m := range *metrics {
			var value string
			if m.MType == common.MetricTypeCounter {
				value = common.AnyToString(*m.Delta)
			} else {
				value = common.AnyToString(*m.Value)
			}
			fmt.Fprintf(w, "%s=%s\n", m.ID, value)
		}

		fmt.Fprintln(w, "</pre></body></html>")

	}
}

func MetricUpdateJSONHandler(store Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var metricModelRequest models.MetricModel
		var metricModelResponse models.MetricModel
		var buf bytes.Buffer

		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &metricModelRequest); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if metricModelRequest.ID == "" {
			http.Error(w, "Bad Request: bad metric name", http.StatusBadRequest)
			return
		}

		if metricModelRequest.MType == common.MetricTypeGauge {
			if metricModelRequest.Value == nil {
				http.Error(w, "Bad Request: missing value", http.StatusBadRequest)
				return
			}

			newValue, err := store.SetGauge(metricModelRequest.ID, *metricModelRequest.Value)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			metricModelResponse = models.MetricModel{
				ID:    metricModelRequest.ID,
				MType: metricModelRequest.MType,
				Value: &newValue,
			}

		} else if metricModelRequest.MType == common.MetricTypeCounter {
			if metricModelRequest.Delta == nil {
				http.Error(w, "Bad Request: missing delta", http.StatusBadRequest)
				return
			}

			newValue, err := store.IncrementCounter(metricModelRequest.ID, *metricModelRequest.Delta)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			metricModelResponse = models.MetricModel{
				ID:    metricModelRequest.ID,
				MType: metricModelRequest.MType,
				Delta: &newValue,
			}

		} else {
			http.Error(w, "Bad Request: bad metric type", http.StatusBadRequest)
			return
		}

		resp, err := json.Marshal(metricModelResponse)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)

	}

}

func MetricGetJSONHandler(store Storager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var metricGetRequestModel models.MetricGetRequestModel
		var metricModel models.MetricModel
		var buf bytes.Buffer

		_, err := buf.ReadFrom(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err = json.Unmarshal(buf.Bytes(), &metricGetRequestModel); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if metricGetRequestModel.MType == common.MetricTypeGauge {
			value, ok, err := store.GetGauge(metricGetRequestModel.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !ok {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}

			metricModel = models.MetricModel{
				ID:    metricGetRequestModel.ID,
				MType: metricGetRequestModel.MType,
				Value: &value,
			}

		} else if metricGetRequestModel.MType == common.MetricTypeCounter {
			value, ok, err := store.GetCounter(metricGetRequestModel.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !ok {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}

			metricModel = models.MetricModel{
				ID:    metricGetRequestModel.ID,
				MType: metricGetRequestModel.MType,
				Delta: &value,
			}

		} else {
			http.Error(w, "Bad Request: bad metric type", http.StatusBadRequest)
			return
		}

		resp, err := json.Marshal(metricModel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)

	}

}
