package nats

import (
	"github.com/go-logr/logr"
	engine "github.com/nats-io/nats.go"

	"github.com/LeKovr/mqbridge/types"
)

// Server holds used nats signatures, see nats_test.go
type Server interface {
	ChanSubscribe(subj string, ch chan *engine.Msg) (*engine.Subscription, error)
	Publish(subj string, data []byte) error
	Close()
}

// EndPoint holds nats linked endpoint
type EndPoint struct {
	types.EndPointAttr
	nc Server
}

// BufferSize holds channel buffer size for subscription (some messages are lost if chan unbufferred)
const BufferSize = 64

// New creates endpoint
func New(epa types.EndPointAttr, dsn string) (types.EndPoint, error) {
	epa.Log.Info("Endpoint", "dsn", dsn)
	nc, err := engine.Connect(dsn)
	if err != nil {
		return nil, err
	}
	return NewConnected(epa, nc)
}

// NewConnected creates endpoint for connected service
func NewConnected(epa types.EndPointAttr, nc Server) (types.EndPoint, error) {
	ep := &EndPoint{epa, nc}
	go ep.disconnect()
	return ep, nil
}

// Listen starts all listening goroutines
func (ep EndPoint) Listen(id int, channel string, pipe chan string) error {
	log := ep.Log.WithValues("is_in", true, "channel", channel, "id", id)
	log.Info("Connect NATS consumer")
	ch := make(chan *engine.Msg, BufferSize)
	sub, err := ep.nc.ChanSubscribe(channel, ch)
	if err != nil {
		return err
	}
	log.Info("Endpoint connected")
	go ep.reader(log, sub, ch, pipe)
	return nil
}

// Notify starts all notify goroutines
func (ep EndPoint) Notify(id int, channel string, pipe chan string) error {
	log := ep.Log.WithValues("is_in", false, "channel", channel, "id", id)
	log.Info("Connect NATS producer")
	go ep.writer(log, channel, pipe)
	return nil
}

func (ep EndPoint) reader(log logr.Logger, sub *engine.Subscription, ch chan *engine.Msg, pipe chan string) {
	ep.WG.Add(1)
	defer ep.WG.Done()
	defer func() { _ = sub.Unsubscribe() }()
	for {
		select {
		case ev := <-ch:
			line := string(ev.Data)
			log.V(1).Info("BRIN ", "line", line)
			pipe <- line
		case <-ep.Quit:
			log.V(1).Info("Endpoint close")
			return
		}
	}
}

func (ep EndPoint) writer(log logr.Logger, channel string, pipe chan string) {
	ep.WG.Add(1)
	defer ep.WG.Done()
	for {
		select {
		case line := <-pipe:
			if err := ep.nc.Publish(channel, []byte(line)); err != nil {
				log.Error(err, "Writer")
				// ep.abort <- "channel" // br.ID
				// return
			}
			log.V(1).Info("BROUT", "line", line)
		case <-ep.Quit:
			log.V(1).Info("Endpoint close")
			return
		}
	}
}

func (ep EndPoint) disconnect() {
	ep.WG.Add(1)
	defer ep.WG.Done()
	defer ep.nc.Close()
	<-ep.Quit
	ep.Log.V(1).Info("NATS disconnect")
}
