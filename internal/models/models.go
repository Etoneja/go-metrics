package models

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
