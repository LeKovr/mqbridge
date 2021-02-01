package file

import (
	"fmt"
	"os"
	"sync"

	"github.com/go-logr/logr"
	"github.com/nxadm/tail"
	"gopkg.in/tomb.v1" // need tomb.ErrStillAlive

	"github.com/LeKovr/mqbridge/types"
)

// EndPoint holds endpoint
type EndPoint struct {
	log   logr.Logger
	wg    *sync.WaitGroup
	abort chan string
	quit  chan struct{}
}

// New create endpoint
func New(log logr.Logger, wg *sync.WaitGroup, abort chan string, quit chan struct{}, dsn string) (types.EndPoint, error) {
	log.Info("Endpoint", "dsn", dsn)
	return &EndPoint{log, wg, abort, quit}, nil
}

// Listen starts all listening goroutines
func (ep EndPoint) Listen(channel string, pipe chan string) error {
	log := ep.log.WithValues("is_in", true, "channel", channel)
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
func (ep EndPoint) Notify(channel string, pipe chan string) error {
	log := ep.log.WithValues("is_in", false, "channel", channel)
	if channel == "-" {
		// send to STDOUT
		log.Info("Endpoint for stdout")
		go ep.printer(log, pipe)
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
	ep.wg.Add(1)
	defer ep.wg.Done()
	for {
		select {
		case line := <-tf.Lines:
			if tf.Err().Error() != tomb.ErrStillAlive.Error() {
				_ = tf.Stop()
				log.Error(tf.Err(), "Reader")
				ep.abort <- "channel" // TODO: channel
				return
			}
			log.V(1).Info("BRIN ", "line", line.Text)
			pipe <- line.Text
		case <-ep.quit:
			log.V(1).Info("Endpoint close")
			tf.Stop() // Cleanup()
			return
		}
	}
}

func (ep EndPoint) writer(log logr.Logger, f *os.File, pipe chan string) {
	ep.wg.Add(1)
	defer ep.wg.Done()

	for {
		select {
		case line := <-pipe:
			_, err := f.WriteString(line + "\n")
			if err != nil {
				log.Error(err, "Writer")
				f.Close()
				ep.abort <- "channel" // br.ID
				return
			}
		case <-ep.quit:
			log.V(1).Info("Endpoint close")
			f.Close()
			return
		}
	}
}

func (ep EndPoint) printer(log logr.Logger, pipe chan string) {
	ep.wg.Add(1)
	defer ep.wg.Done()
	for {
		select {
		case line := <-pipe:
			fmt.Println(line)
			// check err
		case <-ep.quit:
			log.V(1).Info("Endpoint close")
			return
		}
	}
}
