package llm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadRuntimeConfig(t *testing.T) {
	envFile := filepath.Join(t.TempDir(), ".env")
	envContent := strings.Join([]string{
		"MODEL=gpt-4.1-mini",
		"API_KEY=test-key",
		"BASE_URL=https://example.com/v1",
		"DEBUG_MESSAGES=true",
		"DEBUG_REQUESTS=true",
		"GRPC_ADDR=127.0.0.1:6000",
		"CHAT_SYSTEM_PROMPT=chat system",
		"CHAT_CONTEXT_TRIM_TOKEN_THRESHOLD=2048",
		"CHAT_CONTEXT_KEEP_RECENT_ROUNDS=5",
		"CHAT_TEMPERATURE=0.4",
		"CHAT_TOP_P=0.8",
		"CHAT_MAX_TOKENS=1024",
		"CHAT_PRESENCE_PENALTY=0.1",
		"CHAT_FREQUENCY_PENALTY=0.2",
		"CHAT_SEED=7",
		"CHAT_REQUEST_TIMEOUT=30",
		"AGENT_SYSTEM_PROMPT=agent system",
		"AGENT_CONTEXT_TRIM_TOKEN_THRESHOLD=4096",
		"AGENT_CONTEXT_KEEP_RECENT_ROUNDS=4",
		"AGENT_TEMPERATURE=0.5",
		"AGENT_TOP_P=0.7",
		"AGENT_MAX_TOKENS=2048",
		"AGENT_PRESENCE_PENALTY=0.3",
		"AGENT_FREQUENCY_PENALTY=0.4",
		"AGENT_SEED=9",
		"AGENT_MAX_REACT_ROUNDS=11",
		"AGENT_REQUEST_TIMEOUT=20s",
	}, "\n")
	if err := os.WriteFile(envFile, []byte(envContent), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg, err := LoadRuntimeConfig(envFile)
	if err != nil {
		t.Fatalf("LoadRuntimeConfig() error = %v", err)
	}
	if cfg.ListenAddr != "127.0.0.1:6000" {
		t.Fatalf("unexpected listen addr: %q", cfg.ListenAddr)
	}
	if cfg.Base.Model != "gpt-4.1-mini" || cfg.Base.APIKey != "test-key" || cfg.Base.BaseURL != "https://example.com/v1" {
		t.Fatalf("unexpected base config: %+v", cfg.Base)
	}
	if !cfg.Base.DebugMessages {
		t.Fatalf("unexpected base debug flag: %+v", cfg.Base)
	}
	if !cfg.Base.DebugRequests {
		t.Fatalf("unexpected base request debug flag: %+v", cfg.Base)
	}
	if cfg.Chat.SystemPrompt != "chat system" || cfg.Chat.ContextTrimTokenThreshold != 2048 || cfg.Chat.ContextKeepRecentRounds != 5 {
		t.Fatalf("unexpected chat defaults: %+v", cfg.Chat)
	}
	if cfg.Chat.Temperature == nil || *cfg.Chat.Temperature != 0.4 || cfg.Chat.TopP == nil || *cfg.Chat.TopP != 0.8 {
		t.Fatalf("unexpected chat optional floats: %+v", cfg.Chat)
	}
	if cfg.Chat.MaxTokens == nil || *cfg.Chat.MaxTokens != 1024 || cfg.Chat.PresencePenalty == nil || *cfg.Chat.PresencePenalty != 0.1 {
		t.Fatalf("unexpected chat optional ints: %+v", cfg.Chat)
	}
	if cfg.Chat.FrequencyPenalty == nil || *cfg.Chat.FrequencyPenalty != 0.2 || cfg.Chat.Seed == nil || *cfg.Chat.Seed != 7 {
		t.Fatalf("unexpected chat optional values: %+v", cfg.Chat)
	}
	if cfg.Chat.RequestTimeout == nil || *cfg.Chat.RequestTimeout != 30*time.Second {
		t.Fatalf("unexpected chat timeout: %+v", cfg.Chat)
	}
	if cfg.Agent.SystemPrompt != "agent system" || cfg.Agent.ContextTrimTokenThreshold != 4096 || cfg.Agent.ContextKeepRecentRounds != 4 {
		t.Fatalf("unexpected agent defaults: %+v", cfg.Agent)
	}
	if cfg.Agent.Temperature == nil || *cfg.Agent.Temperature != 0.5 || cfg.Agent.TopP == nil || *cfg.Agent.TopP != 0.7 {
		t.Fatalf("unexpected agent optional floats: %+v", cfg.Agent)
	}
	if cfg.Agent.MaxTokens == nil || *cfg.Agent.MaxTokens != 2048 || cfg.Agent.PresencePenalty == nil || *cfg.Agent.PresencePenalty != 0.3 {
		t.Fatalf("unexpected agent optional ints: %+v", cfg.Agent)
	}
	if cfg.Agent.FrequencyPenalty == nil || *cfg.Agent.FrequencyPenalty != 0.4 || cfg.Agent.Seed == nil || *cfg.Agent.Seed != 9 {
		t.Fatalf("unexpected agent optional values: %+v", cfg.Agent)
	}
	if cfg.Agent.MaxReactRounds != 11 || cfg.Agent.RequestTimeout == nil || *cfg.Agent.RequestTimeout != 20*time.Second {
		t.Fatalf("unexpected agent defaults: %+v", cfg.Agent)
	}
}

func TestDurationPtrEnvAcceptsNumericSeconds(t *testing.T) {
	envMap := map[string]string{
		"CHAT_REQUEST_TIMEOUT":  "90",
		"AGENT_REQUEST_TIMEOUT": "5m",
	}

	chatTimeout, err := durationPtrEnv(envMap, "CHAT_REQUEST_TIMEOUT")
	if err != nil {
		t.Fatalf("durationPtrEnv() error = %v", err)
	}
	if chatTimeout == nil || *chatTimeout != 90*time.Second {
		t.Fatalf("unexpected chat timeout: %+v", chatTimeout)
	}

	agentTimeout, err := durationPtrEnv(envMap, "AGENT_REQUEST_TIMEOUT")
	if err != nil {
		t.Fatalf("durationPtrEnv() error = %v", err)
	}
	if agentTimeout == nil || *agentTimeout != 5*time.Minute {
		t.Fatalf("unexpected agent timeout: %+v", agentTimeout)
	}
}

func TestDurationPtrEnvRejectsInvalidValue(t *testing.T) {
	envMap := map[string]string{
		"CHAT_REQUEST_TIMEOUT": "nonsense",
	}

	_, err := durationPtrEnv(envMap, "CHAT_REQUEST_TIMEOUT")
	if err == nil {
		t.Fatal("expected error for invalid duration value")
	}
}

func TestLoadRuntimeConfigLeavesOptionalValuesEmpty(t *testing.T) {
	envFile := filepath.Join(t.TempDir(), ".env")
	envContent := strings.Join([]string{
		"MODEL=gpt-4.1-mini",
		"API_KEY=test-key",
		"BASE_URL=https://example.com/v1",
	}, "\n")
	if err := os.WriteFile(envFile, []byte(envContent), 0o600); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	cfg, err := LoadRuntimeConfig(envFile)
	if err != nil {
		t.Fatalf("LoadRuntimeConfig() error = %v", err)
	}
	if cfg.Base.DebugMessages {
		t.Fatal("expected debug messages to default to false")
	}
	if cfg.Base.DebugRequests {
		t.Fatal("expected debug requests to default to false")
	}
	if cfg.Chat.SystemPrompt != "" || cfg.Chat.Temperature != nil || cfg.Chat.TopP != nil || cfg.Chat.MaxTokens != nil || cfg.Chat.RequestTimeout != nil {
		t.Fatalf("expected empty chat defaults: %+v", cfg.Chat)
	}
	if cfg.Agent.SystemPrompt != "" || cfg.Agent.Temperature != nil || cfg.Agent.TopP != nil || cfg.Agent.MaxTokens != nil || cfg.Agent.RequestTimeout != nil {
		t.Fatalf("expected empty agent defaults: %+v", cfg.Agent)
	}
}
