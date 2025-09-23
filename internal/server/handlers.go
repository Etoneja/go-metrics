package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

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

		ctx := r.Context()

		if metricType == common.MetricTypeGauge {
			num, err := strconv.ParseFloat(metricValue, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			_, err = bh.store.SetGauge(ctx, metricName, num)
			if err != nil {
				bh.logger.Error("failed to set metric",
					zap.String("metricType", metricType),
					zap.String("metricName", metricName),
					zap.Error(err),
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

		} else if metricType == common.MetricTypeCounter {
			num, err := strconv.ParseInt(metricValue, 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			_, err = bh.store.IncrementCounter(ctx, metricName, num)
			if err != nil {
				bh.logger.Error("failed to set metric",
					zap.String("metricType", metricType),
					zap.String("metricName", metricName),
					zap.Error(err),
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

		ctx := r.Context()

		switch metricType {
		case common.MetricTypeGauge:
			value, err = bh.store.GetGauge(ctx, metricName)
		case common.MetricTypeCounter:
			value, err = bh.store.GetCounter(ctx, metricName)
		default:
			http.Error(w, "Bad Request: bad metric type", http.StatusBadRequest)
			return
		}
		if errors.Is(err, ErrNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		if err != nil {
			bh.logger.Error("failed to get metric",
				zap.String("metricType", metricType),
				zap.String("metricName", metricName),
				zap.Error(err),
			)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%v", value)
	}
}

func (bh *BaseHandler) MetricListHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		metrics, err := bh.store.GetAll(ctx)
		if err != nil {
			bh.logger.Error("failed to get metrics",
				zap.Error(err),
			)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

		ctx := r.Context()

		if metricModelRequest.MType == common.MetricTypeGauge {
			if metricModelRequest.Value == nil {
				http.Error(w, "Bad Request: missing value", http.StatusBadRequest)
				return
			}

			newValue, err := bh.store.SetGauge(ctx, metricModelRequest.ID, *metricModelRequest.Value)
			if err != nil {
				bh.logger.Error("failed to get metric",
					zap.String("metricType", metricModelRequest.MType),
					zap.String("metricName", metricModelRequest.ID),
					zap.Error(err),
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			metricModelResponse = *models.NewMetricModel(metricModelRequest.ID, metricModelRequest.MType, 0, newValue)

		} else if metricModelRequest.MType == common.MetricTypeCounter {
			if metricModelRequest.Delta == nil {
				http.Error(w, "Bad Request: missing delta", http.StatusBadRequest)
				return
			}

			newValue, err := bh.store.IncrementCounter(ctx, metricModelRequest.ID, *metricModelRequest.Delta)
			if err != nil {
				bh.logger.Error("failed to get metric",
					zap.String("metricType", metricModelRequest.MType),
					zap.String("metricName", metricModelRequest.ID),
					zap.Error(err),
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			metricModelResponse = *models.NewMetricModel(metricModelRequest.ID, metricModelRequest.MType, newValue, 0)

		} else {
			http.Error(w, "Bad Request: bad metric type", http.StatusBadRequest)
			return
		}

		resp, err := json.Marshal(metricModelResponse)
		if err != nil {
			bh.logger.Error("failed to marshal response",
				zap.Error(err),
			)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

		ctx := r.Context()

		if metricGetRequestModel.MType == common.MetricTypeGauge {
			value, err := bh.store.GetGauge(ctx, metricGetRequestModel.ID)
			if errors.Is(err, ErrNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			if err != nil {
				bh.logger.Error("failed to get metric",
					zap.String("metricType", metricGetRequestModel.MType),
					zap.String("metricName", metricGetRequestModel.ID),
					zap.Error(err),
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			metricModel = *models.NewMetricModel(metricGetRequestModel.ID, metricGetRequestModel.MType, 0, value)

		} else if metricGetRequestModel.MType == common.MetricTypeCounter {
			value, err := bh.store.GetCounter(ctx, metricGetRequestModel.ID)
			if errors.Is(err, ErrNotFound) {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			if err != nil {
				bh.logger.Error("failed to get metric",
					zap.String("metricType", metricGetRequestModel.MType),
					zap.String("metricName", metricGetRequestModel.ID),
					zap.Error(err),
				)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			metricModel = *models.NewMetricModel(metricGetRequestModel.ID, metricGetRequestModel.MType, value, 0)

		} else {
			http.Error(w, "Bad Request: bad metric type", http.StatusBadRequest)
			return
		}

		resp, err := json.Marshal(metricModel)
		if err != nil {
			bh.logger.Error("failed to marshal response",
				zap.Error(err),
			)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)

	}

}

func (bh *BaseHandler) PingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		err := bh.store.Ping(ctx)
		if err != nil {
			bh.logger.Error("failed to ping store",
				zap.Error(err),
			)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)

	}

}

// MetricBatchUpdateJSONHandler creates an HTTP handler for batch updating metrics in JSON format.
//
// The handler processes POST requests to /updates/ with a JSON array of metrics in the request body:
//
//	[
//	  {
//	    "id": "string",         // metric identifier
//	    "type": "gauge|counter", // metric type
//	    "value": number,        // value for gauge metrics
//	    "delta": number         // value for counter metrics
//	  }
//	]
//
// Responses:
//   - 200 OK: Returns the updated metrics array
//   - 400 Bad Request: Invalid JSON format or malformed data
//   - 500 Internal Server Error: Server-side processing error
//
// Example request:
//
//	curl -X POST http://localhost:8080/updates/ \
//	  -H "Content-Type: application/json" \
//	  -d '[{"id":"cpu_usage","type":"gauge","value":95.5}]'
//
// Example response:
//
//	[
//	  {
//	    "id": "cpu_usage",
//	    "type": "gauge",
//	    "value": 95.5
//	  }
//	]
func (bh *BaseHandler) MetricBatchUpdateJSONHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var metricModelsRequest []models.MetricModel

		if err := json.NewDecoder(r.Body).Decode(&metricModelsRequest); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		newMetrics, err := bh.store.BatchUpdate(ctx, metricModelsRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resp, err := json.Marshal(newMetrics)
		if err != nil {
			bh.logger.Error("failed to marshal response",
				zap.Error(err),
			)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)

	}

}
