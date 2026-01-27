package service

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"sing-ruleset/internal/domain"
)

type SingBoxRuleCompiler struct{}

var _ domain.RuleCompiler = (*SingBoxRuleCompiler)(nil)

func NewSingBoxRuleCompiler() *SingBoxRuleCompiler {
	return &SingBoxRuleCompiler{}
}

func (c *SingBoxRuleCompiler) Compile(ctx context.Context, sourcePath string, targetPath string) error {
	cmd := exec.CommandContext(ctx, "sing-box", "rule-set", "compile", "--output", targetPath, sourcePath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("sing-box compile output: %s", output)
		return fmt.Errorf("sing-box compile failed: %v", err)
	}

	return nil
}
