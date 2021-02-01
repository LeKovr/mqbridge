package pg_test

import (
	"sync"
	"testing"

	"github.com/go-pg/pg"
	"github.com/stretchr/testify/assert"
	"github.com/wojas/genericr"

	plugin "github.com/LeKovr/mqbridge/plugins/pg"
)

func TestAll(t *testing.T) {
	log := genericr.NewForTesting(t)
	var wg sync.WaitGroup
	abort := make(chan string)
	quit := make(chan struct{})
	db := pg.DB{}
	_, err := plugin.NewConnected(log, &wg, abort, quit, &db)
	assert.NoError(t, err)
}
