package file_test

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wojas/genericr"

	plugin "github.com/LeKovr/mqbridge/plugins/file"
	"github.com/LeKovr/mqbridge/types"
)

func newEPA() types.EndPointAttr {
	var wg sync.WaitGroup
	return types.EndPointAttr{
		Log:   genericr.New(func(e genericr.Entry) {}),
		WG:    &wg,
		Abort: make(chan string),
		Quit:  make(chan struct{}),
	}
}

func TestAll(t *testing.T) {
	epa := newEPA()
	plug, err := plugin.New(epa, "test")
	assert.NoError(t, err)
	pipe := make(chan string)
	err = plug.Listen("file.go", pipe)
	assert.NoError(t, err)
	err = plug.Notify("-", pipe)
	assert.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	close(epa.Quit)
	epa.WG.Wait()
}

/*
func Example_plugin() {
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
	epa := newEPA()
	plug, err := plugin.New(epa, "test")
	assert.NoError(t, err)

	pipe := make(chan string)

	err = plug.Listen("...", pipe)
	assert.NotNil(t, err)
	assert.Equal(t, "open ...: no such file or directory", err.Error())

	err = plug.Notify("testdata/", pipe)
	assert.NotNil(t, err)
	assert.Equal(t, "open testdata/: is a directory", err.Error())
}
