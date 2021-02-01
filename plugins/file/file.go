package file

import (
	"os"

	"github.com/go-logr/logr"
	"github.com/nxadm/tail"
	"gopkg.in/tomb.v1" // need tomb.ErrStillAlive

	"github.com/LeKovr/mqbridge/types"
)

// EndPoint holds endpoint
type EndPoint struct {
	types.EndPointAttr
}

// New creates endpoint
func New(epa types.EndPointAttr, dsn string) (types.EndPoint, error) {
	epa.Log.Info("Endpoint", "dsn", dsn)
	return &EndPoint{epa}, nil
}

// Listen starts all listening goroutines
func (ep EndPoint) Listen(id int, channel string, pipe chan string) error {
	log := ep.Log.WithValues("is_in", true, "channel", channel, "id", id)
	config := tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: true,
	}
	tf, err := tail.TailFile(channel, config)
	if err != nil {
		return err
	}
	log.Info("Endpoint connected")
	go ep.reader(log, tf, pipe)
	return nil
}

// Notify starts all notify goroutines
func (ep EndPoint) Notify(id int, channel string, pipe chan string) error {
	log := ep.Log.WithValues("is_in", false, "channel", channel, "id", id)
	if channel == "-" {
		// send to STDOUT
		log.Info("Endpoint for stdout")
		go ep.Printer(log, pipe)
	} else {
		f, err := os.Create(channel)
		if err != nil {
			return err
		}
		log.Info("Endpoint for file")
		go ep.writer(log, f, pipe)
	}
	return nil
}

func (ep EndPoint) reader(log logr.Logger, tf *tail.Tail, pipe chan string) {
	ep.WG.Add(1)
	defer ep.WG.Done()
	for {
		select {
		case line := <-tf.Lines:
			if tf.Err().Error() != tomb.ErrStillAlive.Error() {
				_ = tf.Stop()
				log.Error(tf.Err(), "Reader")
				ep.Abort <- "channel" // TODO: channel
				return
			}
			log.V(1).Info("BRIN ", "line", line.Text)
			pipe <- line.Text
		case <-ep.Quit:
			log.V(1).Info("Endpoint close")
			_ = tf.Stop() // Cleanup()
			return
		}
	}
}

func (ep EndPoint) writer(log logr.Logger, f *os.File, pipe chan string) {
	ep.WG.Add(1)
	defer ep.WG.Done()

	for {
		select {
		case line := <-pipe:
			_, err := f.WriteString(line + "\n")
			if err != nil {
				log.Error(err, "Writer")
				f.Close()
				ep.Abort <- "channel" // br.ID
				return
			}
		case <-ep.Quit:
			log.V(1).Info("Endpoint close")
			f.Close()
			return
		}
	}
}
