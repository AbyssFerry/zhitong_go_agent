package main

import (
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/abyssferry/zhitong_go_agent/api"
	appllm "github.com/abyssferry/zhitong_go_agent/llm"
)

func main() {
	cfg, err := appllm.LoadConfig(".env")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	listenAddr := os.Getenv("GRPC_ADDR")
	if listenAddr == "" {
		listenAddr = ":50051"
	}
	listenAddr = normalizeListenAddr(listenAddr)

	grpcServer, err := api.NewServer(listenAddr, cfg)
	if err != nil {
		log.Fatalf("create grpc server: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("shutting down gRPC server")
		grpcServer.GracefulStop()
	}()

	if err := grpcServer.Serve(); err != nil {
		log.Fatalf("serve grpc: %v", err)
	}
}

func normalizeListenAddr(listenAddr string) string {
	if strings.TrimSpace(listenAddr) == "" {
		return ":50051"
	}
	if strings.Contains(listenAddr, ":") {
		return listenAddr
	}
	return ":" + listenAddr
}
