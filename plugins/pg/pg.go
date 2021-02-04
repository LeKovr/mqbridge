package pg

import (
	"fmt"

	"github.com/go-logr/logr"
	engine "github.com/go-pg/pg/v9"

	"github.com/LeKovr/mqbridge/types"
)

// EndPoint holds database linked endpoint
type EndPoint struct {
	types.EndPointAttr
	db *engine.DB
}

const (
	// PgMaxRetries holds Pg connect retry count
	PgMaxRetries = 5

	// SQLPubFmt holds notification query format string
	SQLPubFmt = "SELECT %s(?)"
)

// New create endpoint
func New(epa types.EndPointAttr, dsn string) (types.EndPoint, error) {
	epa.Log.Info("Endpoint", "dsn", dsn)

	opts, err := engine.ParseURL(dsn)
	if err != nil {
		return nil, err
	}
	opts.MaxRetries = PgMaxRetries
	db := engine.Connect(opts)
	_, err = db.Exec(`SELECT 1`)
	if err != nil {
		return nil, err
	}
	epa.Log.Info("Endpoint connected", "dsn", dsn)
	return NewConnected(epa, db)
}

// NewConnected creates endpoint for connected service
func NewConnected(epa types.EndPointAttr, db *engine.DB) (types.EndPoint, error) {
	ep := &EndPoint{epa, db}
	go ep.disconnect()
	return ep, nil
}

// Listen starts listening goroutine
func (ep EndPoint) Listen(id int, channel string, pipe chan string) error {
	log := ep.Log.WithValues("is_in", true, "channel", channel, "id", id)
	log.Info("Connect PG consumer")

	listener := ep.db.Listen(channel)
	log.Info("Endpoint connected")
	go ep.reader(log, listener, pipe)
	return nil
}

// Notify starts all notify goroutines
func (ep EndPoint) Notify(id int, channel string, pipe chan string) error {
	log := ep.Log.WithValues("is_in", false, "channel", channel, "id", id)
	log.Info("Connect PG producer")
	go ep.writer(log, channel, pipe)
	return nil
}

func (ep EndPoint) reader(log logr.Logger, ln *engine.Listener, pipe chan string) {
	ep.WG.Add(1)
	defer ep.WG.Done()
	defer ln.Close()
	ch := ln.Channel()
	for {
		select {
		case line := <-ch:
			log.V(1).Info("BRIN ", "line", line.Payload)
			pipe <- line.Payload
		case <-ep.Quit:
			log.V(1).Info("Endpoint close")
			return
		}
	}
}

func (ep EndPoint) writer(log logr.Logger, channel string, pipe chan string) {
	ep.WG.Add(1)
	defer ep.WG.Done()
	sql := fmt.Sprintf(SQLPubFmt, channel)
	for {
		select {
		case line := <-pipe:
			_, err := ep.db.Exec(sql, line)
			if err != nil {
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
	<-ep.Quit
	if *ep.db != (engine.DB{}) {
		ep.Log.V(1).Info("PG disconnect")
		ep.db.Close()
	}
}
