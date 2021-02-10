package file_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"

	plugin "github.com/LeKovr/mqbridge/plugins/file"
	"github.com/LeKovr/mqbridge/types"
)

const (
	TestRow0 = "test row one"
	TestRow1 = "test row two"
)

func TestAll(t *testing.T) {
	// Prepare bridge
	ctx, cancel := context.WithCancel(context.Background())
	epa := types.NewBlankEndPointAttr(ctx)
	plug, err := plugin.New(epa, "test")
	assert.NoError(t, err)
	pipe := make(chan string)
	// Prepare source with 1st data line
	in, err := ioutil.TempFile(".", "mqbridge-test-in-")
	assert.NoError(t, err)
	defer os.Remove(in.Name())
	in.WriteString(TestRow0 + "\n")
	// Start listening source
	err = plug.Listen(0, in.Name(), pipe)
	assert.NoError(t, err)
	// Prepare destination
	out, err := ioutil.TempFile(".", "mqbridge-test-out-")
	assert.NoError(t, err)
	defer os.Remove(out.Name())
	// Start outgoing endpoint
	err = plug.Notify(1, out.Name(), pipe)
	assert.NoError(t, err)

	// Write 2d data line
	in.WriteString(TestRow1 + "\n")

	// Give sime time to world (TODO: is there any way better?)
	time.Sleep(100 * time.Millisecond)
	// Fetch out and compare with in
	bytes, err := ioutil.ReadFile(out.Name())
	require.NoError(t, err)
	require.Equal(t, TestRow0+"\n"+TestRow1+"\n", string(bytes))
	// Close all
	cancel()
	epa.WG.Wait()
}

func Example_plugin() {
	ctx, cancel := context.WithCancel(context.Background())
	epa := types.NewBlankEndPointAttr(ctx)
	plug, _ := plugin.New(epa, "test")
	pipe := make(chan string)
	file, _ := ioutil.TempFile(".", "mqbridge-test-in")
	defer os.Remove(file.Name())
	file.WriteString(TestRow0 + "\n")
	plug.Listen(0, file.Name(), pipe)
	plug.Notify(0, "-", pipe)
	file.WriteString(TestRow1 + "\n")
	time.Sleep(100 * time.Millisecond)
	cancel()
	epa.WG.Wait()
	// Output:
	// test row one
	// test row two
}

func TestErrors(t *testing.T) {
	epa := types.NewBlankEndPointAttr(context.Background())
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
