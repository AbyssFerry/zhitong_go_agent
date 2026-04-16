package llm

import "testing"

func TestNormalizeMaxTokens(t *testing.T) {
	tests := []struct {
		name string
		in   *int
		want *int
	}{
		{name: "nil stays nil", in: nil, want: nil},
		{name: "zero is preserved", in: intPtr(0), want: intPtr(0)},
		{name: "negative is preserved", in: intPtr(-1), want: intPtr(-1)},
		{name: "valid value preserved", in: intPtr(4096), want: intPtr(4096)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeMaxTokens(tt.in)
			if tt.want == nil {
				if got != nil {
					t.Fatalf("normalizeMaxTokens() = %v, want nil", *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("normalizeMaxTokens returned nil")
			}
			if *got != *tt.want {
				t.Fatalf("normalizeMaxTokens() = %d, want %d", *got, *tt.want)
			}
		})
	}
}

func intPtr(value int) *int { return &value }
