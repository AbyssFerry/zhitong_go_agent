package api

import (
	"testing"
	"time"

	minillm "github.com/abyssferry/minichain/llm"
	"github.com/abyssferry/zhitong_go_agent/pb"
	"google.golang.org/protobuf/types/known/durationpb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

func TestChatOptionsFromRequestMapsFields(t *testing.T) {
	request := &pb.ChatStreamRequest{
		Model:                     "qwen-test",
		SystemPrompt:              "system",
		ContextTrimTokenThreshold: 32,
		ContextKeepRecentRounds:   4,
		Temperature:               wrapperspb.Double(0.7),
		TopP:                      wrapperspb.Double(0.9),
		MaxTokens:                 wrapperspb.Int32(1024),
		Stop:                      []string{"stop"},
		PresencePenalty:           wrapperspb.Double(0.1),
		FrequencyPenalty:          wrapperspb.Double(0.2),
		Seed:                      wrapperspb.Int32(42),
		RequestTimeout:            durationpb.New(12 * time.Second),
		DebugMessages:             true,
	}

	options := chatOptionsFromRequest(request)
	if options.Model != "qwen-test" || options.SystemPrompt != "system" {
		t.Fatalf("unexpected basic fields: %+v", options)
	}
	if options.ContextTrimTokenThreshold != 32 || options.ContextKeepRecentRounds != 4 {
		t.Fatalf("unexpected context fields: %+v", options)
	}
	if options.MaxTokens == nil || *options.MaxTokens != 1024 {
		t.Fatalf("unexpected max tokens: %+v", options.MaxTokens)
	}
	if options.RequestTimeout == nil || *options.RequestTimeout != 12*time.Second {
		t.Fatalf("unexpected timeout: %+v", options.RequestTimeout)
	}
}

func TestSummaryToProtoPreservesUsageAndToolCalls(t *testing.T) {
	summary := minillm.StreamSummary{
		Content:      "final answer",
		FinishReason: "stop",
		ID:           "resp-1",
		ModelName:    "qwen-test",
		ToolCalls: []minillm.ToolCall{{
			ID:   "tool-1",
			Type: "function",
			Function: minillm.ToolCallFunction{
				Name:      "get_current_time",
				Arguments: "{}",
			},
		}},
		Usage: minillm.Usage{
			PromptTokens:     10,
			CompletionTokens: 20,
			TotalTokens:      30,
		},
	}

	response := summaryToProto(summary)
	if !response.GetIsFinal() {
		t.Fatal("summaryToProto should mark final event")
	}
	if response.GetContent() != "final answer" || response.GetFinishReason() != "stop" {
		t.Fatalf("unexpected response content: %+v", response)
	}
	if response.GetUsage().GetTotalTokens() != 30 {
		t.Fatalf("unexpected usage total tokens: %+v", response.GetUsage())
	}
	if len(response.GetToolCalls()) != 1 || response.GetToolCalls()[0].GetFunction().GetName() != "get_current_time" {
		t.Fatalf("unexpected tool calls: %+v", response.GetToolCalls())
	}
}
