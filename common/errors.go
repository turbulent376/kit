package common

import (
	"context"
	"git.jetbrains.space/orbi/fcsd/kit/er"
)

const (
	ErrCodeBaseModelCannotPublishToQueue = "CMN-001"
)

var (
	ErrBaseModelCannotPublishToQueue = func(ctx context.Context, topic string) error {
		return er.WithBuilder(ErrCodeBaseModelCannotPublishToQueue, "cannot publish to topic").C(ctx).F(er.FF{"topic": topic}).Err()
	}
)
