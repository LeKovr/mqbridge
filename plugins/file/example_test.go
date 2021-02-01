package file_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	plugin "github.com/LeKovr/mqbridge/plugins/file"
	"github.com/LeKovr/mqbridge/types"
)

func TestAll(t *testing.T) {
	epa := types.NewBlankEndPointAttr()
	plug, err := plugin.New(epa, "test")
	assert.NoError(t, err)
	pipe := make(chan string)
	err = plug.Listen(0, "file.go", pipe)
	assert.NoError(t, err)
	err = plug.Notify(0, "-", pipe)
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
	epa := types.NewBlankEndPointAttr()
	plug, err := plugin.New(epa, "test")
	assert.NoError(t, err)

	pipe := make(chan string)

	err = plug.Listen(0, "...", pipe)
	assert.NotNil(t, err)
	assert.Equal(t, "open ...: no such file or directory", err.Error())

	err = plug.Notify(0, "testdata/", pipe)
	assert.NotNil(t, err)
	assert.Equal(t, "open testdata/: is a directory", err.Error())
}
