package filestorage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/memstorage"
	"os"
)

type FileStorage struct {
	fileName   *string
	writer     *bufio.Writer
	memStorage *memstorage.MemStorage
}

func NewFileStorage(filename *string) (*FileStorage, error) {
	log := logger.Log
	file, err := os.OpenFile(*filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		log.Errorf("FileStorage.NewFileStorage: can not open file '%v': %v'", filename, err)
		return nil, fmt.Errorf("FileStorage.NewFileStorage: can not open file '%v': %v'", filename, err)
	}
	memStorage := memstorage.NewMemStorage()
	return &FileStorage{
		fileName:   filename,
		writer:     bufio.NewWriter(file),
		memStorage: memStorage,
	}, nil
}

func (f *FileStorage) SaveMetric(metric models.Metric) error {
	return f.memStorage.SaveMetric(metric)
}

func (f *FileStorage) GetMetric(name string) (*models.Metric, error) {
	return f.memStorage.GetMetric(name)
}

func (f *FileStorage) GetAllMetrics() map[string]models.Metric {
	return f.memStorage.GetAllMetrics()
}

func (f *FileStorage) StartupRestore() error {
	log := logger.Log
	counter := 0
	log.Infof("Restoring metrics from a file '%v'", *f.fileName)
	file, err := os.OpenFile(*f.fileName, os.O_RDONLY|os.O_CREATE, 0660)
	if err != nil {
		return fmt.Errorf("can not openfile: %v", err)
	}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var metric *models.Metric
		counter++
		line := scanner.Bytes()
		log.Debugf("Proccessing line: %v", string(line))
		if err := json.Unmarshal(line, &metric); err != nil {
			log.Errorf("FileStorage.restoreMetrics: can not unmarshal metric. Error: %v. Line: %v", err, line)
			return fmt.Errorf("FileStorage.restoreMetrics: can not unmarshal metric. Error: %v. Line: %v", err, line)
		}
		if err := f.SaveMetric(*metric); err != nil {
			log.Errorf("FileStorage.restoreMetrics: can not restore metric '%v': %v", metric, err)
			return fmt.Errorf("FileStorage.restoreMetrics: can not restore metric '%v': %v", metric, err)
		}
	}
	if scanner.Err() != nil {
		log.Errorf("FileStorage.restoreMetrics: can not restore metrics from the fileName: %v", scanner.Err())
		return fmt.Errorf("ileStorage.restoreMetrics: can not restore metrics from the fileName: %v", scanner.Err())
	}
	log.Infof("Successfully restored %v metrics from a file", counter)
	return nil
}
