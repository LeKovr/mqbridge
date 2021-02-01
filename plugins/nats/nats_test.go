package nats_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wojas/genericr"

	plugin "github.com/LeKovr/mqbridge/plugins/nats"
)

func TestAll(t *testing.T) {
	log := genericr.NewForTesting(t)
	var wg sync.WaitGroup
	abort := make(chan string)
	quit := make(chan struct{})
	srv := Server{}
	_, err := plugin.NewConnected(log, &wg, abort, quit, &srv)
	assert.NoError(t, err)
}
