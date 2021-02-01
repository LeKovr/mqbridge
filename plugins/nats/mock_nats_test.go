package nats_test

import (
	"github.com/nats-io/go-nats"
)

type Server struct {
}

func (srv Server) ChanSubscribe(subj string, ch chan *nats.Msg) (*nats.Subscription, error) {
	return &nats.Subscription{}, nil
}

func (srv Server) Publish(subj string, data []byte) error {
	return nil
}

func (srv Server) Close() {
}
