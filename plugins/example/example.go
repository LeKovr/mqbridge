package example

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/LeKovr/mqbridge/types"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
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
	parts := strings.SplitN(channel, ":", 2)
	count, err := strconv.Atoi(parts[0])
	if err != nil {
		return errors.Wrap(err, "count")
	}
	delay, err := strconv.Atoi(parts[1])
	if err != nil {
		return errors.Wrap(err, "delay")
	}
	log.Info("Endpoint connected")
	go ep.reader(log, count, delay, pipe)
	return nil
}

// Notify starts all notify goroutines
func (ep EndPoint) Notify(channel string, pipe chan string) error {
	log := ep.log.WithValues("is_in", false, "channel", channel)
	log.Info("Endpoint for STDOUT")
	go ep.printer(log, pipe)
	return nil
}

func (ep EndPoint) reader(log logr.Logger, count, delay int, pipe chan string) {
	ep.wg.Add(1)
	defer ep.wg.Done()
	ticker := time.NewTicker(time.Duration(delay) * time.Millisecond)
	defer ticker.Stop()
	i := 0
	for {
		select {
		case <-ticker.C:
			line := fmt.Sprintf("sample %d", i)
			log.V(1).Info("BRIN ", "line", line)
			pipe <- line
			i++
			if i >= count {
				ep.abort <- "channel"
			}
		case <-ep.quit:
			log.V(1).Info("Endpoint close")
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
		case <-ep.quit:
			log.V(1).Info("Endpoint close")
			return
		}
	}
}
