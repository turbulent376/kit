package service

import (
	"git.jetbrains.space/orbi/fcsd/kit/log"
	"github.com/nats-io/graft"
	"github.com/nats-io/nats.go"
)

type OnLeaderChangedEvent func(leader bool)

type Raft interface {
	Init(opt *Options, ev OnLeaderChangedEvent) error
	Start() error
	Close()
	AmILeader() bool
}

type raftImpl struct {
	logger          log.CLoggerFunc
	rpc             *graft.NatsRpcDriver
	ci              *graft.ClusterInfo
	handler         *graft.ChanHandler
	node            *graft.Node
	opt             *Options
	errChan         chan error
	stateChangeChan chan graft.StateChange
}

type Options struct {
	// ClusterName - name must be the same for all nodes within cluster
	ClusterName string
	// ClusterSize - expected number of nodes in a cluster. Must be odd number
	ClusterSize int
	// NATS url
	NatsUrl string
	// logs
	LogPath string
}

func NewRaft(logger log.CLoggerFunc) Raft {
	return &raftImpl{
		logger:          logger,
		errChan:         make(chan error),
		stateChangeChan: make(chan graft.StateChange),
	}
}

func (r *raftImpl) l() log.CLogger {
	return r.logger().Cmp("raft")
}

func (r *raftImpl) Init(opt *Options, onLeaderChangedEvent OnLeaderChangedEvent) error {

	l := r.l().Mth("init").F(log.FF{"size": opt.ClusterSize})

	if opt.ClusterSize <= 1 {
		// no cluster needed
		l.Warn("no cluster needed for the given size")
		return nil
	}

	if opt.ClusterSize%2 == 0 {
		return ErrRaftOddSize()
	}

	r.opt = opt

	options := nats.GetDefaultOptions()
	options.Url = opt.NatsUrl

	r.ci = &graft.ClusterInfo{Name: opt.ClusterName, Size: opt.ClusterSize}

	rpc, err := graft.NewNatsRpc(&options)
	if err != nil {
		return ErrNatsRpc(err)
	}
	r.rpc = rpc

	r.handler = graft.NewChanHandler(r.stateChangeChan, r.errChan)

	go func() {
		l := r.l().Mth("state-handler")
		for {
			select {

			case sc := <-r.stateChangeChan:
				onLeaderChangedEvent(r.AmILeader())
				l.DbgF("state changed: from %s to %s", sc.From.String(), sc.To.String())
				if sc.To != sc.From {
					if sc.To == graft.LEADER {
						onLeaderChangedEvent(true)
					} else {
						onLeaderChangedEvent(false)
					}
				}

			case err := <-r.errChan:
				r.l().Mth("err-handler").E(err).Err()
			}
		}
	}()

	l.Inf("ok")

	return nil
}

func (r *raftImpl) Start() error {

	l := r.l().Mth("start")

	if r.rpc != nil {
		node, err := graft.New(*r.ci, r.handler, r.rpc, r.opt.LogPath)
		if err != nil {
			return ErrStart(err)
		}
		r.node = node

		l.Inf("ok")
	}

	return nil
}

func (r *raftImpl) AmILeader() bool {
	return r.node != nil && r.node.State() == graft.LEADER
}

func (r *raftImpl) Close() {

	l := r.l().Mth("close")

	if r.rpc != nil {
		r.node.Close()
		r.rpc.Close()
	}

	close(r.errChan)
	close(r.stateChangeChan)

	l.Inf("ok")

}
