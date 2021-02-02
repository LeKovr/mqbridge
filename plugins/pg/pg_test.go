package pg_test

import (
	"testing"

	"github.com/go-pg/pg/v9"
	"github.com/stretchr/testify/assert"

	plugin "github.com/LeKovr/mqbridge/plugins/pg"
	"github.com/LeKovr/mqbridge/types"
)

func TestAll(t *testing.T) {
	epa := types.NewBlankEndPointAttr()
	db := pg.DB{}
	_, err := plugin.NewConnected(epa, &db)
	assert.NoError(t, err)
}
