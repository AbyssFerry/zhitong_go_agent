package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/abyssferry/zhitong_go_agent/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/durationpb"
	wrapperspb "google.golang.org/protobuf/types/known/wrapperspb"
)

var (
	chatModel                                   = ""
	chatSystemPrompt                            = "你是一个简洁、可靠的助手。"
	chatUserMessage                             = "请用两句话介绍你自己。"
	chatContextTrimTokenThreshold int64         = 0
	chatContextKeepRecentRounds   int64         = 6
	chatTemperature               float64       = 0.3
	chatTopP                      float64       = 0.9
	chatMaxTokens                 int32         = 2048
	chatPresencePenalty           float64       = 0
	chatFrequencyPenalty          float64       = 0
	chatSeed                      int32         = 0
	chatRequestTimeout            time.Duration = 5 * time.Minute
	chatDebugMessages             bool          = false

	agentModel                                   = ""
	agentSystemPrompt                            = "你是一个会优先调用工具来回答问题的助手。"
	agentUserMessage                             = "现在几点？请先调用工具再回答。"
	agentContextTrimTokenThreshold int64         = 0
	agentContextKeepRecentRounds   int64         = 6
	agentTemperature               float64       = 0.2
	agentTopP                      float64       = 0.9
	agentMaxTokens                 int32         = 2048
	agentPresencePenalty           float64       = 0
	agentFrequencyPenalty          float64       = 0
	agentSeed                      int32         = 0
	agentMaxReactRounds            int32         = 20
	agentRequestTimeout            time.Duration = 5 * time.Minute
	agentDebugMessages             bool          = false
)

func main() {
	addr := flag.String("addr", "127.0.0.1:50051", "gRPC server address")
	flag.Parse()

	conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("create grpc client: %v", err)
	}
	defer conn.Close()

	client := pb.NewZhitongAgentClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	fmt.Println("== ChatStream ==")
	chatStream, err := client.ChatStream(ctx, &pb.ChatStreamRequest{
		Model:                     chatModel,
		SystemPrompt:              chatSystemPrompt,
		Messages:                  []*pb.Message{{Role: "user", Content: chatUserMessage}},
		ContextTrimTokenThreshold: chatContextTrimTokenThreshold,
		ContextKeepRecentRounds:   chatContextKeepRecentRounds,
		Temperature:               wrapperspb.Double(chatTemperature),
		TopP:                      wrapperspb.Double(chatTopP),
		MaxTokens:                 wrapperspb.Int32(chatMaxTokens),
		PresencePenalty:           wrapperspb.Double(chatPresencePenalty),
		FrequencyPenalty:          wrapperspb.Double(chatFrequencyPenalty),
		Seed:                      wrapperspb.Int32(chatSeed),
		RequestTimeout:            durationpb.New(chatRequestTimeout),
		DebugMessages:             chatDebugMessages,
	})
	if err != nil {
		log.Fatalf("chat stream: %v", err)
	}
	consumeStream(chatStream)

	fmt.Println("\n== AgentStream ==")
	agentStream, err := client.AgentStream(ctx, &pb.AgentStreamRequest{
		Model:                     agentModel,
		SystemPrompt:              agentSystemPrompt,
		Messages:                  []*pb.Message{{Role: "user", Content: agentUserMessage}},
		ContextTrimTokenThreshold: agentContextTrimTokenThreshold,
		ContextKeepRecentRounds:   agentContextKeepRecentRounds,
		Temperature:               wrapperspb.Double(agentTemperature),
		TopP:                      wrapperspb.Double(agentTopP),
		MaxTokens:                 wrapperspb.Int32(agentMaxTokens),
		PresencePenalty:           wrapperspb.Double(agentPresencePenalty),
		FrequencyPenalty:          wrapperspb.Double(agentFrequencyPenalty),
		Seed:                      wrapperspb.Int32(agentSeed),
		RequestTimeout:            durationpb.New(agentRequestTimeout),
		DebugMessages:             agentDebugMessages,
		MaxReactRounds:            agentMaxReactRounds,
	})
	if err != nil {
		log.Fatalf("agent stream: %v", err)
	}
	consumeStream(agentStream)
}

type streamConsumer interface {
	Recv() (*pb.StreamResponse, error)
}

func consumeStream(stream streamConsumer) {
	for {
		event, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("recv stream event: %v", err)
		}

		fmt.Printf("[%s] content=%q tool=%q args=%q finish=%q final=%v error=%q\n",
			event.GetEventType(),
			event.GetContent(),
			event.GetToolName(),
			event.GetToolArgs(),
			event.GetFinishReason(),
			event.GetIsFinal(),
			event.GetError(),
		)
	}
}
