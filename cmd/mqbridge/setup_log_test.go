package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	cmd "github.com/LeKovr/mqbridge/cmd/mqbridge"
)

func TestSetupLog(t *testing.T) {
	tests := []struct {
		name     string
		debug    bool
		wantRows []string
	}{
		{"Debug", true, []string{"debug", "info", "error"}},
		{"NoDebug", false, []string{"info", "error"}},
	}
	for _, tt := range tests {
		logRows := []string{}
		hook := func(e zapcore.Entry) error {
			logRows = append(logRows, e.Message)
			return nil
		}
		l := cmd.SetupLog(tt.debug, zap.Hooks(hook))
		l.V(1).Info("debug")
		l.Info("info")
		l.Error(nil, "error")
		assert.Equal(t, tt.wantRows, logRows)
	}
}
