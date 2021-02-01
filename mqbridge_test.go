package mqbridge_test

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wojas/genericr"

	"github.com/LeKovr/mqbridge"
)

func TestRun(t *testing.T) {
	log := genericr.New(func(e genericr.Entry) {
		t.Log(e.String())
	})
	cfg := mqbridge.Config{
		BridgeDelimiter:  ",",
		PluginPathFormat: "./%s.so",
		//		PluginPathFormat: "./plugins/%s/%s.so",

		EndPoints: []string{"io:example:"},
		Bridges:   []string{"io:5:10,io:"},
	}

	mqbr, err := mqbridge.New(log, &cfg)
	assert.NoError(t, err)

	quit := make(chan os.Signal, 1)
	go func() {
		time.Sleep(1 * time.Second)
		close(quit)
	}()
	err = mqbr.Run(quit)
	assert.NoError(t, err)
}
