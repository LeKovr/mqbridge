package main_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	cmd "github.com/LeKovr/mqbridge/cmd/mqbridge"
)

func TestSetupConfig(t *testing.T) {
	cfg, err := cmd.SetupConfig("--debug")
	assert.NoError(t, err)
	assert.NotNil(t, cfg)
}
