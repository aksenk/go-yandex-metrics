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
	memstorage.MemStorage
	roFIle     *os.File
	rwFile     *os.File
	writer     *bufio.Writer
	scanner    *bufio.Scanner
	memStorage *memstorage.MemStorage
}

func NewFileStorage(filename string) (*FileStorage, error) {
	log := logger.Log
	readOnlyFile, err := os.OpenFile(filename, os.O_CREATE|os.O_RDONLY, 0666)
	if err != nil {
		log.Errorf("FileStorage.NewFileStorage: can not open file '%v': %v'", filename, err)
		return nil, fmt.Errorf("FileStorage.NewFileStorage: can not open file '%v': %v'", filename, err)
	}
	writeOnlyFile, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Errorf("FileStorage.NewFileStorage: can not open file '%v': %v'", filename, err)
		return nil, fmt.Errorf("FileStorage.NewFileStorage: can not open file '%v': %v'", filename, err)
	}
	memStorage := memstorage.NewMemStorage()
	return &FileStorage{
		roFIle:     readOnlyFile,
		rwFile:     writeOnlyFile,
		writer:     bufio.NewWriter(writeOnlyFile),
		scanner:    bufio.NewScanner(readOnlyFile),
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

func (f *FileStorage) FlushMetrics() error {
	log := logger.Log
	counter := 0
	log.Infof("Start collecting metrics for flushing to the file")
	for _, v := range f.memStorage.Metrics {
		jsonMetric, err := json.Marshal(v)
		if err != nil {
			log.Errorf("FileStorage.FlushMetrics: can not marsgal metric '%v': %v", v, err)
			return fmt.Errorf("FileStorage.FlushMetrics: can not marshal metric '%v': %v", v, err)
		}
		_, err = f.writer.Write(jsonMetric)
		if err != nil {
			log.Errorf("FileStorage.FlushMetrics: can not write metric '%v' to the file: %v", v, err)
			return fmt.Errorf("FileStorage.FlushMetrics: can not write metric '%v' to the file: %v", v, err)
		}
		err = f.writer.WriteByte('\n')
		if err != nil {
			log.Errorf("FileStorage.FlushMetrics: can not write metric '%v' to the file: %v", v, err)
			return fmt.Errorf("FileStorage.FlushMetrics: can not write metric '%v' to the file: %v", v, err)
		}
		counter++
	}
	f.writer.Flush()
	log.Infof("Successfully flushed %v metrics to the file", counter)
	return nil
}

func (f *FileStorage) RestoreMetrics() error {
	log := logger.Log
	counter := 0
	var metric *models.Metric
	log.Infof("Restoring metrics from the file '%v'", f.roFIle)
	for f.scanner.Scan() {
		counter++
		line := f.scanner.Bytes()
		if err := json.Unmarshal(line, metric); err != nil {
			log.Errorf("FileStorage.RestoreMetrics: can not unmarshal metric. Error: %v. Line: %v", err, line)
			return fmt.Errorf("FileStorage.RestoreMetrics: can not unmarshal metric. Error: %v. Line: %v", err, line)
		}
		if err := f.SaveMetric(*metric); err != nil {
			log.Errorf("FileStorage.RestoreMetrics: can not restore metric '%v': %v", metric, err)
			return fmt.Errorf("FileStorage.RestoreMetrics: can not restore metric '%v': %v", metric, err)
		}
	}
	if f.scanner.Err() != nil {
		log.Errorf("FileStorage.RestoreMetrics: can not restore metrics from the file: %v", f.scanner.Err())
		return fmt.Errorf("ileStorage.RestoreMetrics: can not restore metrics from the file: %v", f.scanner.Err())
	}
	log.Infof("Successfully restored %v metrics from the file")
	return nil
}
