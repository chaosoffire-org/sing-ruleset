package service

import (
	"sing-ruleset/internal/domain"
	"testing"
)

func TestNewSingBoxRuleCompiler(t *testing.T) {
	compiler := NewSingBoxRuleCompiler()
	if compiler == nil {
		t.Fatal("NewSingBoxRuleCompiler() should not return nil")
	}
}

func TestSingBoxRuleCompiler_ImplementsInterface(t *testing.T) {
	var _ domain.RuleCompiler = (*SingBoxRuleCompiler)(nil)
}
