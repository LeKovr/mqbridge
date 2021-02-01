package types_test

// This code does not used for testing (there is nothing to test here)
// It is intended for changing "[no test files]" to "coverage: [no statements]"

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/LeKovr/mqbridge/types"
)

type EndPoint struct{}

func New() (EndPoint, error)                                      { return EndPoint{}, nil }
func (ep EndPoint) Listen(channel string, pipe chan string) error { return nil }
func (ep EndPoint) Notify(channel string, pipe chan string) error { return nil }

func TestRun(t *testing.T) {
	plug, _ := New()
	_, ok := interface{}(plug).(types.EndPoint)
	assert.True(t, ok)
}
