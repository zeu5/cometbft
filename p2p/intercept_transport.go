package p2p

import (
	"net/http"

	"github.com/zeu5/cometbft/p2p/conn"
)

type InterceptNetworkClient struct {
	Addr   string
	server *http.Server
}

type InterceptConfig struct {
	ChannelIDs map[byte]bool
}

type InterceptTransport struct {
	*MultiplexTransport
	cfg     InterceptConfig
	iClient *InterceptNetworkClient
	peers   map[string]*InterceptPeer
}

func NewInterceptTransport(nodeInfo NodeInfo, nodeKey NodeKey, mConfig conn.MConnConfig, iConfig InterceptConfig) *InterceptTransport {
	return &InterceptTransport{
		iClient:            nil,
		cfg:                iConfig,
		MultiplexTransport: NewMultiplexTransport(nodeInfo, nodeKey, mConfig),
		peers:              make(map[string]*InterceptPeer),
	}
}

func (i *InterceptTransport) Accept(pConfig peerConfig) (Peer, error) {
	p, err := i.MultiplexTransport.Accept(pConfig)
	if err != nil {
		return nil, err
	}
	peer := p.(*peer)

	out := &InterceptPeer{
		cfg:        pConfig,
		iTransport: i,
		peer:       peer,
	}
	i.peers[peer.String()] = out
	return out, nil
}

func (i *InterceptTransport) Dial(addr NetAddress, pConfig peerConfig) (Peer, error) {
	p, err := i.MultiplexTransport.Dial(addr, pConfig)
	if err != nil {
		return nil, err
	}
	peer := p.(*peer)

	out := &InterceptPeer{
		cfg:        pConfig,
		iTransport: i,
		peer:       peer,
	}
	i.peers[peer.String()] = out
	return out, nil
}

func (i *InterceptTransport) sendMessage(e Envelope) bool {
	// TODO:
	// 1. Check if the message should be intercepted or not
	// 2. Send with client if so
	return false
}

func (i *InterceptTransport) receiveMessage() {
	// TODO:
	// 1. Need to poll the client for messages
	// 2. Should be part of the transportLifecycle
}

var _ Transport = &InterceptTransport{}
var _ transportLifecycle = &InterceptTransport{}

type InterceptPeer struct {
	cfg        peerConfig
	iTransport *InterceptTransport
	*peer
}

var _ Peer = &InterceptPeer{}

// func (i *InterceptPeer) Send(e Envelope) bool {
// 	return i.iTransport.sendMessage(e)
// }
