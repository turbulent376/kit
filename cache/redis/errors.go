package redis

import "git.jetbrains.space/orbi/fcsd/kit/er"

const (
	ErrCodeRedisPingErr = "RDS-001"
)

var (
	ErrRedisPingErr = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeRedisPingErr, "").Err() }
)
