package nats_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	plugin "github.com/LeKovr/mqbridge/plugins/nats"
	"github.com/LeKovr/mqbridge/types"
)

func TestAll(t *testing.T) {
	epa := types.NewBlankEndPointAttr()
	srv := Server{}
	_, err := plugin.NewConnected(epa, &srv)
	assert.NoError(t, err)
}
