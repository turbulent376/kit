package stan

import "git.jetbrains.space/orbi/fcsd/kit/er"

const (
	ErrCodeStanNoOpenConn           = "STAN-001"
	ErrCodeStanQtNotSupported       = "STAN-002"
	ErrCodeStanConnect              = "STAN-003"
	ErrCodeStanClose                = "STAN-004"
	ErrCodeStanPublishAtLeastOnce   = "STAN-005"
	ErrCodeStanPublishAtMostOnce    = "STAN-006"
	ErrCodeStanSubscribeAtLeastOnce = "STAN-007"
	ErrCodeStanSubscribeAtMostOnce  = "STAN-008"
)

var (
	ErrStanNoOpenConn     = func() error { return er.WithBuilder(ErrCodeStanNoOpenConn, "no open connections").Err() }
	ErrStanQtNotSupported = func(qt int) error {
		return er.WithBuilder(ErrCodeStanQtNotSupported, "queue type not supported").F(er.FF{"qt": qt}).Err()
	}
	ErrStanConnect              = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeStanConnect, "").Err() }
	ErrStanClose                = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeStanClose, "").Err() }
	ErrStanPublishAtLeastOnce   = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeStanPublishAtLeastOnce, "").Err() }
	ErrStanPublishAtMostOnce    = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeStanPublishAtMostOnce, "").Err() }
	ErrStanSubscribeAtLeastOnce = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeStanSubscribeAtLeastOnce, "").Err() }
	ErrStanSubscribeAtMostOnce  = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeStanSubscribeAtMostOnce, "").Err() }
)
