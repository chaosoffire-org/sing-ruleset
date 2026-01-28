// Package adapter provides adapters for processing different source types.
package adapter

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"sing-ruleset/internal/domain"
	"strings"

	"github.com/sirupsen/logrus"
)

// IPListProcessor implements domain.SourceProcessor for IP list files.
type IPListProcessor struct{}

// Ensure IPListProcessor implements domain.SourceProcessor
var _ domain.SourceProcessor = (*IPListProcessor)(nil)

// NewIPListProcessor creates a new IPListProcessor instance.
func NewIPListProcessor() *IPListProcessor {
	return &IPListProcessor{}
}

// Process reads a raw IP list file, filters valid IPs/CIDRs, and writes a sing-box headless rule-set JSON to targetPath.
func (p *IPListProcessor) Process(ctx context.Context, sourcePath string, targetPath string) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		inFile, err := os.Open(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to open source file: %w", err)
		}

		defer func() {
			if err := inFile.Close(); err != nil {
				logrus.Errorf("failed to close source file: %v", err)
			}
		}()

		var validIPs []string

		scanner := bufio.NewScanner(inFile)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			// Use netip for faster/stricter parsing (Go 1.18+)
			if _, err := netip.ParseAddr(line); err == nil {
				validIPs = append(validIPs, line)
				continue
			}

			if _, err := netip.ParsePrefix(line); err == nil {
				validIPs = append(validIPs, line)
				continue
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scan error: %w", err)
		}

		if len(validIPs) == 0 {
			return fmt.Errorf("no valid IPs found in %s", sourcePath)
		}

		// Define structure inline to avoid package pollution, or could use map["version"]interface{}
		type Rule struct {
			IPCidr []string `json:"ip_cidr,omitempty"`
		}

		type HeadlessRuleSet struct {
			Version int    `json:"version"`
			Rules   []Rule `json:"rules"`
		}

		config := HeadlessRuleSet{
			Version: 1,
			Rules: []Rule{
				{IPCidr: validIPs},
			},
		}

		outFile, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("failed to create target file: %w", err)
		}

		enc := json.NewEncoder(outFile)
		enc.SetIndent("", "  ")

		if err := enc.Encode(config); err != nil {
			closeErr := outFile.Close()
			if closeErr != nil {
				logrus.Errorf("failed to close target file: %v", closeErr)
			}

			removeErr := os.Remove(targetPath)
			if removeErr != nil {
				logrus.Errorf("failed to remove target file: %v", removeErr)
			}

			return fmt.Errorf("failed to encode json: %w", err)
		}

		if err := outFile.Close(); err != nil {
			return fmt.Errorf("failed to close target file: %w", err)
		}
	}

	return nil
}
