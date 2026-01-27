package adapter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/netip"
	"os"
	"sing-ruleset/internal/domain"
	"strings"
)

type IPListProcessor struct{}

// Ensure IPListProcessor implements domain.SourceProcessor
var _ domain.SourceProcessor = (*IPListProcessor)(nil)

func NewIPListProcessor() *IPListProcessor {
	return &IPListProcessor{}
}

// Process reads a raw IP list file, filters valid IPs/CIDRs, and writes a sing-box headless rule-set JSON to targetPath.
func (p *IPListProcessor) Process(sourcePath string, targetPath string) error {
	inFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer inFile.Close()

	// Use temporary slice to hold valid IPs.
	// For very large lists, we could write directly to the JSON encoder if we knew the structure structure,
	// but standard "rules": [{"ip_cidr": []}] requires all IPs in one array.
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
		return err
	}
	defer outFile.Close()

	enc := json.NewEncoder(outFile)
	enc.SetIndent("", "  ")
	if err := enc.Encode(config); err != nil {
		return fmt.Errorf("failed to encode json: %w", err)
	}

	return nil
}
