package agent

import "fmt"

func NewMetricClient(cfg Configer) (MetricClienter, error) {
	switch cfg.GetServerProtocol() {
	case "grpc":
		return NewGRPCMetricClient(cfg)
	case "http":
		return NewHTTPMetricClient(cfg), nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", cfg.GetServerProtocol())
	}
}
