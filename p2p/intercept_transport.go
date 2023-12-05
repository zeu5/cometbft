package p2p

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/cosmos/gogoproto/proto"
	"github.com/gin-gonic/gin"
	"github.com/zeu5/cometbft/p2p/conn"
)

type interceptConfig struct {
	ID              int
	ListenAddr      string
	ServerAddr      string
	BecomeByzantine func()
	NodeKey         *NodeKey
}

type interceptNetworkClient struct {
	ID              int
	Addr            string
	ServerAddr      string
	ctr             map[string]int
	ctrLock         *sync.Mutex
	nodeKey         *NodeKey
	becomeByzantine func()

	receivedMessages []*interceptedMessage
	lock             *sync.Mutex

	server *http.Server
}

func newInterceptNetworkClient(iconfig *interceptConfig) *interceptNetworkClient {

	i := &interceptNetworkClient{
		ID:               iconfig.ID,
		Addr:             iconfig.ListenAddr,
		ServerAddr:       iconfig.ServerAddr,
		ctr:              make(map[string]int),
		ctrLock:          new(sync.Mutex),
		nodeKey:          iconfig.NodeKey,
		receivedMessages: make([]*interceptedMessage, 0),
		lock:             new(sync.Mutex),
		becomeByzantine:  iconfig.BecomeByzantine,

		server: nil,
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.POST("/message", i.handleMessage)
	r.POST("/byzantine", i.handleByzantine)

	i.server = &http.Server{
		Addr:    iconfig.ListenAddr,
		Handler: r,
	}

	return i
}

func (i *interceptNetworkClient) nextID(from, to string) string {
	key := from + "_" + to

	i.ctrLock.Lock()
	_, ok := i.ctr[key]
	if !ok {
		i.ctr[key] = 0
	}
	ctr := i.ctr[key]
	i.ctr[key] = ctr + 1
	i.ctrLock.Unlock()

	return fmt.Sprintf("%s_%d", key, ctr)
}

func (i *interceptNetworkClient) handleByzantine(ctx *gin.Context) {
	i.becomeByzantine()
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (i *interceptNetworkClient) handleMessage(ctx *gin.Context) {
	m := &interceptedMessage{}
	if err := ctx.ShouldBindJSON(&m); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to unmarshal request"})
		return
	}

	m.we = &wrappedEnvelope{}
	if err := json.Unmarshal(m.Data, m.we); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "failed to unmarshal request"})
		return
	}

	i.lock.Lock()
	i.receivedMessages = append(i.receivedMessages, m)
	i.lock.Unlock()
	ctx.JSON(http.StatusOK, gin.H{"message": "ok"})
}

func (i *interceptNetworkClient) Start() error {
	go func() {
		i.server.ListenAndServe()
	}()

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DisableKeepAlives:     true,
		},
	}

	replica := map[string]interface{}{
		"id":    i.ID,
		"alias": i.nodeKey.ID(),
		"addr":  i.Addr,
		"keys": map[string]interface{}{
			"public":  i.nodeKey.PubKey(),
			"private": i.nodeKey.PrivKey,
		},
	}
	bs, err := json.Marshal(replica)
	if err != nil {
		return err
	}

	resp, err := client.Post("http://"+i.ServerAddr+"/replica", "application/json", bytes.NewBuffer(bs))
	if err == nil {
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}
	return err
}

func (i *interceptNetworkClient) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return i.server.Shutdown(ctx)
}

func (i *interceptNetworkClient) getMessages() []*interceptedMessage {
	i.lock.Lock()
	out := make([]*interceptedMessage, len(i.receivedMessages))
	copy(out, i.receivedMessages)
	i.receivedMessages = make([]*interceptedMessage, 0)
	i.lock.Unlock()
	return out
}

func (i *interceptNetworkClient) sendMessage(msg *interceptedMessage) error {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 5 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DisableKeepAlives:     true,
		},
	}
	msg.ID = i.nextID(msg.From, msg.To)

	bs, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	resp, err := client.Post("http://"+i.ServerAddr+"/message", "application/json", bytes.NewBuffer(bs))
	if err == nil {
		io.ReadAll(resp.Body)
		resp.Body.Close()
	}
	return err
}

type interceptedMessage struct {
	From string
	To   string
	Data []byte
	Type string
	ID   string

	we *wrappedEnvelope `json:"-"`
}

type wrappedEnvelope struct {
	ChID byte   `json:"chid"`
	Msg  []byte `json:"msg"`
}

type InterceptConfig struct {
	ChannelIDs      map[byte]bool
	MessageTypeFunc func(proto.Message) string
	BecomeByzantine func()
	ListenAddr      string
	ServerAddr      string
	ID              int
}

type InterceptTransport struct {
	*MultiplexTransport
	Alias     string
	cfg       InterceptConfig
	iClient   *interceptNetworkClient
	peers     map[string]*InterceptPeer
	peersLock *sync.Mutex

	doneCh chan struct{}
}

func NewInterceptTransport(nodeInfo NodeInfo, nodeKey *NodeKey, mConfig conn.MConnConfig, iConfig InterceptConfig) *InterceptTransport {
	return &InterceptTransport{
		iClient: newInterceptNetworkClient(&interceptConfig{
			ID:              iConfig.ID,
			ListenAddr:      iConfig.ListenAddr,
			ServerAddr:      iConfig.ServerAddr,
			NodeKey:         nodeKey,
			BecomeByzantine: iConfig.BecomeByzantine,
		}),
		Alias:              string(nodeInfo.ID()),
		cfg:                iConfig,
		MultiplexTransport: NewMultiplexTransport(nodeInfo, *nodeKey, mConfig),
		peers:              make(map[string]*InterceptPeer),
		peersLock:          new(sync.Mutex),
		doneCh:             make(chan struct{}),
	}
}

func (i *InterceptTransport) Close() error {
	err := i.MultiplexTransport.Close()
	i.iClient.Stop()
	close(i.doneCh)
	return err
}

func (i *InterceptTransport) Listen(addr NetAddress) error {
	err := i.iClient.Start()
	if err != nil {
		return err
	}
	go i.receiveMessage()
	return i.MultiplexTransport.Listen(addr)
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
		id:         string(peer.ID()),
		peer:       peer,
	}
	i.peersLock.Lock()
	i.peers[string(peer.ID())] = out
	i.peersLock.Unlock()
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
		id:         string(peer.ID()),
	}
	i.peersLock.Lock()
	i.peers[string(peer.ID())] = out
	i.peersLock.Unlock()
	return out, nil
}

func (i *InterceptTransport) Cleanup(p Peer) {
	i.MultiplexTransport.Cleanup(p)

	i.peersLock.Lock()
	defer i.peersLock.Unlock()
	delete(i.peers, string(p.ID()))
}

func (i *InterceptTransport) sendMessage(to string, e Envelope) bool {
	select {
	case <-i.doneCh:
		return false
	default:
	}

	if _, ok := i.cfg.ChannelIDs[e.ChannelID]; ok {
		pMsg := e.Message
		if w, ok := pMsg.(Wrapper); ok {
			pMsg = w.Wrap()
		}
		msgBytes, err := proto.Marshal(pMsg)
		if err != nil {
			fmt.Printf("Error marshalling envelop message: %s", err)
			return false
		}
		we := &wrappedEnvelope{
			ChID: e.ChannelID,
			Msg:  msgBytes,
		}
		bs, err := json.Marshal(we)
		if err != nil {
			fmt.Printf("Error marshalling wrapped envelope: %s", err)
			return false
		}
		msg := &interceptedMessage{
			From: i.Alias,
			To:   to,
			Data: bs,
			Type: i.cfg.MessageTypeFunc(e.Message),
		}
		i.iClient.sendMessage(msg)
		return true
	}
	return false
}

func (i *InterceptTransport) receiveMessage() {
	for {
		select {
		case <-i.doneCh:
			return
		default:
		}
		messages := i.iClient.getMessages()
		if len(messages) > 0 {
			for _, m := range messages {
				i.peersLock.Lock()
				peer, ok := i.peers[m.From]
				i.peersLock.Unlock()
				if ok {
					peer.receive(m.we.ChID, m.we.Msg)
				}
			}
		}
		time.Sleep(1 * time.Microsecond)
	}
}

var _ TransportWithLifeCycle = &InterceptTransport{}

type InterceptPeer struct {
	cfg        peerConfig
	iTransport *InterceptTransport
	id         string
	*peer
}

var _ Peer = &InterceptPeer{}

func (i *InterceptPeer) Send(e Envelope) bool {
	if intercepted := i.iTransport.sendMessage(i.id, e); intercepted {
		i.Logger.Info("Sent message", "messages", e.Message.String())
		return intercepted
	}
	return i.peer.Send(e)
}

func (i *InterceptPeer) receive(chID byte, message []byte) {
	if !i.IsRunning() {
		return
	}
	reactor := i.cfg.reactorsByCh[chID]
	if reactor == nil {
		// Note that its ok to panic here as it's caught in the conn._recover,
		// which does onPeerError.
		panic(fmt.Sprintf("Unknown channel %X", chID))
	}
	mt := i.cfg.msgTypeByChID[chID]
	msg := proto.Clone(mt)
	err := proto.Unmarshal(message, msg)
	if err != nil {
		panic(fmt.Errorf("unmarshaling message: %s into type: %s", err, reflect.TypeOf(mt)))
	}
	if w, ok := msg.(Unwrapper); ok {
		msg, err = w.Unwrap()
		if err != nil {
			panic(fmt.Errorf("unwrapping message: %s", err))
		}
	}
	i.Logger.Info("Received message", "message", msg.String())
	reactor.Receive(Envelope{
		ChannelID: chID,
		Src:       i.peer,
		Message:   msg,
	})
}
