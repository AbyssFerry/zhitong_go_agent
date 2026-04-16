package llm

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/abyssferry/minichain/utils"
)

// Config 保存运行 minichain 所需的基础环境配置。
type Config struct {
	Model         string
	APIKey        string
	BaseURL       string
	DebugMessages bool
	DebugRequests bool
}

// LoadConfig 从 .env 读取基础配置。
func LoadConfig(envFile string) (Config, error) {
	envMap, err := utils.LoadEnv(envFile)
	if err != nil {
		return Config{}, fmt.Errorf("load env file: %w", err)
	}

	cfg := Config{
		Model:   strings.TrimSpace(utils.GetEnv(envMap, "MODEL", "")),
		APIKey:  strings.TrimSpace(utils.GetEnv(envMap, "API_KEY", "")),
		BaseURL: strings.TrimSpace(utils.GetEnv(envMap, "BASE_URL", "")),
	}

	if cfg.Model == "" {
		return Config{}, fmt.Errorf("MODEL is required")
	}
	if cfg.APIKey == "" {
		return Config{}, fmt.Errorf("API_KEY is required")
	}
	if cfg.BaseURL == "" {
		return Config{}, fmt.Errorf("BASE_URL is required")
	}

	debugRaw := strings.TrimSpace(utils.GetEnv(envMap, "DEBUG_MESSAGES", "false"))
	debugVal, err := strconv.ParseBool(debugRaw)
	if err != nil {
		return Config{}, fmt.Errorf("parse DEBUG_MESSAGES: %w", err)
	}
	cfg.DebugMessages = debugVal

	requestDebugRaw := strings.TrimSpace(utils.GetEnv(envMap, "DEBUG_REQUESTS", "false"))
	requestDebugVal, err := strconv.ParseBool(requestDebugRaw)
	if err != nil {
		return Config{}, fmt.Errorf("parse DEBUG_REQUESTS: %w", err)
	}
	cfg.DebugRequests = requestDebugVal

	return cfg, nil
}

// ChatOptions 表示普通流模型构造参数。
type ChatOptions struct {
	Model                     string
	SystemPrompt              string
	ContextTrimTokenThreshold int
	ContextKeepRecentRounds   int
	Temperature               *float64
	TopP                      *float64
	MaxTokens                 *int
	Stop                      []string
	PresencePenalty           *float64
	FrequencyPenalty          *float64
	Seed                      *int
	RequestTimeout            *time.Duration
	DebugMessages             bool
}

// AgentOptions 表示 Agent 构造参数。
type AgentOptions struct {
	Model                     string
	SystemPrompt              string
	ContextTrimTokenThreshold int
	ContextKeepRecentRounds   int
	Temperature               *float64
	TopP                      *float64
	MaxTokens                 *int
	Stop                      []string
	PresencePenalty           *float64
	FrequencyPenalty          *float64
	Seed                      *int
	RequestTimeout            *time.Duration
	DebugMessages             bool
	MaxReactRounds            int
}
