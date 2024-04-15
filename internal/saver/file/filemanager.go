// Package file предоставляет функционал для сохранения и восстановления состояния метрик из файла.
package file

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/ZnNr/go-musthave-metrics.git/internal/collector"
	"os"
)

// Restore восстанавливает состояние метрик из файла.
func (m *manager) Restore(ctx context.Context) ([]collector.StoredMetric, error) {
	file, err := os.OpenFile(m.fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return nil, scanner.Err()
	}

	data := scanner.Bytes()
	var metricsFromFile []collector.StoredMetric
	if err := json.Unmarshal(data, &metricsFromFile); err != nil {
		return nil, err
	}
	return metricsFromFile, nil
}

// Save сохраняет состояние метрик в файл.
func (m *manager) Save(ctx context.Context, metrics []collector.StoredMetric) error {
	file, err := os.OpenFile(m.fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			err = cerr
		}
	}()

	writer := bufio.NewWriter(file)

	data, err := json.Marshal(&metrics)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	if err != nil {
		return err
	}

	if err = writer.WriteByte('\n'); err != nil {
		return err
	}
	if err = writer.Flush(); err != nil {
		return err
	}
	return nil
}

// New создает новый менеджер для работы с файлами.
func New(path string) *manager {
	return &manager{fileName: path}
}

type manager struct {
	fileName string
}
