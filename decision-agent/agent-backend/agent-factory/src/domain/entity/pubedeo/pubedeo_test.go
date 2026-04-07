package pubedeo

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
)

func TestPublishedAgentEo(t *testing.T) {
	t.Parallel()

	eo := &PublishedAgentEo{
		Config:          &daconfvalobj.Config{},
		PublishedByName: "John Doe",
	}

	if eo.Config == nil {
		t.Error("Config should not be nil")
	}

	if eo.PublishedByName != "John Doe" {
		t.Errorf("PublishedByName = %q, want %q", eo.PublishedByName, "John Doe")
	}
}

func TestPublishedTpl(t *testing.T) {
	t.Parallel()

	eo := &PublishedTpl{
		Config:      &daconfvalobj.Config{},
		ProductName: "Test Product",
	}

	if eo.Config == nil {
		t.Error("Config should not be nil")
	}

	if eo.ProductName != "Test Product" {
		t.Errorf("ProductName = %q, want %q", eo.ProductName, "Test Product")
	}
}
