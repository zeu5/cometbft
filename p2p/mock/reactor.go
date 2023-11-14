package mock

import (
	"github.com/zeu5/cometbft/libs/log"
	"github.com/zeu5/cometbft/p2p"
	"github.com/zeu5/cometbft/p2p/conn"
)

type Reactor struct {
	p2p.BaseReactor

	Channels []*conn.ChannelDescriptor
}

func NewReactor() *Reactor {
	r := &Reactor{}
	r.BaseReactor = *p2p.NewBaseReactor("Mock-PEX", r)
	r.SetLogger(log.TestingLogger())
	return r
}

func (r *Reactor) GetChannels() []*conn.ChannelDescriptor { return r.Channels }
func (r *Reactor) AddPeer(_ p2p.Peer)                     {}
func (r *Reactor) RemovePeer(_ p2p.Peer, _ interface{})   {}
func (r *Reactor) Receive(_ p2p.Envelope)                 {}
