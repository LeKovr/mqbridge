package nats_test

//go:generate mockgen -destination=generated_mock_test.go -package nats_test github.com/LeKovr/mqbridge/plugins/nats Server

import (
	"context"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	plugin "github.com/LeKovr/mqbridge/plugins/nats"
	"github.com/LeKovr/mqbridge/types"
)

func TestAll(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockEngine := NewMockServer(mockCtrl)

	ctx, cancel := context.WithCancel(context.Background())
	epa := types.NewBlankEndPointAttr(ctx)
	plug, err := plugin.NewConnected(epa, mockEngine)
	assert.NoError(t, err)
	assert.NotNil(t, plug)
	pipe := make(chan string)

	//	ch := make(chan *engine.Msg, nats.BufferSize)
	mockEngine.EXPECT().ChanSubscribe("testchannel", gomock.Any())
	mockEngine.EXPECT().Close()
	plug.Listen(0, "testchannel", pipe)
	plug.Notify(0, "testchannel", pipe)
	time.Sleep(100 * time.Millisecond)
	cancel()
	epa.WG.Wait()
}
