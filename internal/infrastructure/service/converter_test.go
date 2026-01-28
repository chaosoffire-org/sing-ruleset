package service

import (
	"sing-ruleset/internal/domain"
	"testing"
)

func TestNewSingBoxConverter(t *testing.T) {
	converter := NewSingBoxConverter()
	if converter == nil {
		t.Fatal("NewSingBoxConverter() should not return nil")
	}
}

func TestSingBoxConverter_ImplementsInterface(t *testing.T) {
	var _ domain.RuleConverter = (*SingBoxConverter)(nil)
}
