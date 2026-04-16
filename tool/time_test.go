package tool

import "testing"

func TestNewRegistryRegistersCurrentTimeTool(t *testing.T) {
	registry, err := NewRegistry()
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	definitions := registry.Definitions()
	if len(definitions) != 1 {
		t.Fatalf("Definitions() len = %d, want 1", len(definitions))
	}

	definition := definitions[0]
	if definition.Function.Name != "get_current_time" {
		t.Fatalf("Definition.Function.Name = %q, want %q", definition.Function.Name, "get_current_time")
	}
	if definition.Function.Description == "" {
		t.Fatal("Definition.Description is empty")
	}
}
