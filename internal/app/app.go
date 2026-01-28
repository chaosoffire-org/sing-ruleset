// Package app provides the application layer for orchestrating rule generation.
package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sing-ruleset/internal/domain"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Application orchestrates the rule generation process.
type Application struct {
	GlobalContext   context.Context
	Repo            domain.Repository
	Downloader      domain.Downloader
	RuleConverter   domain.RuleConverter // For Adguard -> SRS
	RuleCompiler    domain.RuleCompiler  // For JSON -> SRS
	IPListProcessor domain.SourceProcessor
}

// NewApplication creates a new Application with the given dependencies.
func NewApplication(ctx context.Context, repo domain.Repository, downloader domain.Downloader, ruleConverter domain.RuleConverter, ruleCompiler domain.RuleCompiler, ipProcessor domain.SourceProcessor) *Application {
	return &Application{
		GlobalContext:   ctx,
		Repo:            repo,
		Downloader:      downloader,
		RuleConverter:   ruleConverter,
		RuleCompiler:    ruleCompiler,
		IPListProcessor: ipProcessor,
	}
}

// GenerateRules generates rule sets from the given configuration.
func (a *Application) GenerateRules(configPath string, outputDir string, workers int) error {
	config, err := a.Repo.GetConfig(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	length := 0
	for _, values := range config.Sources {
		length += len(values)
	}

	if length == 0 {
		return fmt.Errorf("no sources found in config")
	}

	sem := make(chan struct{}, workers)
	errCh := make(chan error, length)

	var wg sync.WaitGroup

loop:
	for category, sources := range config.Sources {
		for _, source := range sources {
			select {
			case <-a.GlobalContext.Done():
				break loop
			case sem <- struct{}{}: // Acquire token
				wg.Add(1)

				go func(errCh chan<- error, cat string, src domain.Source) {
					defer wg.Done()
					defer func() { <-sem }() // Release token

					err := a.processSource(cat, src, outputDir)
					if err != nil {
						errCh <- fmt.Errorf("error processing %s/%s: %v", cat, src.Name, err)
					} else {
						logrus.Infof("Successfully processed %s/%s", cat, src.Name)
					}
				}(errCh, category, source)
			}
		}
	}

	wg.Wait()
	close(sem)
	close(errCh)

	for err := range errCh {
		logrus.Error(err.Error())
	}

	return nil
}

func (a *Application) processSource(category string, source domain.Source, outputDir string) error {
	categoryPath := filepath.Join(outputDir, category)

	if err := os.MkdirAll(categoryPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create category directory: %w", err)
	}

	rawFilePath := filepath.Join(categoryPath, source.Name+".txt")
	srsFilePath := filepath.Join(categoryPath, source.Name+".srs")

	// 1. Download
	ctx, cancel := context.WithTimeout(a.GlobalContext, 5*time.Minute)
	defer cancel()

	if err := a.Downloader.Download(ctx, source.URL, rawFilePath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	switch source.Type {
	case "iplist":
		jsonFilePath := filepath.Join(categoryPath, source.Name+".json")

		ctx, cancel = context.WithTimeout(a.GlobalContext, 5*time.Minute)
		defer cancel()

		if err := a.IPListProcessor.Process(ctx, rawFilePath, jsonFilePath); err != nil {
			return fmt.Errorf("iplist processing failed: %w", err)
		}

		compileCtx, compileCancel := context.WithTimeout(a.GlobalContext, 5*time.Minute)
		defer compileCancel()

		if err := a.RuleCompiler.Compile(compileCtx, jsonFilePath, srsFilePath); err != nil {
			return fmt.Errorf("sing-box compile failed: %w", err)
		}

	default:
		convertCtx, convertCancel := context.WithTimeout(a.GlobalContext, 5*time.Minute)
		defer convertCancel()

		if err := a.RuleConverter.Convert(convertCtx, rawFilePath, srsFilePath, "adguard"); err != nil {
			return fmt.Errorf("sing-box convert failed: %w", err)
		}
	}

	return nil
}
