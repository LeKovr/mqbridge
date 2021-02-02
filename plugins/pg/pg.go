package pg

import (
	"github.com/go-logr/logr"
	"github.com/go-pg/pg/v9"
	"github.com/go-pg/pg/v9/orm"

	"github.com/LeKovr/mqbridge/types"
)

// Result TODO
type Result interface {
}

// Listener TODO
type Listener interface {
	Close() error
	Channel() <-chan pg.Notification
}

// Server holds used pg signatures, see mock_pg_test.go
type Server interface {
	Exec(query interface{}, params ...interface{}) (orm.Result, error)
	Listen(channels ...string) *pg.Listener
	Close() error
}

/*
_, err := conn.Exec(context.Background(), "listen channelname")
if err != nil {
    return nil
}

if notification, err := conn.WaitForNotification(context.Background()); err != nil {
    // do something with notification
}
*/

// EndPoint holds endpoint
type EndPoint struct {
	types.EndPointAttr
	db Server // *pg.DB
}

// PgMaxRetries holds Pg connect retry count
const PgMaxRetries = 5

// New create endpoint
func New(epa types.EndPointAttr, dsn string) (types.EndPoint, error) {
	epa.Log.Info("Endpoint", "dsn", dsn)

	opts, err := pg.ParseURL(dsn)
	if err != nil {
		return nil, err
	}
	opts.MaxRetries = PgMaxRetries

	/*
		done := make(chan struct{})
		opts.OnConnect = func(c *pg.Conn) error {
			done <- struct{}{}
			return nil
		}
	*/
	db := pg.Connect(opts)
	_, err = db.Exec(`SELECT 1`)
	//	<-done
	if err != nil {
		return nil, err
	}
	epa.Log.Info("Endpoint connected", "dsn", dsn)
	return NewConnected(epa, db)
}

/*
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	// config.Logger = log15adapter.NewLogger(log.New("module", "pgx"))
		conn, err := pgx.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}
*/

// NewConnected creates endpoint for connected service
func NewConnected(epa types.EndPointAttr, db Server) (types.EndPoint, error) {
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

func (ep EndPoint) reader(log logr.Logger, ln *pg.Listener, pipe chan string) {
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
	for {
		select {
		case line := <-pipe:
			_, err := ep.db.Exec("SELECT "+channel+"(?)", line)
			if err != nil {
				log.Error(err, "Writer")
				//				ep.abort <- "channel" // br.ID
				//				return
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
	defer ep.db.Close()
	<-ep.Quit
	ep.Log.V(1).Info("NATS disconnect")
}
