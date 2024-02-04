package filestorage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/logger"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/memstorage"
	"os"
	"sync"
)

type FileStorage struct {
	*memstorage.MemStorage
	FileName         string
	File             *os.File
	Writer           *bufio.Writer
	synchronousFlush bool
	FileLock         *sync.Mutex
}

func NewFileStorage(filename string, synchronousFlush bool) (*FileStorage, error) {
	log := logger.Log
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		log.Errorf("FileStorage.NewFileStorage: can not open file '%v': %v'", filename, err)
		return nil, fmt.Errorf("FileStorage.NewFileStorage: can not open file '%v': %v'", filename, err)
	}
	memStorage := memstorage.NewMemStorage()
	return &FileStorage{
		MemStorage:       memStorage,
		File:             file,
		FileName:         filename,
		Writer:           bufio.NewWriter(file),
		synchronousFlush: synchronousFlush,
		FileLock:         &sync.Mutex{},
	}, nil
}

func (f *FileStorage) SaveMetric(metric models.Metric) error {
	err := f.MemStorage.SaveMetric(metric)
	if err != nil {
		return err
	}
	if f.synchronousFlush {
		f.FlushMetrics()
	}
	return nil
}

func (f *FileStorage) StartupRestore() error {
	log := logger.Log
	counter := 0
	log.Infof("Restoring metrics from a file '%v'", f.FileName)
	// saving current state of synchronousFlush because we need to disable it for startup restoring
	sf := f.synchronousFlush
	f.synchronousFlush = false
	file, err := os.OpenFile(f.FileName, os.O_RDONLY|os.O_CREATE, 0660)
	if err != nil {
		return fmt.Errorf("can not openfile: %v", err)
	}
	defer file.Close()
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
		log.Errorf("FileStorage.restoreMetrics: can not restore metrics from the FileName: %v", scanner.Err())
		return fmt.Errorf("ileStorage.restoreMetrics: can not restore metrics from the FileName: %v", scanner.Err())
	}
	log.Infof("Successfully restored %v metrics from a file", counter)
	// restoring state of synchronousFlush
	f.synchronousFlush = sf
	return nil
}

func (f *FileStorage) FlushMetrics() error {
	log := logger.Log
	counter := 0
	log.Debugf("Start collecting metrics for flushing to the file")
	for _, v := range f.Metrics {
		jsonMetric, err := json.Marshal(v)
		if err != nil {
			log.Errorf("Сan not marsgal metric '%v': %v", v, err)
			return fmt.Errorf("FileStorage.FlushMetrics: can not marshal metric '%v': %v", v, err)
		}
		jsonMetric = append(jsonMetric, '\n')
		_, err = f.Writer.Write(jsonMetric)
		if err != nil {
			log.Errorf("Сan not write metric '%v' to the file: %v", v, err)
			return fmt.Errorf("FileStorage.FlushMetrics: can not write metric '%v' to the file: %v", v, err)
		}
		counter++
	}
	log.Infof("Start saving %v metrics to the file", counter)
	f.FileLock.Lock()
	f.File.Truncate(0)
	f.Writer.Flush()
	f.FileLock.Unlock()
	log.Infof("Metrics successfully saved")
	return nil
}
