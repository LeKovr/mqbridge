package types

import (
	"github.com/LeKovr/go-base/log"
	"sync"
)

// Side holds interpocess communication for one side
type Side struct {
	WG        *sync.WaitGroup
	WGControl *sync.WaitGroup
	Log       log.Logger

	// Quit signalls to workes for exiting
	Quit chan struct{}

	// Abort used by workers for error reporting
	Abort chan int
}

// Bridge holds bridge attributes
type Bridge struct {
	// Bridge ID
	ID int

	// Source / dest channel name
	Channel string

	// Dest / source channel
	Pipe chan string //[]byte

}

// ListenFunc holds signature for Listen func which starts all listening goroutines
type ListenFunc func(side *Side, connect string, pairs []string) ([]*Bridge, error)

// NotifyFunc holds signature for Notify func which starts all notify goroutines
type NotifyFunc func(side *Side, connect string, bridges []*Bridge) error
