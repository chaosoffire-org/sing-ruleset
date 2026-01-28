package service

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"sing-ruleset/internal/domain"
)

// SingBoxConverter implements domain.RuleConverter using sing-box CLI.
type SingBoxConverter struct{}

// Ensure SingBoxConverter implements domain.RuleConverter
var _ domain.RuleConverter = (*SingBoxConverter)(nil)

// NewSingBoxConverter creates a new SingBoxConverter instance.
func NewSingBoxConverter() *SingBoxConverter {
	return &SingBoxConverter{}
}

// Convert converts a source file to SRS format using sing-box CLI.
func (c *SingBoxConverter) Convert(ctx context.Context, sourcePath string, targetPath string, ruleType string) error {
	var cmd *exec.Cmd

	switch ruleType {
	case "adguard":
		cmd = exec.CommandContext(ctx, "sing-box", "rule-set", "convert", "--type", "adguard", "--output", targetPath, sourcePath)
	default:
		return fmt.Errorf("unsupported conversion type: %s", ruleType)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("sing-box convert output: %s", output)
		return fmt.Errorf("sing-box conversion failed: %v", err)
	}

	return nil
}
