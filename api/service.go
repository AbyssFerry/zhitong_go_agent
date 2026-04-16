package api

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	minillm "github.com/abyssferry/minichain/llm"
	appllm "github.com/abyssferry/zhitong_go_agent/llm"
	"github.com/abyssferry/zhitong_go_agent/pb"
	"github.com/abyssferry/zhitong_go_agent/tool"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

const (
	defaultChatSystemPrompt  = "You are a concise and reliable assistant."
	defaultAgentSystemPrompt = "You are a helpful assistant. Use tools whenever needed to provide accurate answers."
)

// ZhitongAgentService 实现 proto 中的单一 gRPC service。
type ZhitongAgentService struct {
	pb.UnimplementedZhitongAgentServer
	cfg   appllm.Config
	tools *minillm.ToolRegistry
}

// NewZhitongAgentService 创建服务实例。
func NewZhitongAgentService(cfg appllm.Config) (*ZhitongAgentService, error) {
	registry, err := tool.NewRegistry()
	if err != nil {
		return nil, err
	}
	return &ZhitongAgentService{cfg: cfg, tools: registry}, nil
}

// ChatStream 使用 minichain 普通流能力返回逐步事件。
func (s *ZhitongAgentService) ChatStream(req *pb.ChatStreamRequest, stream pb.ZhitongAgent_ChatStreamServer) error {
	options := chatOptionsFromRequest(req)
	if strings.TrimSpace(options.SystemPrompt) == "" {
		options.SystemPrompt = defaultChatSystemPrompt
	}

	model, err := appllm.NewChatModel(s.cfg, options)
	if err != nil {
		return sendError(stream.Context(), stream, fmt.Errorf("init chat model: %w", err))
	}

	result, err := model.Stream(minillm.InvokeInput{Messages: messagesFromProto(req.GetMessages())})
	if err != nil {
		return sendError(stream.Context(), stream, fmt.Errorf("start chat stream: %w", err))
	}

	return relayStream(stream.Context(), result, stream)
}

// AgentStream 使用 minichain Agent 流能力返回逐步事件。
func (s *ZhitongAgentService) AgentStream(req *pb.AgentStreamRequest, stream pb.ZhitongAgent_AgentStreamServer) error {
	options := agentOptionsFromRequest(req)
	if strings.TrimSpace(options.SystemPrompt) == "" {
		options.SystemPrompt = defaultAgentSystemPrompt
	}
	options.MaxReactRounds = int(req.GetMaxReactRounds())

	agent, err := appllm.NewAgent(s.cfg, options, s.tools)
	if err != nil {
		return sendError(stream.Context(), stream, fmt.Errorf("init agent: %w", err))
	}

	result, err := agent.Stream(minillm.InvokeInput{Messages: messagesFromProto(req.GetMessages())})
	if err != nil {
		return sendError(stream.Context(), stream, fmt.Errorf("start agent stream: %w", err))
	}

	return relayStream(stream.Context(), result, stream)
}

func relayStream(ctx context.Context, result *minillm.StreamResult, stream interface {
	Send(*pb.StreamResponse) error
}) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-result.Events:
			if !ok {
				summary, waitErr := result.Wait()
				if waitErr != nil {
					return sendError(ctx, stream, waitErr)
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

func messagesFromProto(messages []*pb.Message) []minillm.Message {
	if len(messages) == 0 {
		return nil
	}
	converted := make([]minillm.Message, 0, len(messages))
	for _, message := range messages {
		if message == nil {
			continue
		}
		converted = append(converted, minillm.Message{
			Role:       message.GetRole(),
			Content:    message.GetContent(),
			ToolCallID: message.GetToolCallId(),
			ToolCalls:  toolCallsFromProto(message.GetToolCalls()),
		})
	}
	return converted
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

func chatOptionsFromRequest(req *pb.ChatStreamRequest) appllm.ChatOptions {
	if req == nil {
		return appllm.ChatOptions{}
	}
	return appllm.ChatOptions{
		Model:                     req.GetModel(),
		SystemPrompt:              req.GetSystemPrompt(),
		ContextTrimTokenThreshold: int(req.GetContextTrimTokenThreshold()),
		ContextKeepRecentRounds:   int(req.GetContextKeepRecentRounds()),
		Temperature:               doubleValue(req.GetTemperature()),
		TopP:                      doubleValue(req.GetTopP()),
		MaxTokens:                 intValue(req.GetMaxTokens()),
		Stop:                      append([]string(nil), req.GetStop()...),
		PresencePenalty:           doubleValue(req.GetPresencePenalty()),
		FrequencyPenalty:          doubleValue(req.GetFrequencyPenalty()),
		Seed:                      intValue(req.GetSeed()),
		RequestTimeout:            durationValue(req.GetRequestTimeout()),
		DebugMessages:             req.GetDebugMessages(),
	}
}

func agentOptionsFromRequest(req *pb.AgentStreamRequest) appllm.AgentOptions {
	if req == nil {
		return appllm.AgentOptions{}
	}
	return appllm.AgentOptions{
		Model:                     req.GetModel(),
		SystemPrompt:              req.GetSystemPrompt(),
		ContextTrimTokenThreshold: int(req.GetContextTrimTokenThreshold()),
		ContextKeepRecentRounds:   int(req.GetContextKeepRecentRounds()),
		Temperature:               doubleValue(req.GetTemperature()),
		TopP:                      doubleValue(req.GetTopP()),
		MaxTokens:                 intValue(req.GetMaxTokens()),
		Stop:                      append([]string(nil), req.GetStop()...),
		PresencePenalty:           doubleValue(req.GetPresencePenalty()),
		FrequencyPenalty:          doubleValue(req.GetFrequencyPenalty()),
		Seed:                      intValue(req.GetSeed()),
		RequestTimeout:            durationValue(req.GetRequestTimeout()),
		DebugMessages:             req.GetDebugMessages(),
		MaxReactRounds:            int(req.GetMaxReactRounds()),
	}
}

func doubleValue(value *wrapperspb.DoubleValue) *float64 {
	if value == nil {
		return nil
	}
	v := value.GetValue()
	return &v
}

func intValue(value interface{ GetValue() int32 }) *int {
	if value == nil {
		return nil
	}
	v := int(value.GetValue())
	return &v
}

func durationValue(value *durationpb.Duration) *time.Duration {
	if value == nil {
		return nil
	}
	d := value.AsDuration()
	return &d
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
