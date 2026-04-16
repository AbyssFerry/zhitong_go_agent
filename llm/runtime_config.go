package llm

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/abyssferry/minichain/utils"
)

const defaultListenAddr = ":50051"

// RuntimeConfig 保存服务启动和默认请求参数。
type RuntimeConfig struct {
	ListenAddr string
	Base       Config
	Chat       ChatDefaults
	Agent      AgentDefaults
}

// ChatDefaults 保存普通流的默认参数。
type ChatDefaults struct {
	SystemPrompt              string
	ContextTrimTokenThreshold int64
	ContextKeepRecentRounds   int64
	Temperature               *float64
	TopP                      *float64
	MaxTokens                 *int
	PresencePenalty           *float64
	FrequencyPenalty          *float64
	Seed                      *int
	RequestTimeout            *time.Duration
}

// AgentDefaults 保存 Agent 流的默认参数。
type AgentDefaults struct {
	SystemPrompt              string
	ContextTrimTokenThreshold int64
	ContextKeepRecentRounds   int64
	Temperature               *float64
	TopP                      *float64
	MaxTokens                 *int
	PresencePenalty           *float64
	FrequencyPenalty          *float64
	Seed                      *int
	RequestTimeout            *time.Duration
	MaxReactRounds            int32
}

// LoadRuntimeConfig 从 .env 读取服务基础配置和默认请求参数。
func LoadRuntimeConfig(envFile string) (RuntimeConfig, error) {
	envMap, err := utils.LoadEnv(envFile)
	if err != nil {
		return RuntimeConfig{}, fmt.Errorf("load env file: %w", err)
	}

	base, err := LoadConfig(envFile)
	if err != nil {
		return RuntimeConfig{}, err
	}

	cfg := RuntimeConfig{
		ListenAddr: strings.TrimSpace(utils.GetEnv(envMap, "GRPC_ADDR", defaultListenAddr)),
		Base:       base,
	}

	chat, err := loadChatDefaults(envMap)
	if err != nil {
		return RuntimeConfig{}, err
	}
	agent, err := loadAgentDefaults(envMap)
	if err != nil {
		return RuntimeConfig{}, err
	}
	cfg.Chat = chat
	cfg.Agent = agent

	return cfg, nil
}

func loadChatDefaults(envMap map[string]string) (ChatDefaults, error) {
	temperature, err := float64PtrEnv(envMap, "CHAT_TEMPERATURE")
	if err != nil {
		return ChatDefaults{}, err
	}
	topP, err := float64PtrEnv(envMap, "CHAT_TOP_P")
	if err != nil {
		return ChatDefaults{}, err
	}
	maxTokens, err := intPtrEnv(envMap, "CHAT_MAX_TOKENS")
	if err != nil {
		return ChatDefaults{}, err
	}
	presencePenalty, err := float64PtrEnv(envMap, "CHAT_PRESENCE_PENALTY")
	if err != nil {
		return ChatDefaults{}, err
	}
	frequencyPenalty, err := float64PtrEnv(envMap, "CHAT_FREQUENCY_PENALTY")
	if err != nil {
		return ChatDefaults{}, err
	}
	seed, err := intPtrEnv(envMap, "CHAT_SEED")
	if err != nil {
		return ChatDefaults{}, err
	}
	timeout, err := durationPtrEnv(envMap, "CHAT_REQUEST_TIMEOUT")
	if err != nil {
		return ChatDefaults{}, err
	}
	contextTrimTokenThreshold, err := int64Env(envMap, "CHAT_CONTEXT_TRIM_TOKEN_THRESHOLD")
	if err != nil {
		return ChatDefaults{}, err
	}
	contextKeepRecentRounds, err := int64Env(envMap, "CHAT_CONTEXT_KEEP_RECENT_ROUNDS")
	if err != nil {
		return ChatDefaults{}, err
	}
	return ChatDefaults{
		SystemPrompt:              strings.TrimSpace(utils.GetEnv(envMap, "CHAT_SYSTEM_PROMPT", "")),
		ContextTrimTokenThreshold: contextTrimTokenThreshold,
		ContextKeepRecentRounds:   contextKeepRecentRounds,
		Temperature:               temperature,
		TopP:                      topP,
		MaxTokens:                 maxTokens,
		PresencePenalty:           presencePenalty,
		FrequencyPenalty:          frequencyPenalty,
		Seed:                      seed,
		RequestTimeout:            timeout,
	}, nil
}

func loadAgentDefaults(envMap map[string]string) (AgentDefaults, error) {
	temperature, err := float64PtrEnv(envMap, "AGENT_TEMPERATURE")
	if err != nil {
		return AgentDefaults{}, err
	}
	topP, err := float64PtrEnv(envMap, "AGENT_TOP_P")
	if err != nil {
		return AgentDefaults{}, err
	}
	maxTokens, err := intPtrEnv(envMap, "AGENT_MAX_TOKENS")
	if err != nil {
		return AgentDefaults{}, err
	}
	presencePenalty, err := float64PtrEnv(envMap, "AGENT_PRESENCE_PENALTY")
	if err != nil {
		return AgentDefaults{}, err
	}
	frequencyPenalty, err := float64PtrEnv(envMap, "AGENT_FREQUENCY_PENALTY")
	if err != nil {
		return AgentDefaults{}, err
	}
	seed, err := intPtrEnv(envMap, "AGENT_SEED")
	if err != nil {
		return AgentDefaults{}, err
	}
	timeout, err := durationPtrEnv(envMap, "AGENT_REQUEST_TIMEOUT")
	if err != nil {
		return AgentDefaults{}, err
	}
	contextTrimTokenThreshold, err := int64Env(envMap, "AGENT_CONTEXT_TRIM_TOKEN_THRESHOLD")
	if err != nil {
		return AgentDefaults{}, err
	}
	contextKeepRecentRounds, err := int64Env(envMap, "AGENT_CONTEXT_KEEP_RECENT_ROUNDS")
	if err != nil {
		return AgentDefaults{}, err
	}
	maxReactRounds, err := int32Env(envMap, "AGENT_MAX_REACT_ROUNDS")
	if err != nil {
		return AgentDefaults{}, err
	}
	return AgentDefaults{
		SystemPrompt:              strings.TrimSpace(utils.GetEnv(envMap, "AGENT_SYSTEM_PROMPT", "")),
		ContextTrimTokenThreshold: contextTrimTokenThreshold,
		ContextKeepRecentRounds:   contextKeepRecentRounds,
		Temperature:               temperature,
		TopP:                      topP,
		MaxTokens:                 maxTokens,
		PresencePenalty:           presencePenalty,
		FrequencyPenalty:          frequencyPenalty,
		Seed:                      seed,
		RequestTimeout:            timeout,
		MaxReactRounds:            maxReactRounds,
	}, nil
}

// ChatOptions 将运行时默认值转换为普通流构造参数。
func (cfg RuntimeConfig) ChatOptions() ChatOptions {
	return ChatOptions{
		SystemPrompt:              cfg.Chat.SystemPrompt,
		ContextTrimTokenThreshold: int(cfg.Chat.ContextTrimTokenThreshold),
		ContextKeepRecentRounds:   int(cfg.Chat.ContextKeepRecentRounds),
		Temperature:               cfg.Chat.Temperature,
		TopP:                      cfg.Chat.TopP,
		MaxTokens:                 cfg.Chat.MaxTokens,
		PresencePenalty:           cfg.Chat.PresencePenalty,
		FrequencyPenalty:          cfg.Chat.FrequencyPenalty,
		Seed:                      cfg.Chat.Seed,
		RequestTimeout:            cfg.Chat.RequestTimeout,
		DebugMessages:             cfg.Base.DebugMessages,
	}
}

// AgentOptions 将运行时默认值转换为 Agent 构造参数。
func (cfg RuntimeConfig) AgentOptions() AgentOptions {
	return AgentOptions{
		SystemPrompt:              cfg.Agent.SystemPrompt,
		ContextTrimTokenThreshold: int(cfg.Agent.ContextTrimTokenThreshold),
		ContextKeepRecentRounds:   int(cfg.Agent.ContextKeepRecentRounds),
		Temperature:               cfg.Agent.Temperature,
		TopP:                      cfg.Agent.TopP,
		MaxTokens:                 cfg.Agent.MaxTokens,
		PresencePenalty:           cfg.Agent.PresencePenalty,
		FrequencyPenalty:          cfg.Agent.FrequencyPenalty,
		Seed:                      cfg.Agent.Seed,
		RequestTimeout:            cfg.Agent.RequestTimeout,
		DebugMessages:             cfg.Base.DebugMessages,
		MaxReactRounds:            int(cfg.Agent.MaxReactRounds),
	}
}

func int64Env(envMap map[string]string, key string) (int64, error) {
	raw := strings.TrimSpace(utils.GetEnv(envMap, key, ""))
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return value, nil
}

func int32Env(envMap map[string]string, key string) (int32, error) {
	raw := strings.TrimSpace(utils.GetEnv(envMap, key, ""))
	if raw == "" {
		return 0, nil
	}
	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	return int32(value), nil
}

func float64PtrEnv(envMap map[string]string, key string) (*float64, error) {
	raw := strings.TrimSpace(utils.GetEnv(envMap, key, ""))
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", key, err)
	}
	return &value, nil
}

func intPtrEnv(envMap map[string]string, key string) (*int, error) {
	raw := strings.TrimSpace(utils.GetEnv(envMap, key, ""))
	if raw == "" {
		return nil, nil
	}
	value, err := strconv.ParseInt(raw, 10, 0)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", key, err)
	}
	intValue := int(value)
	return &intValue, nil
}

func durationPtrEnv(envMap map[string]string, key string) (*time.Duration, error) {
	raw := strings.TrimSpace(utils.GetEnv(envMap, key, ""))
	if raw == "" {
		return nil, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil {
		seconds, secondsErr := strconv.ParseInt(raw, 10, 64)
		if secondsErr != nil {
			return nil, fmt.Errorf("parse %s: %w", key, err)
		}
		duration := time.Duration(seconds) * time.Second
		return &duration, nil
	}
	return &value, nil
}
