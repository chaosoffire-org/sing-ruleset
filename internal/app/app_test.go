package app_test

import (
	"context"
	"errors"
	"sing-ruleset/internal/app"
	"sing-ruleset/internal/domain"
	"testing"
)

// Mocks

type mockRepo struct {
	config *domain.Config
	err    error
}

func (m *mockRepo) GetConfig(path string) (*domain.Config, error) {
	return m.config, m.err
}

type mockDownloader struct {
	err error
}

func (m *mockDownloader) Download(ctx context.Context, url string, filePath string) error {
	return m.err
}

type mockConverter struct {
	err error
}

func (m *mockConverter) Convert(ctx context.Context, inputPath, outputPath, ruleType string) error {
	return m.err
}

type mockCompiler struct {
	err error
}

func (mc *mockCompiler) Compile(ctx context.Context, inputPath, outputPath string) error {
	return mc.err
}

type mockProcessor struct {
	err error
}

func (mp *mockProcessor) Process(ctx context.Context, inputPath, outputPath string) error {
	return mp.err
}

func TestApplication_GenerateRules(t *testing.T) {
	validConfig := &domain.Config{
		Sources: map[string][]domain.Source{
			"test-category": {
				{Name: "test-source-adguard", URL: "http://example.com/list.txt", Type: "adguard"},
				{Name: "test-source-iplist", URL: "http://example.com/ip.txt", Type: "iplist"},
			},
		},
	}

	tests := []struct {
		name        string
		configRef   *domain.Config
		configErr   error
		downloadErr error
		convertErr  error
		compileErr  error
		processErr  error
		expectErr   bool
		ctx         context.Context
	}{
		{
			name:      "Success",
			configRef: validConfig,
			expectErr: false,
			ctx:       context.Background(),
		},
		{
			name:      "Config Load Failure",
			configRef: nil,
			configErr: errors.New("config error"),
			expectErr: true,
			ctx:       context.Background(),
		},
		{
			name:        "Download Failure",
			configRef:   validConfig,
			downloadErr: errors.New("download fail"),
			expectErr:   false, // Log only
			ctx:         context.Background(),
		},
		{
			name:       "Convert Failure",
			configRef:  validConfig,
			convertErr: errors.New("convert fail"),
			expectErr:  false, // Log only
			ctx:        context.Background(),
		},
		{
			name:       "Compile Failure",
			configRef:  validConfig,
			compileErr: errors.New("compile fail"),
			expectErr:  false, // Log only
			ctx:        context.Background(),
		},
		{
			name:       "Process Failure",
			configRef:  validConfig,
			processErr: errors.New("process fail"),
			expectErr:  false, // Log only
			ctx:        context.Background(),
		},
		{
			name:      "Context Cancelled",
			configRef: validConfig,
			expectErr: false, // Should exit cleanly or error? Logic says break loop, return nil
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				return ctx
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockRepo{config: tt.configRef, err: tt.configErr}
			downloader := &mockDownloader{err: tt.downloadErr}
			converter := &mockConverter{err: tt.convertErr}
			compiler := &mockCompiler{err: tt.compileErr}
			processor := &mockProcessor{err: tt.processErr}

			app := app.NewApplication(tt.ctx, repo, downloader, converter, compiler, processor)

			err := app.GenerateRules("config.json", t.TempDir(), 2)

			if (err != nil) != tt.expectErr {
				t.Errorf("GenerateRules() error = %v, expectErr %v", err, tt.expectErr)
			}
		})
	}
}
