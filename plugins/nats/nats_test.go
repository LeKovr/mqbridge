package nats_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wojas/genericr"

	plugin "github.com/LeKovr/mqbridge/plugins/nats"
	"github.com/LeKovr/mqbridge/types"
)

func TestAll(t *testing.T) {
	epa := newEPA()
	srv := Server{}
	_, err := plugin.NewConnected(epa, &srv)
	assert.NoError(t, err)
}

func newEPA() types.EndPointAttr {
	var wg sync.WaitGroup
	return types.EndPointAttr{
		Log:   genericr.New(func(e genericr.Entry) {}),
		WG:    &wg,
		Abort: make(chan string),
		Quit:  make(chan struct{}),
	}
}
