package nats

import (
	"sync"

	"github.com/go-logr/logr"
	"github.com/nats-io/go-nats"

	"github.com/LeKovr/mqbridge/types"
)

// Server holds used nats signatures, see mock_nats_test.go
type Server interface {
	ChanSubscribe(subj string, ch chan *nats.Msg) (*nats.Subscription, error)
	Publish(subj string, data []byte) error
	Close()
}

// EndPoint holds endpoint
type EndPoint struct {
	log   logr.Logger
	wg    *sync.WaitGroup
	abort chan string
	quit  chan struct{}
	nc    Server //*nats.Conn
}

// New create endpoint
func New(log logr.Logger, wg *sync.WaitGroup, abort chan string, quit chan struct{}, dsn string) (types.EndPoint, error) {
	log.Info("Endpoint", "dsn", dsn)
	nc, err := nats.Connect(dsn)
	if err != nil {
		return nil, err
	}
	return NewConnected(log, wg, abort, quit, nc)
}

// NewConnected creates endpoint for connected service
func NewConnected(log logr.Logger, wg *sync.WaitGroup, abort chan string, quit chan struct{}, nc Server) (types.EndPoint, error) {
	ep := &EndPoint{log, wg, abort, quit, nc}
	go ep.disconnect()
	return ep, nil
}

// Listen starts all listening goroutines
func (ep EndPoint) Listen(channel string, pipe chan string) error {
	log := ep.log.WithValues("is_in", true, "channel", channel)
	log.Info("Connect NATS producer")
	ch := make(chan *nats.Msg, 64)
	sub, err := ep.nc.ChanSubscribe(channel, ch)
	if err != nil {
		return err
	}
	log.Info("Endpoint connected")
	go ep.reader(log, sub, ch, pipe)
	return nil
}

func (ep EndPoint) reader(log logr.Logger, sub *nats.Subscription, ch chan *nats.Msg, pipe chan string) {
	ep.wg.Add(1)
	defer ep.wg.Done()
	defer sub.Unsubscribe()
	for {
		select {
		case ev := <-ch:
			line := string(ev.Data)
			log.V(1).Info("BRIN ", "line", line)
			pipe <- line
		case <-ep.quit:
			log.V(1).Info("Endpoint close")
			return
		}
	}
}

// Notify starts all notify goroutines
func (ep EndPoint) Notify(channel string, pipe chan string) error {
	log := ep.log.WithValues("is_in", false, "channel", channel)
	log.Info("Connect NATS producer")
	go ep.writer(log, channel, pipe)
	return nil
}

func (ep EndPoint) writer(log logr.Logger, channel string, pipe chan string) {
	ep.wg.Add(1)
	defer ep.wg.Done()
	for {
		select {
		case line := <-pipe:
			err := ep.nc.Publish(channel, []byte(line))
			if err != nil {
				log.Error(err, "Writer")
				//				ep.abort <- "channel" // br.ID
				//				return
			}
		case <-ep.quit:
			log.V(1).Info("Endpoint close")
			return
		}
	}
}

func (ep EndPoint) disconnect() {
	ep.wg.Add(1)
	defer ep.wg.Done()
	defer ep.nc.Close()
	<-ep.quit
	ep.log.V(1).Info("NATS disconnect")
}
