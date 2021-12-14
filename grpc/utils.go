package grpc

import (
	"encoding/json"
	"git.jetbrains.space/orbi/fcsd/kit/er"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGrpcStatus(err error) error {

	// check if it's app error
	if appErr, ok := er.Is(err); ok {

		// is app error has grpc status populated, then set it up
		var grpcStatus = codes.Unknown
		if appErr.GrpcStatus() != nil {
			c := *appErr.GrpcStatus()
			grpcStatus = codes.Code(c)
		}
		st := status.New(grpcStatus, appErr.Message())

		// marshal fields
		ff, _ := json.Marshal(appErr.Fields())

		// put details to gRPC status
		st, _ = st.WithDetails(&AppErrorDetails{
			Code:   appErr.Code(),
			Fields: ff,
		})

		return st.Err()

	} else {
		return status.New(codes.Unknown, err.Error()).Err()
	}
}

//toAppError converts gRPC status to AppError
func toAppError(err error) error {

	res := err

	st := status.Convert(err)
	details := st.Details()
	if len(details) > 0 {
		errDet := details[0]
		if appErrMsg, ok := errDet.(*AppErrorDetails); ok {

			var ff er.FF
			if e := json.Unmarshal(appErrMsg.Fields, &ff); e == nil {
				res = er.WithBuilder(appErrMsg.Code, st.Message()).F(ff).GrpcSt(uint32(st.Code())).Err()
			}

		}
	}

	return res

}
