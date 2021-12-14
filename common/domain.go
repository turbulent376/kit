package common

import (
	"context"
	kitContext "git.jetbrains.space/orbi/fcsd/kit/context"
	"git.jetbrains.space/orbi/fcsd/kit/queue"
)

// BaseService can be used as a base service providing some helpers
type BaseService struct {
	Queue queue.Queue
}

// Publish is helper method to publish a message to queue
// Note, it covers payload with &queue.Message, so you have to pass a pure payload object
func (s *BaseService) Publish(ctx context.Context, o interface{}, qt queue.QueueType, topic string) error {

	m := &queue.Message{Payload: o}

	if rCtx, ok := kitContext.Request(ctx); ok {
		m.Ctx = rCtx
	} else {
		return ErrBaseModelCannotPublishToQueue(ctx, topic)
	}

	return s.Queue.Publish(ctx, qt, topic, m)

}
