package example_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	plugin "github.com/LeKovr/mqbridge/plugins/example"
	"github.com/LeKovr/mqbridge/types"
)

func Example_plugin() {
	epa := types.NewBlankEndPointAttr()
	plug, _ := plugin.New(epa, "test")
	pipe := make(chan string)
	plug.Listen(0, "1:100", pipe)
	plug.Notify(0, "", pipe)
	<-epa.Abort
	close(epa.Quit)
	epa.WG.Wait()
	// Output:
	// sample 0
}

func TestListenErrors(t *testing.T) {
	epa := types.NewBlankEndPointAttr()
	plug, err := plugin.New(epa, "test")
	assert.NoError(t, err)

	pipe := make(chan string)

	err = plug.Listen(0, "1:BadDelay", pipe)
	assert.NotNil(t, err)
	assert.Equal(t, "delay: strconv.Atoi: parsing \"BadDelay\": invalid syntax", err.Error())

	err = plug.Listen(0, "BadCount:", pipe)
	assert.NotNil(t, err)
	assert.Equal(t, "count: strconv.Atoi: parsing \"BadCount\": invalid syntax", err.Error())
}
