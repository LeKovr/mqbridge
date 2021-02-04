//+build !docker

package pg_test

// These tests runs when TEST_DSN_PG ENV var holds DSN to postgresql

import (
	"fmt"
	"os"
	"testing"
	"time"

	engine "github.com/go-pg/pg/v9"
	"github.com/stretchr/testify/require"

	plugin "github.com/LeKovr/mqbridge/plugins/pg"
)

const PgDsnEnv = "TEST_DSN_PG"

func TestAll(t *testing.T) {
	dsn := os.Getenv(PgDsnEnv)
	if dsn == "" {
		t.Skip("Skipping testing when no PG DSN given")
	}
	opts, err := engine.ParseURL(dsn)
	require.NoError(t, err)

	db := engine.Connect(opts)
	sql := fmt.Sprintf(funcSQLFmt, funcChannel, eventChannel)
	for i := 0; i < plugin.PgMaxRetries; i++ {
		_, err = db.Exec(sql)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	_ = db.Close()
	require.NoError(t, err)
	runTest(t, dsn)
}
