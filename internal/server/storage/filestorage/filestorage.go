package filestorage

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aksenk/go-yandex-metrics/internal/models"
	"github.com/aksenk/go-yandex-metrics/internal/server/storage/memstorage"
	"go.uber.org/zap"
	"os"
	"sync"
)

type FileStorage struct {
	*memstorage.MemStorage
	FileName         string
	File             *os.File
	Writer           *bufio.Writer
	SynchronousFlush bool
	FileLock         *sync.Mutex
	Logger           *zap.SugaredLogger
}

func NewFileStorage(filename string, synchronousFlush bool, logger *zap.SugaredLogger) (*FileStorage, error) {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		logger.Errorf("FileStorage.NewFileStorage: can not open file '%v': %v'", filename, err)
		return nil, fmt.Errorf("FileStorage.NewFileStorage: can not open file '%v': %v'", filename, err)
	}
	memStorage := memstorage.NewMemStorage(logger)
	return &FileStorage{
		MemStorage:       memStorage,
		File:             file,
		FileName:         filename,
		Writer:           bufio.NewWriter(file),
		SynchronousFlush: synchronousFlush,
		FileLock:         &sync.Mutex{},
		Logger:           logger,
	}, nil
}

func (f *FileStorage) SaveMetric(ctx context.Context, metric models.Metric) error {
	err := f.MemStorage.SaveMetric(ctx, metric)
	if err != nil {
		return err
	}
	if f.SynchronousFlush {
		f.FlushMetrics()
	}
	return nil
}

func (f *FileStorage) SaveBatchMetrics(ctx context.Context, metrics []models.Metric) error {
	for _, metric := range metrics {
		err := f.MemStorage.SaveMetric(ctx, metric)
		if err != nil {
			return err
		}
	}
	if f.SynchronousFlush {
		f.FlushMetrics()
	}
	return nil
}

func (f *FileStorage) StartupRestore(ctx context.Context) error {
	counter := 0
	f.Logger.Infof("Restoring metrics from a file '%v'", f.FileName)
	// saving current state of SynchronousFlush because we need to disable it for startup restoring
	sf := f.SynchronousFlush
	f.SynchronousFlush = false
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
		f.Logger.Debugf("Proccessing line: %v", string(line))
		if err := json.Unmarshal(line, &metric); err != nil {
			f.Logger.Errorf("FileStorage.restoreMetrics: can not unmarshal metric. Error: %v. Line: %v", err, line)
			return fmt.Errorf("FileStorage.restoreMetrics: can not unmarshal metric. Error: %v. Line: %v", err, line)
		}
		if err := f.SaveMetric(ctx, *metric); err != nil {
			f.Logger.Errorf("FileStorage.restoreMetrics: can not restore metric '%v': %v", metric, err)
			return fmt.Errorf("FileStorage.restoreMetrics: can not restore metric '%v': %v", metric, err)
		}
	}
	if scanner.Err() != nil {
		f.Logger.Errorf("FileStorage.restoreMetrics: can not restore metrics from the FileName: %v", scanner.Err())
		return fmt.Errorf("ileStorage.restoreMetrics: can not restore metrics from the FileName: %v", scanner.Err())
	}
	f.Logger.Infof("Successfully restored %v metrics from a file", counter)
	// restoring state of SynchronousFlush
	f.SynchronousFlush = sf
	return nil
}

func (f *FileStorage) FlushMetrics() error {
	counter := 0
	f.Logger.Debug("Start collecting metrics for flushing to the file")
	for _, v := range f.Metrics {
		jsonMetric, err := json.Marshal(v)
		if err != nil {
			f.Logger.Errorf("Сan not marsgal metric '%v': %v", v, err)
			return fmt.Errorf("FileStorage.FlushMetrics: can not marshal metric '%v': %v", v, err)
		}
		jsonMetric = append(jsonMetric, '\n')
		_, err = f.Writer.Write(jsonMetric)
		if err != nil {
			f.Logger.Errorf("Сan not write metric '%v' to the file: %v", v, err)
			return fmt.Errorf("FileStorage.FlushMetrics: can not write metric '%v' to the file: %v", v, err)
		}
		counter++
	}
	f.Logger.Debugf("Start saving %v metrics to the file", counter)
	f.FileLock.Lock()
	err := f.File.Truncate(0)
	if err != nil {
		f.Logger.Errorf("Сan not truncate file '%v': %v", f.FileName, err)
		return err
	}
	err = f.Writer.Flush()
	if err != nil {
		f.Logger.Errorf("Сan not flush file '%v': %v", f.FileName, err)
		return err
	}
	f.FileLock.Unlock()
	f.Logger.Info("Metrics successfully saved")
	return nil
}

func (f *FileStorage) Close() error {
	return f.File.Close()
}

func (f *FileStorage) Status(ctx context.Context) error {
	return nil
}
