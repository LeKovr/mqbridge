package types

import (
	"fmt"
	"sync"

	"github.com/go-logr/logr"
	"github.com/wojas/genericr"
)

// EndPoint declares plugin interface
type EndPoint interface {
	// Listen holds signature for Listen func which starts listening goroutine
	Listen(id int, channel string, pipe chan string) error

	// Notify holds signature for Notify func which starts notify goroutine
	Notify(id int, channel string, pipe chan string) error
}

// EndPointAttr holds common endpoint attributes
type EndPointAttr struct {
	Log   logr.Logger
	WG    *sync.WaitGroup
	Abort chan string
	Quit  chan struct{}
}

// Printer prints pipe lines to STDOUT
func (ep EndPointAttr) Printer(log logr.Logger, pipe chan string) {
	ep.WG.Add(1)
	defer ep.WG.Done()
	for {
		select {
		case line := <-pipe:
			fmt.Println(line)
		case <-ep.Quit:
			log.V(1).Info("Endpoint close")
			return
		}
	}
}

// NewBlankEndPointAttr creates new EndPointAttr for testing purposes
func NewBlankEndPointAttr() EndPointAttr {
	var wg sync.WaitGroup
	return EndPointAttr{
		Log:   genericr.New(func(e genericr.Entry) {}),
		WG:    &wg,
		Abort: make(chan string),
		Quit:  make(chan struct{}),
	}
}
