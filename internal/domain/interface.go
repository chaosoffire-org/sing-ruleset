package domain

import "context"

type Repository interface {
	GetConfig(path string) (*Config, error)
}

type Downloader interface {
	Download(ctx context.Context, url string, filepath string) error
}

type RuleConverter interface {
	Convert(ctx context.Context, sourcePath string, targetPath string, ruleType string) error
}

type RuleCompiler interface {
	Compile(ctx context.Context, sourcePath string, targetPath string) error
}

type SourceProcessor interface {
	Process(sourcePath string, targetPath string) error
}
