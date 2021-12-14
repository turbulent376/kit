package grpc

import "git.jetbrains.space/orbi/fcsd/kit/er"

const (
	ErrCodeGrpcClientDial  = "GRPC-001"
	ErrCodeGrpCInvoke      = "GRPC-002"
	ErrCodeGrpcSrvListen   = "GRPC-003"
	ErrCodeGrpcSrvServe    = "GRPC-004"
	ErrCodeGrpcSrvNotReady = "GRPC-005"
)

var (
	ErrGrpcClientDial  = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeGrpcClientDial, "").Err() }
	ErrGrpCInvoke      = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeGrpCInvoke, "").Err() }
	ErrGrpcSrvListen   = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeGrpcSrvListen, "").Err() }
	ErrGrpcSrvServe    = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeGrpcSrvServe, "").Err() }
	ErrGrpcSrvNotReady = func(svc string) error {
		return er.WithBuilder(ErrCodeGrpcSrvNotReady, "service isn't ready within timeout").F(er.FF{"svc": svc}).Err()
	}
)
