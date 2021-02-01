package types

import (
	"sync"

	"github.com/go-logr/logr"
)

// EndPoint declares plugin interface
type EndPoint interface {
	// Listen holds signature for Listen func which starts listening goroutine
	Listen(channel string, pipe chan string) error

	// Notify holds signature for Notify func which starts notify goroutine
	Notify(channel string, pipe chan string) error
}

// EndPointAttr holds common endpoint attributes
type EndPointAttr struct {
	Log   logr.Logger
	WG    *sync.WaitGroup
	Abort chan string
	Quit  chan struct{}
}
