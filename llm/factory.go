package llm

import minillm "github.com/abyssferry/minichain/llm"

// NewChatModel 构造普通聊天模型。
func NewChatModel(cfg Config, opts ChatOptions) (*minillm.ChatModel, error) {
	modelName := cfg.Model
	if opts.Model != "" {
		modelName = opts.Model
	}

	return minillm.InitChatModel(minillm.ChatModelOptions{
		Model:                     modelName,
		SystemPrompt:              opts.SystemPrompt,
		APIKey:                    cfg.APIKey,
		BaseURL:                   cfg.BaseURL,
		ContextTrimTokenThreshold: opts.ContextTrimTokenThreshold,
		ContextKeepRecentRounds:   opts.ContextKeepRecentRounds,
		Temperature:               opts.Temperature,
		TopP:                      opts.TopP,
		MaxTokens:                 opts.MaxTokens,
		Stop:                      append([]string(nil), opts.Stop...),
		PresencePenalty:           opts.PresencePenalty,
		FrequencyPenalty:          opts.FrequencyPenalty,
		Seed:                      opts.Seed,
		RequestTimeout:            opts.RequestTimeout,
		DebugMessages:             cfg.DebugMessages || opts.DebugMessages,
	})
}

// NewAgent 构造 Agent 模型。
func NewAgent(cfg Config, opts AgentOptions, tools *minillm.ToolRegistry) (*minillm.Agent, error) {
	modelName := cfg.Model
	if opts.Model != "" {
		modelName = opts.Model
	}

	return minillm.CreateAgent(minillm.AgentOptions{
		Model:                     modelName,
		SystemPrompt:              opts.SystemPrompt,
		APIKey:                    cfg.APIKey,
		BaseURL:                   cfg.BaseURL,
		MaxReactRounds:            opts.MaxReactRounds,
		Tools:                     tools,
		ContextTrimTokenThreshold: opts.ContextTrimTokenThreshold,
		ContextKeepRecentRounds:   opts.ContextKeepRecentRounds,
		Temperature:               opts.Temperature,
		TopP:                      opts.TopP,
		MaxTokens:                 opts.MaxTokens,
		Stop:                      append([]string(nil), opts.Stop...),
		PresencePenalty:           opts.PresencePenalty,
		FrequencyPenalty:          opts.FrequencyPenalty,
		Seed:                      opts.Seed,
		RequestTimeout:            opts.RequestTimeout,
		DebugMessages:             cfg.DebugMessages || opts.DebugMessages,
	})
}

func normalizeMaxTokens(maxTokens *int) *int {
	return maxTokens
}
