package example

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/LeKovr/mqbridge/types"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
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
func (ep EndPoint) Notify(id int, channel string, pipe chan string) error {
	log := ep.Log.WithValues("is_in", false, "channel", channel, "id", id)
	log.Info("Endpoint for STDOUT")
	go ep.Printer(log, pipe)
	return nil
}

func (ep EndPoint) reader(log logr.Logger, count, delay int, pipe chan string) {
	ep.WG.Add(1)
	defer ep.WG.Done()
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
				ep.Abort <- "channel"
			}
		case <-ep.Quit:
			log.V(1).Info("Endpoint close")
			return
		}
	}
}
