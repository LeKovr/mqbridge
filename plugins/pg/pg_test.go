package pg_test

import (
	"sync"
	"testing"

	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/wojas/genericr"

	plugin "github.com/LeKovr/mqbridge/plugins/pg"
	"github.com/LeKovr/mqbridge/types"
)

func TestAll(t *testing.T) {
	epa := newEPA()
	db := pg.DB{}
	_, err := plugin.NewConnected(epa, &db)
	assert.NoError(t, err)
}

func newEPA() types.EndPointAttr {
	var wg sync.WaitGroup
	return types.EndPointAttr{
		Log:   genericr.New(func(e genericr.Entry) {}),
		WG:    &wg,
		Abort: make(chan string),
		Quit:  make(chan struct{}),
	}
}
