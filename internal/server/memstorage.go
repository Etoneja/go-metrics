package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"maps"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
)

type MemStorage struct {
	mu *sync.RWMutex

	filePath           string
	syncDump           bool
	dumpInProgress     atomic.Bool
	shutdownInProgress atomic.Bool
	stopChan           chan struct{}
	doneChan           chan struct{}

	gauge   map[string]float64
	counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		mu:       &sync.RWMutex{},
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
		gauge:    make(map[string]float64),
		counter:  make(map[string]int64),
	}
}

type StorageConfig struct {
	StoreInterval   uint
	FileStoragePath string
	Restore         bool
}

func NewMemStorageFromStorageConfig(sc *StorageConfig) *MemStorage {
	ms := NewMemStorage()

	ms.syncDump = sc.StoreInterval == 0
	ms.filePath = sc.FileStoragePath

	if sc.Restore {
		err := ms.load()
		if err != nil {
			log.Printf("Error occurred: %v", err)
		}
	}

	if !ms.syncDump {
		ms.startPeriodicDump(sc.StoreInterval)
	}

	return ms
}

func (ms *MemStorage) GetGauge(ctx context.Context, key string) (float64, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		return 0, err
	}

	val, ok := ms.gauge[key]
	if !ok {
		return 0, fmt.Errorf("%s %s: %w", common.MetricTypeGauge, key, ErrNotFound)
	}
	return val, nil
}

func (ms *MemStorage) SetGauge(ctx context.Context, key string, value float64) (float64, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return 0, err
	}

	prevValue, ok := ms.gauge[key]

	ms.gauge[key] = value

	if ms.syncDump {
		err := ms.dump()
		if err != nil {
			if ok {
				ms.gauge[key] = prevValue
			} else {
				delete(ms.gauge, key)
			}
			return 0, err
		}
	}

	return value, nil
}

func (ms *MemStorage) GetCounter(ctx context.Context, key string) (int64, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		return 0, err
	}

	val, ok := ms.counter[key]
	if !ok {
		return 0, fmt.Errorf("%s %s: %w", common.MetricTypeCounter, key, ErrNotFound)
	}
	return val, nil
}

func (ms *MemStorage) IncrementCounter(ctx context.Context, key string, value int64) (int64, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return 0, err
	}

	val, ok := ms.counter[key]

	if ok {
		value += val
	}
	ms.counter[key] = value

	if ms.syncDump {
		err := ms.dump()
		if err != nil {
			if ok {
				ms.counter[key] = val
			} else {
				delete(ms.counter, key)
			}
			return 0, err
		}
	}

	return value, nil
}

func (ms *MemStorage) GetAll(ctx context.Context) ([]models.MetricModel, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	return ms.getAll(), nil
}

func (ms *MemStorage) getAll() []models.MetricModel {
	metrics := make([]models.MetricModel, 0, len(ms.gauge)+len(ms.counter))
	for k, v := range ms.gauge {
		metrics = append(metrics, *models.NewMetricModel(k, common.MetricTypeGauge, 0, v))
	}
	for k, v := range ms.counter {
		metrics = append(metrics, *models.NewMetricModel(k, common.MetricTypeCounter, v, 0))
	}
	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].ID < metrics[j].ID
	})

	return metrics
}

func (ms *MemStorage) Dump() error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	return ms.dump()
}

func (ms *MemStorage) dump() error {
	if !ms.dumpInProgress.CompareAndSwap(false, true) {
		return fmt.Errorf("dump already in progress")
	}

	defer func() {
		ms.dumpInProgress.Store(false)
	}()

	metrics := ms.getAll()

	tmpPath := ms.filePath + ".tmp"
	file, err := os.Create(tmpPath)

	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			log.Printf("warning: failed to close temp file: %v", closeErr)
		}

		if err != nil {
			if removeErr := os.Remove(tmpPath); removeErr != nil && !os.IsNotExist(removeErr) {
				log.Printf("warning: failed to remove temp file %s: %v", tmpPath, removeErr)
			}
		}
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	err = encoder.Encode(metrics)
	if err != nil {
		return fmt.Errorf("failed to encode metrics: %w", err)
	}

	err = file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	err = os.Rename(tmpPath, ms.filePath)
	if err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	log.Println("Data dumped successfully")

	return nil
}

func (ms *MemStorage) load() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	file, err := os.Open(ms.filePath)
	if err != nil {
		return err
	}
	defer func() {
		if err = file.Close(); err != nil {
			log.Printf("failed to close file: %v", err)
		}
	}()

	decoder := json.NewDecoder(file)

	var metrics []models.MetricModel
	err = decoder.Decode(&metrics)
	if err != nil {
		return err
	}

	log.Printf("Load %d entries", len(metrics))

	for _, m := range metrics {
		switch m.MType {
		case common.MetricTypeGauge:
			ms.gauge[m.ID] = *m.Value
		case common.MetricTypeCounter:
			ms.counter[m.ID] = *m.Delta
		default:
			return fmt.Errorf("unknown metric type %s", m.MType)
		}
	}

	log.Printf("Load %d gauges, %d counters", len(ms.gauge), len(ms.counter))

	return nil
}

func (ms *MemStorage) startPeriodicDump(period uint) {
	dur := time.Second * time.Duration(period)
	ticker := time.NewTicker(dur)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				err := ms.Dump()
				if err != nil {
					log.Printf("Error occurred: %v", err)
				}
			case <-ms.stopChan:
				defer close(ms.doneChan)
				err := ms.Dump()
				if err != nil {
					log.Printf("Error occurred: %v", err)
				}
				return
			}

		}
	}()

}

func (ms *MemStorage) ShutDown() {
	if ms.shutdownInProgress.Load() {
		return
	}

	log.Println("MemStorage shutdowning...")
	if ms.syncDump {
		for ms.dumpInProgress.Load() {
			log.Println("Dump in progress. Waiting...")
			time.Sleep(1 * time.Second)
		}
	} else {
		close(ms.stopChan)
		<-ms.doneChan
	}
	log.Println("MemStorage shutdowned")
}

func (ms *MemStorage) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func (ms *MemStorage) BatchUpdate(ctx context.Context, metrics []models.MetricModel) ([]models.MetricModel, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	backupCounters := make(map[string]int64, len(ms.counter))
	maps.Copy(backupCounters, ms.counter)

	backupGauges := make(map[string]float64, len(ms.gauge))
	maps.Copy(backupGauges, ms.gauge)

	newMetrics := make([]models.MetricModel, 0, len(metrics))

	var err error
	for _, m := range metrics {
		switch m.MType {
		case common.MetricTypeCounter:
			val, ok := ms.counter[m.ID]
			if ok {
				val += *m.Delta
			} else {
				val = *m.Delta
			}
			ms.counter[m.ID] = val
			newMetrics = append(newMetrics, *models.NewMetricModel(m.ID, m.MType, *m.Delta, 0))

		case common.MetricTypeGauge:
			val := *m.Value
			ms.gauge[m.ID] = val
			newMetrics = append(newMetrics, *models.NewMetricModel(m.ID, m.MType, 0, *m.Value))

		default:
			err = fmt.Errorf("bad metric type %s", m.MType)
		}

		if err != nil {
			break
		}
	}

	restoreBackup := func() {
		ms.counter = backupCounters
		ms.gauge = backupGauges
	}

	if err != nil {
		restoreBackup()
		return nil, err
	}

	if ms.syncDump {
		err := ms.dump()
		if err != nil {
			restoreBackup()
			return nil, err
		}
	}

	return newMetrics, nil
}
