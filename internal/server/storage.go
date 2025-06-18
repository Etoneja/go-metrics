package server

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/etoneja/go-metrics/internal/common"
	"github.com/etoneja/go-metrics/internal/models"
)

type MemStorage struct {
	mu sync.RWMutex

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

	syncDump := false
	if sc.StoreInterval == 0 {
		syncDump = true
	}
	ms.syncDump = syncDump
	ms.filePath = sc.FileStoragePath

	if sc.Restore {
		err := ms.load()
		if err != nil {
			log.Printf("Error occurred: %v", err)
		}
	}

	if !syncDump {
		ms.startPeriodicDump(sc.StoreInterval)
	}

	return ms
}

func (ms *MemStorage) GetGauge(key string) (float64, bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	val, ok := ms.gauge[key]
	return val, ok, nil
}

func (ms *MemStorage) SetGauge(key string, value float64) (float64, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

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

func (ms *MemStorage) GetCounter(key string) (int64, bool, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	val, ok := ms.counter[key]
	return val, ok, nil
}

func (ms *MemStorage) IncrementCounter(key string, value int64) (int64, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

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

func (ms *MemStorage) GetAll() *[]models.MetricModel {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.getAll()
}

func (ms *MemStorage) getAll() *[]models.MetricModel {
	metrics := make([]models.MetricModel, 0, len(ms.gauge)+len(ms.counter))
	for k, v := range ms.gauge {
		metrics = append(metrics, *models.NewMetricModel(k, common.MetricTypeGauge, 0, v))
	}
	for k, v := range ms.counter {
		metrics = append(metrics, *models.NewMetricModel(k, common.MetricTypeCounter, v, 0))
	}
	return &metrics
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
		file.Close()
		if err != nil {
			os.Remove(tmpPath) 
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
	defer file.Close()

	decoder := json.NewDecoder(file)

	var metrics []models.MetricModel
	err = decoder.Decode(&metrics)
	if err != nil {
		return err
	}

	log.Printf("Load %d entries", len(metrics))

	for _, m := range metrics {
		if m.MType == common.MetricTypeCounter {
			ms.counter[m.ID] = *m.Delta
		} else if m.MType == common.MetricTypeGauge {
			ms.gauge[m.ID] = *m.Value
		} else {
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
