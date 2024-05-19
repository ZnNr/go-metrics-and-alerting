// Package file предоставляет функционал для сохранения и восстановления состояния метрик из файла.
package file

import (
	"bufio"
	"context"
	"encoding/json"
	"github.com/ZnNr/go-musthave-metrics.git/internal/agent/collector"
	"os"
)

// Restore восстанавливает состояние метрик из файла.
func (m *Manager) Restore(ctx context.Context) ([]collector.StoredMetric, error) {
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
func (m *Manager) Save(ctx context.Context, metrics []collector.StoredMetric) error {
	var saveError error
	file, err := os.OpenFile(m.fileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			saveError = err
		}
	}()

	writer := bufio.NewWriter(file)

	data, err := json.Marshal(&metrics)
	if err != nil {
		return err
	}
	if _, err = writer.Write(data); err != nil {
		return err
	}
	if err = writer.WriteByte('\n'); err != nil {
		return err
	}
	if err = writer.Flush(); err != nil {
		return err
	}
	return saveError
}

// New создает новый менеджер для работы с файлами.
func New(path string) *Manager {
	return &Manager{fileName: path}
}

type Manager struct {
	fileName string
}
