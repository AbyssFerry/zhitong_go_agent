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
)

var (
	chatUserMessage  = "请用两句话介绍你自己。"
	agentUserMessage = "现在几点？请先调用工具再回答。"
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
		Message: chatUserMessage,
	})
	if err != nil {
		log.Fatalf("chat stream: %v", err)
	}
	consumeStream(chatStream)

	fmt.Println("\n== AgentStream ==")
	agentStream, err := client.AgentStream(ctx, &pb.AgentStreamRequest{
		Message: agentUserMessage,
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
