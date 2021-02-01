package example_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wojas/genericr"

	plugin "github.com/LeKovr/mqbridge/plugins/example"
)

func Example_plugin() {
	log := genericr.New(func(e genericr.Entry) {})
	var wg sync.WaitGroup
	abort := make(chan string)
	quit := make(chan struct{})
	plug, _ := plugin.New(log, &wg, abort, quit, "test")
	pipe := make(chan string)
	plug.Listen("1:100", pipe)
	plug.Notify("", pipe)
	<-abort
	close(quit)
	wg.Wait()
	// Output:
	// sample 0
}

func TestListenErrors(t *testing.T) {
	log := genericr.NewForTesting(t)
	var wg sync.WaitGroup
	abort := make(chan string)
	quit := make(chan struct{})
	plug, err := plugin.New(log, &wg, abort, quit, "test")
	assert.NoError(t, err)

	pipe := make(chan string)

	err = plug.Listen("1:BadDelay", pipe)
	assert.NotNil(t, err)
	assert.Equal(t, "delay: strconv.Atoi: parsing \"BadDelay\": invalid syntax", err.Error())

	err = plug.Listen("BadCount:", pipe)
	assert.NotNil(t, err)
	assert.Equal(t, "count: strconv.Atoi: parsing \"BadCount\": invalid syntax", err.Error())
}
