package service

import "git.jetbrains.space/orbi/fcsd/kit/er"

const (
	ErrCodeRaftOddSize           = "SVC-001"
	ErrCodeNatsRpc               = "SVC-002"
	ErrCodeStart                 = "SVC-003"
	ErrCodeSvcClusterInitOddSize = "SVC-004"
	ErrCodeRaftInit              = "SVC-005"
	ErrCodeRaftStart             = "SVC-006"
)

var (
	ErrRaftOddSize           = func() error { return er.WithBuilder(ErrCodeRaftOddSize, "cannot start cluster with odd size").Err() }
	ErrNatsRpc               = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeNatsRpc, "").Err() }
	ErrStart                 = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeStart, "").Err() }
	ErrSvcClusterInitOddSize = func() error {
		return er.WithBuilder(ErrCodeSvcClusterInitOddSize, "cannot start cluster with odd size").Err()
	}
	ErrRaftInit  = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeRaftInit, "").Err() }
	ErrRaftStart = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeRaftStart, "").Err() }
)
