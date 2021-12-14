package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	kitContext "git.jetbrains.space/orbi/fcsd/kit/context"
	"git.jetbrains.space/orbi/fcsd/kit/log"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"
	"net"
)

// ServerConfig represents gRPC server configuration
type ServerConfig struct {
	Host  string
	Port  string
	Trace bool
}

type Server struct {
	healthpb.HealthServer
	Srv     *grpc.Server
	Service string
	logger  log.CLoggerFunc
	config  *ServerConfig
}

func NewServer(service string, logger log.CLoggerFunc, config *ServerConfig) (*Server, error) {

	s := &Server{
		Service:      service,
		HealthServer: NewHealthServer(),
		logger:       logger,
		config:       config,
	}

	s.Srv = grpc.NewServer(grpc_middleware.WithUnaryServerChain(s.unaryServerInterceptor()), grpc_middleware.WithStreamServerChain(s.streamServerInterceptor()))

	healthpb.RegisterHealthServer(s.Srv, s)

	return s, nil
}

func (s *Server) Listen() error {

	s.logger().Cmp(s.Service).Pr("grpc").Mth("listen").F(log.FF{"port": s.config.Port}).Inf("start listening")

	lis, err := net.Listen("tcp", fmt.Sprint(":", s.config.Port))
	if err != nil {
		return ErrGrpcSrvListen(err)
	}

	err = s.Srv.Serve(lis)
	if err != nil {
		return ErrGrpcSrvServe(err)
	}

	return nil

}

func (s *Server) Close() {
	s.Srv.Stop()
}

// this middleware is applied on server side
// it retrieves gRPC metadata and puts it to the context
func (s *Server) unaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		// convert metadata to request context
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			ctx = kitContext.FromGrpcMD(ctx, md)
		}

		resp, err := handler(ctx, req)

		// tracing
		if s.config.Trace {
			rqJ, _ := json.Marshal(req)
			rsJ, _ := json.Marshal(resp)
			s.logger().Pr("grpc").Cmp(s.Service).Mth(info.FullMethod).C(ctx).
				F(log.FF{"rq": string(rqJ), "rs": string(rsJ)}).
				Trc()
		}

		// logging errors
		if err != nil {
			s.logger().Pr("grpc").Cmp(s.Service).Mth(info.FullMethod).E(err).St().Err()
		}

		// convert to grpc status
		if err != nil {
			err = toGrpcStatus(err)
		}

		return resp, err
	}
}

// this middleware is applied on server side
// it retrieves gRPC metadata and puts it to the context
func (s *Server) streamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		err := handler(srv, ss)

		// logging errors
		if err != nil {
			s.logger().Pr("grpc").Cmp(s.Service).Mth(info.FullMethod).E(err).St().Err()
		}

		// convert to grpc status
		if err != nil {
			err = toGrpcStatus(err)
		}

		return err
	}
}
