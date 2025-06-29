package models

import (
	"encoding/json"
	"fmt"

	"github.com/etoneja/go-metrics/internal/common"
)

type MetricModel struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *int64   `json:"delta,omitempty"`
	Value *float64 `json:"value,omitempty"`
}
type MetricGetRequestModel struct {
	ID    string `json:"id"`
	MType string `json:"type"`
}

func NewMetricModel(id string, mtype string, delta int64, value float64) *MetricModel {
	return &MetricModel{
		ID:    id,
		MType: mtype,
		Delta: &delta,
		Value: &value,
	}
}

func (m MetricModel) MarshalJSON() ([]byte, error) {
	type Alias MetricModel

	switch m.MType {
	case common.MetricTypeCounter:
		m.Value = nil
	case common.MetricTypeGauge:
		m.Delta = nil
	}

	return json.Marshal(Alias(m))

}

func (m *MetricModel) UnmarshalJSON(data []byte) error {
	type Alias MetricModel
	aux := (*Alias)(m)

	err := json.Unmarshal(data, &aux)
	if err != nil {
		return err
	}

	switch m.MType {
	case common.MetricTypeCounter:
		if aux.Delta == nil {
			return fmt.Errorf("bad metric delta for metric: metrics=%s, type=%s", m.ID, m.MType)
		}
		m.Delta = aux.Delta
		m.Value = nil
	case common.MetricTypeGauge:
		if aux.Value == nil {
			return fmt.Errorf("bad metric value for metric: metrics=%s, type=%s", m.ID, m.MType)
		}
		m.Value = aux.Value
		m.Delta = nil
	default:
		return fmt.Errorf("unknown metric type for metric: metrics=%s, type=%s", m.ID, m.MType)
	}

	return nil
}
