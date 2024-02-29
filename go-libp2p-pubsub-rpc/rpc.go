package rpc

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"
	"sync"

	util "github.com/ipfs/boxo/util"
	"github.com/ipfs/go-cid"
	cbor "github.com/ipfs/go-ipld-cbor"
	golog "github.com/ipfs/go-log/v2"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/peer"
)

var (
	log = golog.Logger("psrpc")

	// ErrResponseNotReceived indicates a response was not received after publishing a message.
	ErrResponseNotReceived = errors.New("response not received")
)

// EventHandler is used to receive topic peer events.
type EventHandler func(from peer.ID, topic string, msg []byte)

// MessageHandler is used to receive topic messages.
type MessageHandler func(from peer.ID, topic string, msg []byte) ([]byte, error)

// Response wraps a message response.
type Response struct {
	// ID is the cid.Cid of the received message.
	ID string
	// From is the peer.ID of the sender.
	From peer.ID
	// Data is the message data.
	Data []byte
	// Err is an error from the sender.
	Err error
}

type internalResponse struct {
	ID   string
	From []byte
	Data []byte
	Err  string
}

func init() {
	cbor.RegisterCborType(internalResponse{})
}

func responseTopic(base string, pid peer.ID) string {
	return path.Join(base, pid.String(), "_response")
}

type ongoingMessage struct {
	ctx       context.Context
	data      []byte
	opts      []pubsub.PubOpt
	respCh    chan internalResponse
	republish bool
}

// Topic provides a nice interface to a libp2p pubsub topic.
type Topic struct {
	ps             *pubsub.PubSub
	host           peer.ID
	eventHandler   EventHandler
	messageHandler MessageHandler

	ongoing  map[cid.Cid]ongoingMessage
	resTopic *Topic

	t *pubsub.Topic
	h *pubsub.TopicEventHandler
	s *pubsub.Subscription

	ctx    context.Context
	cancel context.CancelFunc

	lk sync.Mutex
}

// NewTopic returns a new topic for the host.
func NewTopic(ctx context.Context, ps *pubsub.PubSub, host peer.ID, topic string, subscribe bool) (*Topic, error) {
	t, err := newTopic(ctx, ps, host, topic, subscribe)
	if err != nil {
		return nil, fmt.Errorf("creating topic: %v", err)
	}
	t.resTopic, err = newTopic(ctx, ps, host, responseTopic(topic, host), true)
	if err != nil {
		_ = t.Close()
		return nil, fmt.Errorf("creating response topic: %v", err)
	}
	t.resTopic.eventHandler = t.resEventHandler
	t.resTopic.messageHandler = t.resMessageHandler
	return t, nil
}

func newTopic(ctx context.Context, ps *pubsub.PubSub, host peer.ID, topic string, subscribe bool) (*Topic, error) {
	top, err := ps.Join(topic)
	if err != nil {
		return nil, fmt.Errorf("joining topic: %v", err)
	}

	handler, err := top.EventHandler()
	if err != nil {
		_ = top.Close()
		return nil, fmt.Errorf("getting topic handler: %v", err)
	}

	var sub *pubsub.Subscription
	if subscribe {
		sub, err = top.Subscribe()
		if err != nil {
			handler.Cancel()
			_ = top.Close()
			return nil, fmt.Errorf("subscribing to topic: %v", err)
		}
	}

	t := &Topic{
		ps:      ps,
		host:    host,
		t:       top,
		h:       handler,
		s:       sub,
		ongoing: make(map[cid.Cid]ongoingMessage),
	}
	t.ctx, t.cancel = context.WithCancel(ctx)

	go t.watch()
	if t.s != nil {
		go t.listen()
	}

	return t, nil
}

// Close the topic.
func (t *Topic) Close() error {
	t.lk.Lock()
	defer t.lk.Unlock()
	t.cancel()
	t.h.Cancel()
	if t.s != nil {
		t.s.Cancel()
	}
	if err := t.t.Close(); err != nil {
		return err
	}
	if t.resTopic != nil {
		return t.resTopic.Close()
	}
	return nil
}

// SetEventHandler sets a handler func that will be called with peer events.
func (t *Topic) SetEventHandler(handler EventHandler) {
	t.lk.Lock()
	defer t.lk.Unlock()
	t.eventHandler = handler
}

// SetMessageHandler sets a handler func that will be called with topic messages.
// A subscription is required for the handler to be called.
func (t *Topic) SetMessageHandler(handler MessageHandler) {
	t.lk.Lock()
	defer t.lk.Unlock()
	t.messageHandler = handler
}

// Publish data. Note that the data may arrive peers duplicated. And as a
// result, if WithMultiResponse is supplied, the response may be duplicated as
// well. See PublishOptions for option details.
func (t *Topic) Publish(
	ctx context.Context,
	data []byte,
	opts ...PublishOption,
) (<-chan Response, error) {
	args := defaultOptions
	for _, op := range opts {
		if err := op(&args); err != nil {
			return nil, fmt.Errorf("applying option: %v", err)
		}
	}

	var respCh chan internalResponse
	msgID := cid.NewCidV1(cid.Raw, util.Hash(data))
	if !args.ignoreResponse {
		respCh = make(chan internalResponse)
	}

	if !args.ignoreResponse || args.republish {
		t.lk.Lock()
		t.ongoing[msgID] = ongoingMessage{
			ctx:       ctx,
			data:      data,
			opts:      args.pubOpts,
			respCh:    respCh,
			republish: args.republish,
		}
		t.lk.Unlock()
	}

	if err := t.t.Publish(ctx, data, args.pubOpts...); err != nil {
		return nil, fmt.Errorf("publishing to topic: %v", err)
	}

	if args.ignoreResponse && !args.republish {
		return nil, nil
	}
	resultCh := make(chan Response)
	go func() {
		defer func() {
			t.lk.Lock()
			delete(t.ongoing, msgID)
			t.lk.Unlock()
			close(resultCh)
		}()
		for {
			select {
			case <-ctx.Done():
				if args.ignoreResponse {
					return
				}
				if !args.multiResponse {
					resultCh <- Response{Err: ErrResponseNotReceived}
				}
				return
			case r := <-respCh:
				res := Response{
					ID:   r.ID,
					From: peer.ID(r.From),
					Data: r.Data,
				}
				if r.Err != "" {
					res.Err = errors.New(r.Err)
				}
				if !args.ignoreResponse {
					resultCh <- res
				}
				if !args.multiResponse {
					return
				}
			}
		}
	}()
	return resultCh, nil
}

func (t *Topic) watch() {
	for {
		e, err := t.h.NextPeerEvent(t.ctx)
		if err != nil {
			break
		}
		var msg string
		switch e.Type {
		case pubsub.PeerJoin:
			msg = "JOINED"
			// Note: it looks like we are publishing to this
			// specific peer, but the rpc library doesn't have the
			// ability, so it actually does is to republish to all
			// peers.
			t.republishTo(e.Peer)
		case pubsub.PeerLeave:
			msg = "LEFT"
		default:
			continue
		}
		t.lk.Lock()
		if t.eventHandler != nil {
			t.eventHandler(e.Peer, t.t.String(), []byte(msg))
		}
		t.lk.Unlock()
	}
}

func (t *Topic) republishTo(p peer.ID) {
	t.lk.Lock()
	for _, m := range t.ongoing {
		if m.republish {
			go func(m ongoingMessage) {
				log.Debugf("republishing %s because peer %s newly joins", t.t, p)
				if err := t.t.Publish(m.ctx, m.data, m.opts...); err != nil {
					log.Errorf("republishing to topic: %v", err)
				}
			}(m)
		}
	}
	t.lk.Unlock()
}

func (t *Topic) listen() {
	for {
		msg, err := t.s.Next(t.ctx)
		if err != nil {
			break
		}
		if msg.ReceivedFrom.String() == t.host.String() {
			continue
		}
		t.lk.Lock()
		handler := t.messageHandler
		t.lk.Unlock()
		if handler == nil {
			log.Warnf("didn't process topic message since we don't have a handler")
			continue
		}
		go processSubscriptionMessage(handler, msg.ReceivedFrom, t, msg.Data)
	}
}

func processSubscriptionMessage(handler MessageHandler, from peer.ID, t *Topic, msgData []byte) {
	res, err := handler(from, t.t.String(), msgData)
	if err != nil {
		log.Errorf("subcription message handler: %v", err)
		// Intentionally not returning since we send the error
		// to the other side in the response.
	}
	if !strings.Contains(t.t.String(), "/_response") {
		// This is a normal message; respond with data and error
		msgID := cid.NewCidV1(cid.Raw, util.Hash(msgData))
		t.publishResponse(from, msgID, res, err)
	}
}

func (t *Topic) publishResponse(from peer.ID, id cid.Cid, data []byte, e error) {
	topic, err := newTopic(t.ctx, t.ps, t.host, responseTopic(t.t.String(), from), false)
	if err != nil {
		log.Errorf("creating response topic: %v", err)
		return
	}
	defer func() {
		if err := topic.Close(); err != nil {
			log.Errorf("closing response topic: %v", err)
		}
	}()
	topic.SetEventHandler(t.resEventHandler)

	res := internalResponse{
		ID: id.String(),
		// From: Set on the receiver end using validated data from the received pubsub message
		Data: data,
	}
	if e != nil {
		res.Err = e.Error()
	}
	msg, err := cbor.DumpObject(&res)
	if err != nil {
		log.Errorf("encoding response: %v", err)
		return
	}

	if err := topic.t.Publish(t.ctx, msg); err != nil {
		log.Errorf("publishing response: %v", err)
	}
}

func (t *Topic) resEventHandler(from peer.ID, topic string, msg []byte) {
	log.Debugf("%s response peer event: %s %s", topic, from, msg)
}

// resMessageHandler handles responses to messages.
func (t *Topic) resMessageHandler(from peer.ID, topic string, msg []byte) ([]byte, error) {
	var res internalResponse
	if err := cbor.DecodeInto(msg, &res); err != nil {
		return nil, fmt.Errorf("decoding response: %v", err)
	}
	id, err := cid.Decode(res.ID)
	if err != nil {
		return nil, fmt.Errorf("decoding response id: %v", err)
	}
	res.From = []byte(from)

	log.Debugf("%s response from %s: %s", topic, from, res.ID)
	t.lk.Lock()
	m, exists := t.ongoing[id]
	t.lk.Unlock()
	if exists {
		if m.respCh != nil {
			m.respCh <- res
		}
	} else {
		log.Debugf("%s response from %s arrives too late, discarding", topic, from)
	}
	return nil, nil // no response to a response
}
