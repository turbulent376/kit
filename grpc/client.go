package grpc

import (
	"context"
	"fmt"
	kitContext "git.jetbrains.space/orbi/fcsd/kit/context"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ClientConfig is gRPC client configuration
type ClientConfig struct {
	Host string
	Port string
}

type Client struct {
	*readinessAwaiter
	Conn *grpc.ClientConn
}

func NewClient(cfg *ClientConfig) (*Client, error) {

	c := &Client{}

	gc, err := grpc.Dial(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		grpc.WithInsecure(),
		grpc.WithChainUnaryInterceptor(grpc_middleware.ChainUnaryClient(c.unaryClientInterceptor())))
	grpc.WithChainStreamInterceptor(grpc_middleware.ChainStreamClient(c.streamClientInterceptor()))
	if err != nil {
		return nil, ErrGrpcClientDial(err)
	}

	c.Conn = gc
	c.readinessAwaiter = newReadinessAwaiter(gc)

	return c, nil
}

// this middleware is applied on client side
// it retrieves session params from the context (normally it's populated in HTTP middleware or by another caller) and puts it to gRPS metadata
func (c *Client) unaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(parentCtx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx := context.Background()
		if md, ok := kitContext.FromContextToGrpcMD(parentCtx); ok {
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		if err := invoker(ctx, method, req, reply, cc, opts...); err != nil {
			return toAppError(err)
		}
		return nil
	}
}

func (c *Client) streamClientInterceptor() grpc.StreamClientInterceptor {
	return func(parentCtx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx := context.Background()
		if md, ok := kitContext.FromContextToGrpcMD(parentCtx); ok {
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		clStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			err = toAppError(err)
		}
		return clStream, err
	}
}
