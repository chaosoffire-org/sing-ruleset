// Package repository provides data access implementations.
package repository

import (
	"encoding/json"
	"os"
	"sing-ruleset/internal/domain"

	"github.com/sirupsen/logrus"
)

// FileConfigRepository implements domain.Repository for file-based configuration.
type FileConfigRepository struct{}

var _ domain.Repository = (*FileConfigRepository)(nil)

// NewFileConfigRepository creates a new FileConfigRepository instance.
func NewFileConfigRepository() *FileConfigRepository {
	return &FileConfigRepository{}
}

// GetConfig reads and parses the configuration file from the given path.
func (r *FileConfigRepository) GetConfig(path string) (*domain.Config, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}

	defer func() {
		if err := file.Close(); err != nil {
			logrus.Warnf("failed to close config file: %v", err)
		}
	}()

	var config domain.Config

	decoder := json.NewDecoder(file)

	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
