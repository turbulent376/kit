package service

import (
	"context"
	"fmt"
	"git.jetbrains.space/orbi/fcsd/kit/log"
	"go.uber.org/atomic"
)

// Config represents cluster configuration
type Config struct {
	Size uint   // Size is cluster size (how many nodes included)
	Log  string // Log - RAFT log path
}

// Service declares an interface each service must implement
type Service interface {
	// GetCode returns the service unique code
	GetCode() string
	// Init initializes the service
	Init(ctx context.Context) error
	// Start executes all background processes
	Start(ctx context.Context) error
	// Close closes the service
	Close(ctx context.Context)
}

// MetaInfo provides an access to service metadata
// it should be declared as a global var on service level and be accessible across the service
type MetaInfo interface {
	// ServiceCode - service unique code
	ServiceCode() string
	// NodeId - unique id of service instance (in cluster mode there are more than one nodes of the same service)
	NodeId() string
	// Leader indicates if the current node is a leader
	Leader() bool
	// SetMeAsLeader allows set node as a leader
	SetMeAsLeader(l bool)
	// InstanceId
	InstanceId() string
}

type metaInfo struct {
	svcCode    string
	nodeId     string
	instanceId string
	leader     *atomic.Bool
}

func NewMetaInfo(svcCode, nodeId string) MetaInfo {
	return &metaInfo{
		svcCode:    svcCode,
		nodeId:     nodeId,
		instanceId: fmt.Sprintf("%s-%s", svcCode, nodeId),
		leader:     atomic.NewBool(true),
	}
}

func (m *metaInfo) ServiceCode() string {
	return m.svcCode
}

func (m *metaInfo) NodeId() string {
	return m.nodeId
}

func (m *metaInfo) InstanceId() string {
	return m.instanceId
}

func (m *metaInfo) Leader() bool {
	return m.leader.Load()
}

func (m *metaInfo) SetMeAsLeader(l bool) {
	m.leader.Store(l)
}

// Cluster defines cluster of a service with built-in RAFT leader election implementation
type Cluster struct {
	Raft      Raft
	Meta      MetaInfo
	logger    log.CLoggerFunc
	isCluster bool
}

func NewCluster(logger log.CLoggerFunc, meta MetaInfo) Cluster {
	return Cluster{Raft: NewRaft(logger), Meta: meta, logger: logger}
}

// Init initializes a service cluster
//
// size - number of nodes in the cluster. Can be either 1 (cluster mode disabled) or more than 2.
// There should be at least 3 nodes to ensure quorum
//
// natsUrl - NATS connection string (RAFT is implemented based on NATS)
//
// ev - allow to be notified as leader is changed
// if nil, no notification needed
func (c *Cluster) Init(config *Config, natsHost, natsPort string, ev OnLeaderChangedEvent) error {

	l := c.logger().Cmp("cluster").Mth("init")

	if config.Size <= 1 {
		// no cluster needed
		l.Warn("no cluster needed for the given size")
		return nil
	}

	if config.Size%2 == 0 {
		return ErrSvcClusterInitOddSize()
	}

	if config.Log == "" {
		config.Log = "/tmp/raft.log"
	}

	natsUrl := fmt.Sprintf("nats://%s:%s", natsHost, natsPort)
	err := c.Raft.Init(&Options{
		ClusterName: c.Meta.ServiceCode(),
		ClusterSize: int(config.Size),
		NatsUrl:     natsUrl,
		LogPath:     config.Log,
	}, func(l bool) {
		c.Meta.SetMeAsLeader(l)
		if ev != nil {
			ev(l)
		}
	})
	if err != nil {
		return ErrRaftInit(err)
	}

	c.isCluster = true

	l.Inf("ok")

	return nil

}

// Start starts the cluster
func (c *Cluster) Start() error {

	if !c.isCluster {
		return nil
	}

	if err := c.Raft.Start(); err != nil {
		return ErrRaftStart(err)
	}
	c.Meta.SetMeAsLeader(c.Raft.AmILeader())
	c.logger().Cmp("cluster").Mth("start").Inf("ok")
	return nil
}

// Close closes the cluster
func (c *Cluster) Close() {

	if !c.isCluster {
		return
	}

	c.Raft.Close()
	c.logger().Cmp("cluster").Mth("close").Inf("ok")
}
