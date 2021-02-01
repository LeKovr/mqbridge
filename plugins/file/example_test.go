package file_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wojas/genericr"

	plugin "github.com/LeKovr/mqbridge/plugins/file"
)

func TestAll(t *testing.T) {
	log := genericr.NewForTesting(t)
	var wg sync.WaitGroup
	abort := make(chan string)
	quit := make(chan struct{})
	plug, err := plugin.New(log, &wg, abort, quit, "test")
	assert.NoError(t, err)
	pipe := make(chan string)
	err = plug.Listen("file.go", pipe)
	assert.NoError(t, err)
	err = plug.Notify("-", pipe)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	close(quit)
	wg.Wait()
}

/*
func Example_plugin() {
	log := genericr.New(func(e genericr.Entry) {})
	var wg sync.WaitGroup
	abort := make(chan string)
	quit := make(chan struct{})
	plug, _ := plugin.New(log, &wg, abort, quit)
	pipe := make(chan string)
	plug.Listen("testdata/source.txt", pipe)
	plug.Notify("-", pipe)
	<-abort
	close(quit)
	wg.Wait()
	// Output:
	// sample 0
}
*/
func TestErrors(t *testing.T) {
	log := genericr.NewForTesting(t)
	var wg sync.WaitGroup
	abort := make(chan string)
	quit := make(chan struct{})
	plug, err := plugin.New(log, &wg, abort, quit, "test")
	assert.NoError(t, err)

	pipe := make(chan string)

	err = plug.Listen("...", pipe)
	assert.NotNil(t, err)
	assert.Equal(t, "open ...: no such file or directory", err.Error())

	err = plug.Notify("testdata/", pipe)
	assert.NotNil(t, err)
	assert.Equal(t, "open testdata/: is a directory", err.Error())
}
