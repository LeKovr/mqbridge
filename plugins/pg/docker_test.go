//+build docker

package pg_test

// This file used when internal 'docker run' is requested (see `make test-docker-self`)

import (
	"fmt"
	"testing"

	engine "github.com/go-pg/pg/v9"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dbName = "mqbridge"
	dsnFmt = "postgres://postgres:secret@localhost:%s/%s?sslmode=disable"
)

var (
	pgDockerImage = []string{"postgres", "13.1-alpine"}
	pgDockerArgs  = []string{"POSTGRES_PASSWORD=secret", "POSTGRES_DB=" + dbName}
)

func TestAll(t *testing.T) {
	pool, err := dockertest.NewPool("")
	assert.NoError(t, err)

	resource, err := pool.Run(pgDockerImage[0], pgDockerImage[1], pgDockerArgs)
	require.NoError(t, err)
	defer func() {
		_ = pool.Purge(resource)
	}()
	dsn := fmt.Sprintf(dsnFmt, resource.GetPort("5432/tcp"), dbName)
	opts, err := engine.ParseURL(dsn)
	err = pool.Retry(func() error {
		db := engine.Connect(opts)
		sql := fmt.Sprintf(funcSQLFmt, funcChannel, eventChannel)
		_, err = db.Exec(sql)
		if err != nil {
			return err
		}
		_ = db.Close()
		return nil
	})
	runTest(t, dsn)
	err = pool.Purge(resource)
	assert.NoError(t, err)
}
