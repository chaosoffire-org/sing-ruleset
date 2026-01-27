package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sing-ruleset/internal/domain"
	"sync"

	"github.com/sirupsen/logrus"
)

type Application struct {
	Repo            domain.Repository
	Downloader      domain.Downloader
	RuleConverter   domain.RuleConverter // For Adguard -> SRS
	RuleCompiler    domain.RuleCompiler  // For JSON -> SRS
	IPListProcessor domain.SourceProcessor
}

func NewApplication(repo domain.Repository, downloader domain.Downloader, ruleConverter domain.RuleConverter, ruleCompiler domain.RuleCompiler, ipProcessor domain.SourceProcessor) *Application {
	return &Application{
		Repo:            repo,
		Downloader:      downloader,
		RuleConverter:   ruleConverter,
		RuleCompiler:    ruleCompiler,
		IPListProcessor: ipProcessor,
	}
}

func (a *Application) GenerateRules(ctx context.Context, configPath string, outputDir string, workers int) error {
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
	wg.Add(length)

	for category, sources := range config.Sources {
		for _, source := range sources {
			sem <- struct{}{} // Acquire token
			go func(errCh chan<- error, cat string, src domain.Source) {
				defer wg.Done()
				defer func() { <-sem }() // Release token

				err := a.processSource(ctx, cat, src, outputDir)
				if err != nil {
					// logrus.Errorf("Error processing %s/%s: %v", cat, src.Name, err)
					errCh <- fmt.Errorf("Error processing %s/%s: %v", cat, src.Name, err)
				} else {
					logrus.Infof("Successfully processed %s/%s", cat, src.Name)
				}
			}(errCh, category, source)
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

func (a *Application) processSource(ctx context.Context, category string, source domain.Source, outputDir string) error {
	categoryPath := filepath.Join(outputDir, category)
	if err := os.MkdirAll(categoryPath, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create category directory: %w", err)
	}

	rawFilePath := filepath.Join(categoryPath, source.Name+".txt")
	srsFilePath := filepath.Join(categoryPath, source.Name+".srs")

	// 1. Download
	if err := a.Downloader.Download(ctx, source.URL, rawFilePath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	switch source.Type {
	case "iplist":
		jsonFilePath := filepath.Join(categoryPath, source.Name+".json")

		if err := a.IPListProcessor.Process(rawFilePath, jsonFilePath); err != nil {
			return fmt.Errorf("iplist processing failed: %w", err)
		}

		if err := a.RuleCompiler.Compile(ctx, jsonFilePath, srsFilePath); err != nil {
			return fmt.Errorf("sing-box compile failed: %w", err)
		}

	default:
		if err := a.RuleConverter.Convert(ctx, rawFilePath, srsFilePath, "adguard"); err != nil {
			return fmt.Errorf("sing-box convert failed: %w", err)
		}
	}

	return nil
}
