package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	minillm "github.com/abyssferry/minichain/llm"
	appllm "github.com/abyssferry/zhitong_go_agent/llm"
	"github.com/abyssferry/zhitong_go_agent/pb"
	"github.com/abyssferry/zhitong_go_agent/tool"
	"google.golang.org/protobuf/types/known/structpb"
)

// ZhitongAgentService 实现 proto 中的单一 gRPC service。
type ZhitongAgentService struct {
	pb.UnimplementedZhitongAgentServer
	cfg   appllm.RuntimeConfig
	tools *minillm.ToolRegistry
}

// NewZhitongAgentService 创建服务实例。
func NewZhitongAgentService(cfg appllm.RuntimeConfig) (*ZhitongAgentService, error) {
	registry, err := tool.NewRegistry()
	if err != nil {
		return nil, err
	}
	return &ZhitongAgentService{cfg: cfg, tools: registry}, nil
}

// ChatStream 使用 minichain 普通流能力返回逐步事件。
func (s *ZhitongAgentService) ChatStream(req *pb.ChatStreamRequest, stream pb.ZhitongAgent_ChatStreamServer) error {
	if req == nil {
		return sendError(stream.Context(), stream, fmt.Errorf("message is required"))
	}
	message := strings.TrimSpace(req.GetMessage())
	if message == "" {
		return sendError(stream.Context(), stream, fmt.Errorf("message is required"))
	}

	options := s.cfg.ChatOptions()
	if s.cfg.Base.DebugRequests {
		logDebugJSON("chat_request", buildChatRequestLog(s.cfg.Base, options, message))
	}
	model, err := appllm.NewChatModel(s.cfg.Base, options)
	if err != nil {
		return sendError(stream.Context(), stream, fmt.Errorf("init chat model: %w", err))
	}

	result, err := model.Stream(minillm.InvokeInput{Messages: singleUserMessage(message)})
	if err != nil {
		return sendError(stream.Context(), stream, fmt.Errorf("start chat stream: %w", err))
	}

	return relayStream(stream.Context(), result, stream, s.cfg.Base.DebugRequests, "chat")
}

// AgentStream 使用 minichain Agent 流能力返回逐步事件。
func (s *ZhitongAgentService) AgentStream(req *pb.AgentStreamRequest, stream pb.ZhitongAgent_AgentStreamServer) error {
	if req == nil {
		return sendError(stream.Context(), stream, fmt.Errorf("message is required"))
	}
	message := strings.TrimSpace(req.GetMessage())
	if message == "" {
		return sendError(stream.Context(), stream, fmt.Errorf("message is required"))
	}

	options := s.cfg.AgentOptions()
	if s.cfg.Base.DebugRequests {
		logDebugJSON("agent_request", buildAgentRequestLog(s.cfg.Base, options, s.tools.Definitions(), message))
	}

	agent, err := appllm.NewAgent(s.cfg.Base, options, s.tools)
	if err != nil {
		return sendError(stream.Context(), stream, fmt.Errorf("init agent: %w", err))
	}

	result, err := agent.Stream(minillm.InvokeInput{Messages: singleUserMessage(message)})
	if err != nil {
		return sendError(stream.Context(), stream, fmt.Errorf("start agent stream: %w", err))
	}

	return relayStream(stream.Context(), result, stream, s.cfg.Base.DebugRequests, "agent")
}

func relayStream(ctx context.Context, result *minillm.StreamResult, stream interface {
	Send(*pb.StreamResponse) error
}, debug bool, stage string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-result.Events:
			if !ok {
				summary, waitErr := result.Wait()
				if waitErr != nil {
					if debug {
						logDebugJSON(stage+"_response_error", map[string]any{
							"error": waitErr.Error(),
						})
					}
					return sendError(ctx, stream, waitErr)
				}
				if debug {
					logDebugJSON(stage+"_response", summary)
				}
				return stream.Send(summaryToProto(summary))
			}
			if err := stream.Send(eventToProto(event)); err != nil {
				return err
			}
		}
	}
}

func sendError(ctx context.Context, stream interface {
	Send(*pb.StreamResponse) error
}, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return stream.Send(&pb.StreamResponse{EventType: "error", Error: err.Error()})
}

func singleUserMessage(content string) []minillm.Message {
	return []minillm.Message{{Role: "user", Content: content}}
}

func toolCallsFromProto(toolCalls []*pb.ToolCall) []minillm.ToolCall {
	if len(toolCalls) == 0 {
		return nil
	}
	converted := make([]minillm.ToolCall, 0, len(toolCalls))
	for _, call := range toolCalls {
		if call == nil {
			continue
		}
		function := call.GetFunction()
		if function == nil {
			function = &pb.ToolCallFunction{}
		}
		converted = append(converted, minillm.ToolCall{
			ID:    call.GetId(),
			Type:  call.GetType(),
			Index: int(call.GetIndex()),
			Function: minillm.ToolCallFunction{
				Name:      function.GetName(),
				Arguments: function.GetArguments(),
			},
		})
	}
	return converted
}

func eventToProto(event minillm.StreamEvent) *pb.StreamResponse {
	return &pb.StreamResponse{
		EventType:    event.Type,
		Content:      event.Content,
		ToolName:     event.ToolName,
		ToolArgs:     event.RawArguments,
		FinishReason: event.FinishReason,
	}
}

func summaryToProto(summary minillm.StreamSummary) *pb.StreamResponse {
	return &pb.StreamResponse{
		EventType:        "final",
		IsFinal:          true,
		Content:          summary.Content,
		ToolCalls:        toolCallsToProto(summary.ToolCalls),
		Usage:            usageToProto(summary.Usage),
		Id:               summary.ID,
		ModelName:        summary.ModelName,
		AdditionalKwargs: mapToStruct(summary.AdditionalKwargs),
		ResponseMetadata: mapToStruct(summary.ResponseMetadata),
		UsageMetadata:    mapToStruct(summary.UsageMetadata),
		FinishReason:     summary.FinishReason,
	}
}

func toolCallsToProto(toolCalls []minillm.ToolCall) []*pb.ToolCall {
	if len(toolCalls) == 0 {
		return nil
	}
	converted := make([]*pb.ToolCall, 0, len(toolCalls))
	for _, call := range toolCalls {
		converted = append(converted, &pb.ToolCall{
			Id:    call.ID,
			Type:  call.Type,
			Index: int32(call.Index),
			Function: &pb.ToolCallFunction{
				Name:      call.Function.Name,
				Arguments: call.Function.Arguments,
			},
		})
	}
	return converted
}

func usageToProto(usage minillm.Usage) *pb.Usage {
	return &pb.Usage{
		PromptTokens:            int64(usage.PromptTokens),
		CompletionTokens:        int64(usage.CompletionTokens),
		TotalTokens:             int64(usage.TotalTokens),
		PromptTokensDetails:     mapToStruct(usage.PromptTokensDetails),
		CompletionTokensDetails: mapToStruct(usage.CompletionTokensDetails),
	}
}

func mapToStruct(data map[string]any) *structpb.Struct {
	if len(data) == 0 {
		return nil
	}
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	var asMap map[string]any
	if err := json.Unmarshal(jsonBytes, &asMap); err != nil {
		return nil
	}
	result, err := structpb.NewStruct(asMap)
	if err != nil {
		return nil
	}
	return result
}

func logDebugJSON(stage string, value any) {
	payload, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		log.Printf("[debug][%s] marshal failed: %v", stage, err)
		return
	}
	log.Printf("[debug][%s]\n%s", stage, payload)
}

func buildChatRequestLog(cfg appllm.Config, options appllm.ChatOptions, message string) chatRequestDebugPayload {
	return chatRequestDebugPayload{
		Constructor: "appllm.NewChatModel",
		Config: requestDebugConfig{
			Model:         cfg.Model,
			BaseURL:       cfg.BaseURL,
			APIKeySet:     strings.TrimSpace(cfg.APIKey) != "",
			DebugMessages: cfg.DebugMessages,
			DebugRequests: cfg.DebugRequests,
		},
		Options: chatOptionsDebugPayload{
			Model:                     strings.TrimSpace(options.Model),
			SystemPrompt:              options.SystemPrompt,
			ContextTrimTokenThreshold: options.ContextTrimTokenThreshold,
			ContextKeepRecentRounds:   options.ContextKeepRecentRounds,
			Temperature:               options.Temperature,
			TopP:                      options.TopP,
			MaxTokens:                 options.MaxTokens,
			Stop:                      append([]string(nil), options.Stop...),
			PresencePenalty:           options.PresencePenalty,
			FrequencyPenalty:          options.FrequencyPenalty,
			Seed:                      options.Seed,
			RequestTimeout:            durationToSecondsString(options.RequestTimeout),
			DebugMessages:             options.DebugMessages,
		},
		Input: requestDebugInput{
			Messages: []minillm.Message{
				{Role: "system", Content: strings.TrimSpace(options.SystemPrompt)},
				{Role: "user", Content: message},
			},
		},
	}
}

func buildAgentRequestLog(cfg appllm.Config, options appllm.AgentOptions, tools []minillm.ToolDefinition, message string) agentRequestDebugPayload {
	return agentRequestDebugPayload{
		Constructor: "appllm.NewAgent",
		Config: requestDebugConfig{
			Model:         cfg.Model,
			BaseURL:       cfg.BaseURL,
			APIKeySet:     strings.TrimSpace(cfg.APIKey) != "",
			DebugMessages: cfg.DebugMessages,
			DebugRequests: cfg.DebugRequests,
		},
		Options: agentOptionsDebugPayload{
			Model:                     strings.TrimSpace(options.Model),
			SystemPrompt:              options.SystemPrompt,
			ContextTrimTokenThreshold: options.ContextTrimTokenThreshold,
			ContextKeepRecentRounds:   options.ContextKeepRecentRounds,
			Temperature:               options.Temperature,
			TopP:                      options.TopP,
			MaxTokens:                 options.MaxTokens,
			Stop:                      append([]string(nil), options.Stop...),
			PresencePenalty:           options.PresencePenalty,
			FrequencyPenalty:          options.FrequencyPenalty,
			Seed:                      options.Seed,
			RequestTimeout:            durationToSecondsString(options.RequestTimeout),
			DebugMessages:             options.DebugMessages,
			MaxReactRounds:            options.MaxReactRounds,
			Tools:                     append([]minillm.ToolDefinition(nil), tools...),
		},
		Input: requestDebugInput{
			Messages: []minillm.Message{
				{Role: "system", Content: strings.TrimSpace(options.SystemPrompt)},
				{Role: "user", Content: message},
			},
		},
	}
}

type requestDebugConfig struct {
	Model         string `json:"model"`
	BaseURL       string `json:"base_url"`
	APIKeySet     bool   `json:"api_key_set"`
	DebugMessages bool   `json:"debug_messages"`
	DebugRequests bool   `json:"debug_requests"`
}

type requestDebugInput struct {
	Messages []minillm.Message `json:"messages"`
}

type chatOptionsDebugPayload struct {
	Model                     string   `json:"model"`
	SystemPrompt              string   `json:"system_prompt"`
	ContextTrimTokenThreshold int      `json:"context_trim_token_threshold"`
	ContextKeepRecentRounds   int      `json:"context_keep_recent_rounds"`
	Temperature               *float64 `json:"temperature"`
	TopP                      *float64 `json:"top_p"`
	MaxTokens                 *int     `json:"max_tokens"`
	Stop                      []string `json:"stop"`
	PresencePenalty           *float64 `json:"presence_penalty"`
	FrequencyPenalty          *float64 `json:"frequency_penalty"`
	Seed                      *int     `json:"seed"`
	RequestTimeout            *string  `json:"request_timeout"`
	DebugMessages             bool     `json:"debug_messages"`
}

type agentOptionsDebugPayload struct {
	Model                     string                   `json:"model"`
	SystemPrompt              string                   `json:"system_prompt"`
	ContextTrimTokenThreshold int                      `json:"context_trim_token_threshold"`
	ContextKeepRecentRounds   int                      `json:"context_keep_recent_rounds"`
	Temperature               *float64                 `json:"temperature"`
	TopP                      *float64                 `json:"top_p"`
	MaxTokens                 *int                     `json:"max_tokens"`
	Stop                      []string                 `json:"stop"`
	PresencePenalty           *float64                 `json:"presence_penalty"`
	FrequencyPenalty          *float64                 `json:"frequency_penalty"`
	Seed                      *int                     `json:"seed"`
	RequestTimeout            *string                  `json:"request_timeout"`
	DebugMessages             bool                     `json:"debug_messages"`
	MaxReactRounds            int                      `json:"max_react_rounds"`
	Tools                     []minillm.ToolDefinition `json:"tools"`
}

func durationToSecondsString(value *time.Duration) *string {
	if value == nil {
		return nil
	}
	seconds := value.Seconds()
	formatted := strconv.FormatFloat(seconds, 'f', -1, 64) + "s"
	return &formatted
}

type chatRequestDebugPayload struct {
	Constructor string                  `json:"constructor"`
	Config      requestDebugConfig      `json:"config"`
	Options     chatOptionsDebugPayload `json:"options"`
	Input       requestDebugInput       `json:"input"`
}

type agentRequestDebugPayload struct {
	Constructor string                   `json:"constructor"`
	Config      requestDebugConfig       `json:"config"`
	Options     agentOptionsDebugPayload `json:"options"`
	Input       requestDebugInput        `json:"input"`
}
