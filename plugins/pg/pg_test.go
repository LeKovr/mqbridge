package pg_test

import (
	"testing"

	engine "github.com/go-pg/pg/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	plugin "github.com/LeKovr/mqbridge/plugins/pg"
	"github.com/LeKovr/mqbridge/types"
)

const (
	notExistedDSN = "postgres://@:-1/test"

	funcSQLFmt = `create or replace function %s(a text) returns void language plpgsql as
$_$  BEGIN PERFORM pg_notify('%s', a); END $_$;`
	funcChannel  = "test_channel_plugin"
	eventChannel = "test_event_plugin"
	testData     = "message Row"
)

func TestNewError(t *testing.T) {
	epa := types.NewBlankEndPointAttr()
	_, err := plugin.New(epa, notExistedDSN)
	assert.Error(t, err)
}
func TestNewConnected(t *testing.T) {
	epa := types.NewBlankEndPointAttr()
	db := engine.DB{}
	plug, err := plugin.NewConnected(epa, &db)
	assert.NoError(t, err)
	assert.NotNil(t, plug)
	close(epa.Quit)
	epa.WG.Wait()
}

// runTest used in docker_test & env_test
func runTest(t *testing.T, dsn string) {
	pipeIn := make(chan string)
	pipeOut := make(chan string)

	epa := types.NewBlankEndPointAttr()
	plug, err := plugin.New(epa, dsn)
	require.NoError(t, err)
	require.NotNil(t, plug)

	plug.Listen(0, eventChannel, pipeOut)
	plug.Notify(0, funcChannel, pipeIn)
	pipeIn <- testData
	got := <-pipeOut
	require.Equal(t, testData, got)

	close(epa.Quit)
	epa.WG.Wait()
	require.NoError(t, err)
}
