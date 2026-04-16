package api

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	minillm "github.com/abyssferry/minichain/llm"
	appllm "github.com/abyssferry/zhitong_go_agent/llm"
)

func TestSingleUserMessageWrapsContent(t *testing.T) {
	messages := singleUserMessage("hello")
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	if messages[0].Role != "user" || messages[0].Content != "hello" {
		t.Fatalf("unexpected message: %+v", messages[0])
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

func TestBuildChatRequestLogIncludesSystemAndUserMessages(t *testing.T) {
	requestTimeout := 90 * time.Second
	request := buildChatRequestLog(appllm.Config{Model: "gpt-test", BaseURL: "https://example.com/v1", APIKey: "secret", DebugMessages: true, DebugRequests: true}, appllm.ChatOptions{
		Model:          "gpt-test",
		SystemPrompt:   "system prompt",
		Temperature:    float64Ptr(0.2),
		TopP:           float64Ptr(0.9),
		MaxTokens:      intPtr(128),
		RequestTimeout: &requestTimeout,
	}, "hello")

	if request.Constructor != "appllm.NewChatModel" {
		t.Fatalf("unexpected constructor: %s", request.Constructor)
	}
	if request.Config.Model != "gpt-test" || !request.Config.APIKeySet {
		t.Fatalf("unexpected config payload: %+v", request.Config)
	}
	if request.Options.SystemPrompt != "system prompt" || request.Options.Model != "gpt-test" {
		t.Fatalf("unexpected options payload: %+v", request.Options)
	}
	if len(request.Input.Messages) != 2 {
		t.Fatalf("unexpected messages length: %d", len(request.Input.Messages))
	}
	if request.Input.Messages[0].Role != "system" || request.Input.Messages[0].Content != "system prompt" {
		t.Fatalf("unexpected first message: %+v", request.Input.Messages[0])
	}
	if request.Input.Messages[1].Role != "user" || request.Input.Messages[1].Content != "hello" {
		t.Fatalf("unexpected second message: %+v", request.Input.Messages[1])
	}
	if request.Options.RequestTimeout == nil || *request.Options.RequestTimeout != "90s" {
		t.Fatalf("unexpected request timeout debug value: %+v", request.Options.RequestTimeout)
	}
}

func TestBuildAgentRequestLogIncludesToolsAndStreamSettings(t *testing.T) {
	requestTimeout := 1500 * time.Millisecond
	request := buildAgentRequestLog(appllm.Config{Model: "gpt-agent", BaseURL: "https://example.com/v1", APIKey: "secret", DebugMessages: true, DebugRequests: true}, appllm.AgentOptions{
		Model:          "gpt-agent",
		SystemPrompt:   "agent system",
		MaxReactRounds: 20,
		RequestTimeout: &requestTimeout,
	}, []minillm.ToolDefinition{{Type: "function"}}, "hello")

	if request.Constructor != "appllm.NewAgent" {
		t.Fatalf("unexpected constructor: %s", request.Constructor)
	}
	if request.Config.Model != "gpt-agent" || !request.Config.APIKeySet {
		t.Fatalf("unexpected config payload: %+v", request.Config)
	}
	if request.Options.MaxReactRounds != 20 || request.Options.SystemPrompt != "agent system" {
		t.Fatalf("unexpected options payload: %+v", request.Options)
	}
	if len(request.Options.Tools) != 1 {
		t.Fatalf("unexpected tools length: %d", len(request.Options.Tools))
	}
	if len(request.Input.Messages) != 2 {
		t.Fatalf("unexpected messages length: %d", len(request.Input.Messages))
	}
	if request.Options.RequestTimeout == nil || *request.Options.RequestTimeout != "1.5s" {
		t.Fatalf("unexpected request timeout debug value: %+v", request.Options.RequestTimeout)
	}
}

func TestBuildChatRequestLogKeepsEmptyFieldsInJSON(t *testing.T) {
	request := buildChatRequestLog(appllm.Config{}, appllm.ChatOptions{}, "")
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}

	text := string(payload)
	if !containsAll(text, []string{`"model":""`, `"base_url":""`, `"system_prompt":""`, `"temperature":null`, `"request_timeout":null`, `"messages"`}) {
		t.Fatalf("unexpected empty-field payload: %s", text)
	}
}

func TestBuildAgentRequestLogKeepsEmptyFieldsInJSON(t *testing.T) {
	request := buildAgentRequestLog(appllm.Config{}, appllm.AgentOptions{}, nil, "")
	payload, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal request failed: %v", err)
	}

	text := string(payload)
	if !containsAll(text, []string{`"model":""`, `"base_url":""`, `"system_prompt":""`, `"temperature":null`, `"request_timeout":null`, `"tools":null`, `"messages"`}) {
		t.Fatalf("unexpected empty-field payload: %s", text)
	}
}

func containsAll(text string, parts []string) bool {
	for _, part := range parts {
		if !strings.Contains(text, part) {
			return false
		}
	}
	return true
}

func float64Ptr(value float64) *float64 { return &value }

func intPtr(value int) *int { return &value }
