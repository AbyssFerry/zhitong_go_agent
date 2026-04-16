package main

import "testing"

func TestNormalizeListenAddr(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty uses default", in: "", want: ":50051"},
		{name: "port only gets prefixed", in: "50051", want: ":50051"},
		{name: "full address preserved", in: ":50051", want: ":50051"},
		{name: "ip and port preserved", in: "127.0.0.1:50051", want: "127.0.0.1:50051"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizeListenAddr(tt.in)
			if got != tt.want {
				t.Fatalf("normalizeListenAddr(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
