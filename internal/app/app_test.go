package app_test

import (
	"context"
	"errors"
	"sing-ruleset/internal/app"
	"sing-ruleset/internal/domain"
	"testing"
)

// Mocks

type mockRepo struct{}

func (m *mockRepo) GetConfig(path string) (*domain.Config, error) {
	return &domain.Config{
		Sources: map[string][]domain.Source{
			"test-category": {
				{Name: "test-source", URL: "http://example.com/list.txt"},
				{Name: "error-source", URL: "http://example.com/error.txt"},
			},
		},
	}, nil
}

type mockDownloader struct{}

func (m *mockDownloader) Download(ctx context.Context, url string, filePath string) error {
	if url == "http://example.com/error.txt" {
		return errors.New("download failed")
	}
	return nil
}

type mockConverter struct{}

func (m *mockConverter) Convert(ctx context.Context, inputPath, outputPath, ruleType string) error {
	return nil
}

type mockCompiler struct{}

func (mc *mockCompiler) Compile(ctx context.Context, inputPath, outputPath string) error {
	return nil
}

type mockProcessor struct{}

func (mp *mockProcessor) Process(inputPath, outputPath string) error {
	return nil
}

func TestApplication_GenerateRules(t *testing.T) {
	// Setup Mocks
	repo := &mockRepo{}
	downloader := &mockDownloader{}
	converter := &mockConverter{}
	compiler := &mockCompiler{}
	processor := &mockProcessor{}

	app := app.NewApplication(repo, downloader, converter, compiler, processor)

	// Test Execution
	// Expectation: It should not return an error even if one source fails (due to log-only behavior)
	err := app.GenerateRules(context.Background(), "config.json", t.TempDir(), 2)
	if err != nil {
		t.Errorf("GenerateRules() unexpected error = %v", err)
	}
}
