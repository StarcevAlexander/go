package storage

import (
	"encoding/json"
	"os"
	"sync"
)

type DataStorage struct {
	filePath string
	mu       sync.Mutex
}

// NewDataStorage создает новый экземпляр DataStorage
func NewDataStorage(filePath string) *DataStorage {
	return &DataStorage{
		filePath: filePath,
	}
}

// LoadData загружает данные из файла
func (ds *DataStorage) LoadData() (map[string]interface{}, error) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	data := make(map[string]interface{})

	file, err := os.ReadFile(ds.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return data, nil // Возвращаем пустые данные, если файла нет
		}
		return nil, err
	}

	if err := json.Unmarshal(file, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// SaveData сохраняет данные в файл
func (ds *DataStorage) SaveData(data map[string]interface{}) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(ds.filePath, jsonData, 0644)
}
