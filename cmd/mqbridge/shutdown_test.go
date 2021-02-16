package main_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	cmd "github.com/LeKovr/mqbridge/cmd/mqbridge"
)

func TestShutdown(t *testing.T) {
	err := errors.New("unknown")
	logRows := []string{}
	hook := func(e zapcore.Entry) error {
		logRows = append(logRows, e.Message)
		return nil
	}
	l := cmd.SetupLog(false, zap.Hooks(hook))
	var c int

	cmd.Shutdown(func(code int) { c = code }, err, l)
	assert.Equal(t, 1, c)
	assert.Equal(t, []string{"Run error"}, logRows)
}
