package api

import (
	"log"
	"net"

	appllm "github.com/abyssferry/zhitong_go_agent/llm"
	"github.com/abyssferry/zhitong_go_agent/pb"
	"google.golang.org/grpc"
)

// Server 负责 gRPC 服务部署与生命周期管理。
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
}

// NewServer 创建并注册 gRPC 服务。
func NewServer(listenAddr string, cfg appllm.Config) (*Server, error) {
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()
	service, err := NewZhitongAgentService(cfg)
	if err != nil {
		return nil, err
	}
	pb.RegisterZhitongAgentServer(grpcServer, service)

	return &Server{
		grpcServer: grpcServer,
		listener:   listener,
	}, nil
}

// Serve 启动 gRPC 服务。
func (s *Server) Serve() error {
	log.Printf("gRPC server listening on %s", s.listener.Addr().String())
	return s.grpcServer.Serve(s.listener)
}

// GracefulStop 优雅关闭 gRPC 服务。
func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
}

// Run 启动 gRPC 服务并监听退出信号。
func Run(listenAddr string, cfg appllm.Config) error {
	grpcServer, err := NewServer(listenAddr, cfg)
	if err != nil {
		return err
	}
	return grpcServer.Serve()
}
