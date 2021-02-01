package example_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wojas/genericr"

	plugin "github.com/LeKovr/mqbridge/plugins/example"
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
func Example_plugin() {
	epa := newEPA()
	plug, _ := plugin.New(epa, "test")
	pipe := make(chan string)
	plug.Listen("1:100", pipe)
	plug.Notify("", pipe)
	<-epa.Abort
	close(epa.Quit)
	epa.WG.Wait()
	// Output:
	// sample 0
}

func TestListenErrors(t *testing.T) {
	epa := newEPA()
	plug, err := plugin.New(epa, "test")
	assert.NoError(t, err)

	pipe := make(chan string)

	err = plug.Listen("1:BadDelay", pipe)
	assert.NotNil(t, err)
	assert.Equal(t, "delay: strconv.Atoi: parsing \"BadDelay\": invalid syntax", err.Error())

	err = plug.Listen("BadCount:", pipe)
	assert.NotNil(t, err)
	assert.Equal(t, "count: strconv.Atoi: parsing \"BadCount\": invalid syntax", err.Error())
}
