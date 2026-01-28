package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sing-ruleset/internal/domain"
	"testing"
)

func TestNewFileConfigRepository(t *testing.T) {
	repo := NewFileConfigRepository()
	if repo == nil {
		t.Fatal("NewFileConfigRepository() should not return nil")
	}
}

func TestFileConfigRepository_GetConfig_Success(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create a valid config file
	config := domain.Config{
		Version: "1.0",
		Sources: map[string][]domain.Source{
			"geoip": {
				{Name: "cn", URL: "https://example.com/cn.txt", Type: "iplist"},
			},
			"geosite": {
				{Name: "ads", URL: "https://example.com/ads.txt", Type: "adguard"},
			},
		},
	}

	configPath := filepath.Join(tempDir, "config.json")

	configData, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("Failed to marshal config: %v", err)
	}

	if err := os.WriteFile(configPath, configData, 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test GetConfig
	repo := NewFileConfigRepository()

	result, err := repo.GetConfig(configPath)
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	// Verify result
	if result.Version != "1.0" {
		t.Errorf("GetConfig() Version = %v, want %v", result.Version, "1.0")
	}

	if len(result.Sources) != 2 {
		t.Errorf("GetConfig() Sources length = %d, want %d", len(result.Sources), 2)
	}

	geoipSources, ok := result.Sources["geoip"]
	if !ok {
		t.Fatal("GetConfig() missing 'geoip' sources")
	}

	if len(geoipSources) != 1 {
		t.Errorf("GetConfig() geoip sources length = %d, want %d", len(geoipSources), 1)
	}

	if geoipSources[0].Name != "cn" {
		t.Errorf("GetConfig() geoip[0].Name = %v, want %v", geoipSources[0].Name, "cn")
	}

	if geoipSources[0].Type != "iplist" {
		t.Errorf("GetConfig() geoip[0].Type = %v, want %v", geoipSources[0].Type, "iplist")
	}
}

func TestFileConfigRepository_GetConfig_FileNotFound(t *testing.T) {
	repo := NewFileConfigRepository()

	_, err := repo.GetConfig("/nonexistent/path/config.json")
	if err == nil {
		t.Error("GetConfig() should return error for non-existent file")
	}
}

func TestFileConfigRepository_GetConfig_InvalidJSON(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create an invalid JSON file
	configPath := filepath.Join(tempDir, "invalid.json")
	if err := os.WriteFile(configPath, []byte("not valid json {{{"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	repo := NewFileConfigRepository()

	_, err := repo.GetConfig(configPath)
	if err == nil {
		t.Error("GetConfig() should return error for invalid JSON")
	}
}

func TestFileConfigRepository_GetConfig_EmptyFile(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create an empty file
	configPath := filepath.Join(tempDir, "empty.json")
	if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	repo := NewFileConfigRepository()

	_, err := repo.GetConfig(configPath)
	if err == nil {
		t.Error("GetConfig() should return error for empty file")
	}
}

func TestFileConfigRepository_GetConfig_EmptyJSON(t *testing.T) {
	// Create a temporary directory
	tempDir := t.TempDir()

	// Create an empty JSON object
	configPath := filepath.Join(tempDir, "empty_object.json")
	if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	repo := NewFileConfigRepository()

	result, err := repo.GetConfig(configPath)
	if err != nil {
		t.Fatalf("GetConfig() error = %v", err)
	}

	// Should return empty config
	if result.Version != "" {
		t.Errorf("GetConfig() Version = %v, want empty string", result.Version)
	}

	if result.Sources != nil {
		t.Errorf("GetConfig() Sources should be nil, got %v", result.Sources)
	}

	if len(result.Sources) != 0 {
		t.Errorf("GetConfig() Sources should be empty, got %v", result.Sources)
	}
}

func TestFileConfigRepository_ImplementsInterface(t *testing.T) {
	// Compile-time check that FileConfigRepository implements domain.Repository
	var _ domain.Repository = (*FileConfigRepository)(nil)
}
