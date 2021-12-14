package queue

import "git.jetbrains.space/orbi/fcsd/kit/er"

const (
	ErrCodeQueueMsgUnmarshal        = "QUE-001"
	ErrCodeQueueMsgUnmarshalPayload = "QUE-002"
)

var (
	ErrQueueMsgUnmarshal        = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeQueueMsgUnmarshal, "").Err() }
	ErrQueueMsgUnmarshalPayload = func(cause error) error { return er.WrapWithBuilder(cause, ErrCodeQueueMsgUnmarshalPayload, "").Err() }
)
