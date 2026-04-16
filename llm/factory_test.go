package llm

import "testing"

func TestNormalizeMaxTokens(t *testing.T) {
	tests := []struct {
		name string
		in   *int
		want int
	}{
		{name: "nil uses default", in: nil, want: defaultMaxTokens},
		{name: "zero uses default", in: intPtr(0), want: defaultMaxTokens},
		{name: "negative uses default", in: intPtr(-1), want: defaultMaxTokens},
		{name: "valid value preserved", in: intPtr(4096), want: 4096},
		{name: "upper bound is capped", in: intPtr(70000), want: 65536},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeMaxTokens(tt.in)
			if got == nil {
				t.Fatalf("normalizeMaxTokens returned nil")
			}
			if *got != tt.want {
				t.Fatalf("normalizeMaxTokens() = %d, want %d", *got, tt.want)
			}
		})
	}
}

func intPtr(value int) *int { return &value }
