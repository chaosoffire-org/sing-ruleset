package service

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"sing-ruleset/internal/domain"
)

// SingBoxRuleCompiler implements domain.RuleCompiler using sing-box CLI.
type SingBoxRuleCompiler struct{}

var _ domain.RuleCompiler = (*SingBoxRuleCompiler)(nil)

// NewSingBoxRuleCompiler creates a new SingBoxRuleCompiler instance.
func NewSingBoxRuleCompiler() *SingBoxRuleCompiler {
	return &SingBoxRuleCompiler{}
}

// Compile compiles a JSON rule file to SRS format using sing-box CLI.
func (c *SingBoxRuleCompiler) Compile(ctx context.Context, sourcePath string, targetPath string) error {
	cmd := exec.CommandContext(ctx, "sing-box", "rule-set", "compile", "--output", targetPath, sourcePath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("sing-box compile output: %s", output)
		return fmt.Errorf("sing-box compile failed: %v", err)
	}

	return nil
}
