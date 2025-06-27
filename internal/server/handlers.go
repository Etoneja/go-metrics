package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type BaseHandler struct {
	store  Storager
	logger *zap.Logger
}

func (bh *BaseHandler) MetricUpdateHandler() http.HandlerFunc {
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
			_, err = bh.store.SetGauge(metricName, num)
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
			_, err = bh.store.IncrementCounter(metricName, num)
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

func (bh *BaseHandler) MetricGetHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metricType := chi.URLParam(r, "metricType")
		metricName := chi.URLParam(r, "metricName")

		var value any
		var err error
		var ok bool

		switch metricType {
		case common.MetricTypeGauge:
			value, ok, err = bh.store.GetGauge(metricName)
		case common.MetricTypeCounter:
			value, ok, err = bh.store.GetCounter(metricName)
		default:
			http.Error(w, "Bad Request: bad metric type", http.StatusBadRequest)
			return
		}
		if !ok {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%v", value)
	}
}

func (bh *BaseHandler) MetricListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		metrics, err := bh.store.GetAll()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintln(w, "<html><body><pre>")

		for _, m := range metrics {
			var value string
			if m.MType == common.MetricTypeCounter {
				value = common.AnyToString(*m.Delta)
			} else {
				value = common.AnyToString(*m.Value)
			}
			fmt.Fprintf(w, "%s[%s]=%s\n", m.ID, m.MType, value)
		}

		fmt.Fprintln(w, "</pre></body></html>")

	}
}

func (bh *BaseHandler) MetricUpdateJSONHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var metricModelRequest models.MetricModel
		var metricModelResponse models.MetricModel

		if err := json.NewDecoder(r.Body).Decode(&metricModelRequest); err != nil {
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

			newValue, err := bh.store.SetGauge(metricModelRequest.ID, *metricModelRequest.Value)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			metricModelResponse = *models.NewMetricModel(metricModelRequest.ID, metricModelRequest.MType, 0, newValue)

		} else if metricModelRequest.MType == common.MetricTypeCounter {
			if metricModelRequest.Delta == nil {
				http.Error(w, "Bad Request: missing delta", http.StatusBadRequest)
				return
			}

			newValue, err := bh.store.IncrementCounter(metricModelRequest.ID, *metricModelRequest.Delta)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			metricModelResponse = *models.NewMetricModel(metricModelRequest.ID, metricModelRequest.MType, newValue, 0)

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

func (bh *BaseHandler) MetricGetJSONHandler() http.HandlerFunc {
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
			value, ok, err := bh.store.GetGauge(metricGetRequestModel.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !ok {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}

			metricModel = *models.NewMetricModel(metricGetRequestModel.ID, metricGetRequestModel.MType, 0, value)

		} else if metricGetRequestModel.MType == common.MetricTypeCounter {
			value, ok, err := bh.store.GetCounter(metricGetRequestModel.ID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if !ok {
				http.Error(w, "Not found", http.StatusNotFound)
				return
			}

			metricModel = *models.NewMetricModel(metricGetRequestModel.ID, metricGetRequestModel.MType, value, 0)

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

func (bh *BaseHandler) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err := bh.store.Ping(ctx)
		if err != nil {
			log.Printf("failed to ping store %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	}

}

func (bh *BaseHandler) MetricBatchUpdateJSONHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var metricModelsRequest []models.MetricModel

		if err := json.NewDecoder(r.Body).Decode(&metricModelsRequest); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		newMetrics, err := bh.store.BatchUpdate(metricModelsRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := json.Marshal(newMetrics)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)

	}

}
