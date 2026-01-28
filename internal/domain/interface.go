package domain

import "context"

// Repository defines the interface for configuration storage operations.
type Repository interface {
	GetConfig(path string) (*Config, error)
}

// Downloader defines the interface for downloading files from URLs.
type Downloader interface {
	Download(ctx context.Context, url string, filepath string) error
}

// RuleConverter defines the interface for converting rules to sing-box format.
type RuleConverter interface {
	Convert(ctx context.Context, sourcePath string, targetPath string, ruleType string) error
}

// RuleCompiler defines the interface for compiling JSON rules to SRS format.
type RuleCompiler interface {
	Compile(ctx context.Context, sourcePath string, targetPath string) error
}

// SourceProcessor defines the interface for processing source files.
type SourceProcessor interface {
	Process(ctx context.Context, sourcePath string, targetPath string) error
}
