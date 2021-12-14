package stan

import (
	"context"
	"encoding/json"
	"fmt"
	kitContext "git.jetbrains.space/orbi/fcsd/kit/context"
	"git.jetbrains.space/orbi/fcsd/kit/log"
	"git.jetbrains.space/orbi/fcsd/kit/queue"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/stan.go"
)

type stanImpl struct {
	conn     stan.Conn
	clientId string
	logger   log.CLoggerFunc
}

func New(logger log.CLoggerFunc) queue.Queue {
	return &stanImpl{
		logger: logger,
	}
}

func (s *stanImpl) l() log.CLogger {
	return s.logger().Pr("queue").Cmp("stan")
}

func (s *stanImpl) Open(ctx context.Context, clientId string, config *queue.Config) error {

	l := s.l().Mth("open").F(log.FF{"client": clientId, "host": config.Host}).Dbg("connecting")

	s.clientId = clientId
	url := fmt.Sprintf("nats://%s:%s", config.Host, config.Port)
	c, err := stan.Connect(config.ClusterId, clientId, stan.NatsURL(url))
	if err != nil {
		return ErrStanConnect(err)
	}
	s.conn = c

	l.Inf("ok")

	return nil
}

func (s *stanImpl) Close() error {
	if s.conn != nil {
		err := s.conn.Close()
		s.conn = nil
		if err != nil {
			return ErrStanClose(err)
		}
		s.l().Mth("close").Inf("closed")
	}
	return nil
}

func (s *stanImpl) Publish(ctx context.Context, qt queue.QueueType, topic string, msg *queue.Message) error {

	l := s.l().Mth("publish").F(log.FF{"topic": topic, "type": qt.String()})

	if msg.Ctx == nil {
		msg.Ctx = kitContext.NewRequestCtx().Queue().WithNewRequestId()
	}

	if s.conn == nil {
		return ErrStanNoOpenConn()
	}

	m, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	l.Dbg("ok").TrcF("%s\n", string(m))

	if qt == queue.QueueTypeAtLeastOnce {
		if err := s.conn.Publish(topic, m); err != nil {
			return ErrStanPublishAtLeastOnce(err)
		}
	} else if qt == queue.QueueTypeAtMostOnce {
		if err := s.conn.NatsConn().Publish(topic, m); err != nil {
			return ErrStanPublishAtMostOnce(err)
		}
	} else {
		return ErrStanQtNotSupported(int(qt))
	}
	return nil

}

func (s *stanImpl) Subscribe(qt queue.QueueType, topic string, receiverChan chan<- []byte) error {

	l := s.l().Mth("received").F(log.FF{"topic": topic, "type": qt.String()})

	if qt == queue.QueueTypeAtLeastOnce {

		_, err := s.conn.Subscribe(topic, func(m *stan.Msg) {
			l.TrcF("%s\n", string(m.Data))
			receiverChan <- m.Data
		}, stan.DurableName(s.clientId))
		if err != nil {
			return ErrStanSubscribeAtLeastOnce(err)
		}
		return nil

	} else if qt == queue.QueueTypeAtMostOnce {
		_, err := s.conn.NatsConn().Subscribe(topic, func(m *nats.Msg) {
			l.TrcF("%s\n", string(m.Data))
			receiverChan <- m.Data
		})
		if err != nil {
			return ErrStanSubscribeAtMostOnce(err)
		}
		return nil
	} else {
		return ErrStanQtNotSupported(int(qt))
	}

}

func (s *stanImpl) SubscribeLB(qt queue.QueueType, topic, loadBalancingGroup string, receiverChan chan<- []byte) error {
	l := s.l().Mth("received").F(log.FF{"topic": topic, "type": qt.String(), "lbGrp": loadBalancingGroup})

	if qt == queue.QueueTypeAtLeastOnce {

		_, err := s.conn.QueueSubscribe(topic, loadBalancingGroup, func(m *stan.Msg) {
			l.TrcF("%s\n", string(m.Data))
			receiverChan <- m.Data
		}, stan.DurableName(s.clientId))
		if err != nil {
			return ErrStanSubscribeAtLeastOnce(err)
		}
		return nil

	} else if qt == queue.QueueTypeAtMostOnce {
		_, err := s.conn.NatsConn().QueueSubscribe(topic, loadBalancingGroup, func(m *nats.Msg) {
			l.TrcF("%s\n", string(m.Data))
			receiverChan <- m.Data
		})
		if err != nil {
			return ErrStanSubscribeAtMostOnce(err)
		}
		return nil
	} else {
		return ErrStanQtNotSupported(int(qt))
	}
}
