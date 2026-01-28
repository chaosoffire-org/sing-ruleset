// Package cmd provides the command-line interface for sing-ruleset.
package cmd

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"sing-ruleset/internal/adapter"
	"sing-ruleset/internal/app"
	"sing-ruleset/internal/domain"
	"sing-ruleset/internal/infrastructure/repository"
	"sing-ruleset/internal/infrastructure/service"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

var (
	workDir    string
	outputDir  string
	configPath string
	workers    int
)

var rootCmd = &cobra.Command{
	Use:   "sing-ruleset",
	Short: "A tool to generate sing-box rule-sets",
	Long:  `sing-ruleset is a CLI application to download and convert rule-sets for sing-box.`,
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start the rule generation process",
	RunE: func(_ *cobra.Command, _ []string) error {
		// handle relative paths
		if !filepath.IsAbs(workDir) {
			wd, err := os.Getwd()
			if err != nil {
				return err
			}
			workDir = filepath.Join(wd, workDir)
		}

		if !filepath.IsAbs(outputDir) {
			outputDir = filepath.Join(workDir, outputDir)
		}

		if !filepath.IsAbs(configPath) {
			configPath = filepath.Join(workDir, configPath)
		}

		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		log.Printf("Working Directory: %s", workDir)
		log.Printf("Config Path: %s", configPath)
		log.Printf("Output Directory: %s", outputDir)
		log.Printf("Workers: %d", workers)

		// Initialize Infrastructure
		var repo domain.Repository = repository.NewFileConfigRepository()
		client := &http.Client{
			Timeout: 5 * time.Minute,
		}
		var downloader domain.Downloader = service.NewHTTPDownloader(client)
		var ruleConverter domain.RuleConverter = service.NewSingBoxConverter()
		var ruleCompiler domain.RuleCompiler = service.NewSingBoxRuleCompiler()
		var ipProcessor domain.SourceProcessor = adapter.NewIPListProcessor()

		// Initialize Application
		ctx, cancel := context.WithCancel(context.Background())
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		defer cancel()

		go func() {
			<-sigChan
			cancel()
		}()

		application := app.NewApplication(ctx, repo, downloader, ruleConverter, ruleCompiler, ipProcessor)

		// Execute
		return application.GenerateRules(configPath, outputDir, workers)
	},
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Define flags
	runCmd.PersistentFlags().StringVarP(&workDir, "workdir", "d", ".", "Working directory")
	runCmd.PersistentFlags().StringVarP(&outputDir, "output", "o", "output", "Output directory (relative to workdir if not absolute)")
	runCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.json", "Path to config file (relative to workdir if not absolute)")
	runCmd.PersistentFlags().IntVarP(&workers, "workers", "w", runtime.NumCPU(), "Number of concurrent workers")
}
