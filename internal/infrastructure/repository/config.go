package repository

import (
	"encoding/json"
	"os"
	"sing-ruleset/internal/domain"
)

type FileConfigRepository struct{}

var _ domain.Repository = (*FileConfigRepository)(nil)

func NewFileConfigRepository() *FileConfigRepository {
	return &FileConfigRepository{}
}

func (r *FileConfigRepository) GetConfig(path string) (*domain.Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config domain.Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
