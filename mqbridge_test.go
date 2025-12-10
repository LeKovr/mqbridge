package mqbridge_test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/go-logr/logr"
	engine "github.com/go-pg/pg/v10"
	"github.com/stretchr/testify/require"

	"github.com/LeKovr/mqbridge"
)

const (
	PgDsnEnv   = "TEST_DSN_PG"
	NatsDsnEnv = "TEST_DSN_NATS"

	funcChannel  = "test_channel"
	eventChannel = "test_event"
	testData     = "Message Row"

	funcSQL = `create or replace function test_channel(a text) returns void language plpgsql as
	$_$  BEGIN PERFORM pg_notify('test_event', a); END $_$;`

	// PgMaxRetries holds Pg connect retry count
	pgMaxRetries = 5
)

func TestRunExample(t *testing.T) {
	log := logr.Discard()
	cfg := mqbridge.Config{
		BridgeDelimiter:  ",",
		PluginPathFormat: "./%s.so",
		EndPoints:        []string{"io:example:"},
		Bridges:          []string{"io:5:10,io:"},
	}

	mqbr, err := mqbridge.New(log, &cfg)
	require.NoError(t, err)

	quit := make(chan os.Signal, 1)
	go func() {
		time.Sleep(1 * time.Second)
		close(quit)
	}()
	err = mqbr.Run(quit)
	require.NoError(t, err)
}

func TestRunPlugins(t *testing.T) {
	log := logr.Discard()
	cfg := mqbridge.Config{
		BridgeDelimiter:  ",",
		PluginPathFormat: "./%s.so",
	}
	fileIn, _ := ioutil.TempFile(".", "mqbridge-test-in-")
	defer os.Remove(fileIn.Name())

	eps := []string{"io:file:"}
	chains := []string{"io:" + fileIn.Name()}

	pgDSN := os.Getenv(PgDsnEnv)
	if pgDSN != "" {
		eps = append(eps, "pg:pg:"+pgDSN)
		chains = append(chains, "pg:"+funcChannel, "pg:"+eventChannel)
		err := setupDB(pgDSN)
		require.NoError(t, err)
	}
	natsDSN := os.Getenv(NatsDsnEnv)
	if natsDSN != "" {
		eps = append(eps, "mq:nats:"+natsDSN)
		chains = append(chains, "mq:"+funcChannel, "mq:"+funcChannel)
	}

	fileOut, _ := ioutil.TempFile(".", "mqbridge-test-out-")
	defer os.Remove(fileOut.Name())
	chains = append(chains, "io:"+fileOut.Name())
	brs := []string{}
	for i := 0; i < len(chains); i += 2 {
		brs = append([]string{chains[i] + cfg.BridgeDelimiter + chains[i+1]}, brs...)
	}
	log.Info("mqbridge config", "points", eps, "bridges", brs)
	cfg.EndPoints = eps
	cfg.Bridges = brs

	mqbr, err := mqbridge.New(log, &cfg)
	require.NoError(t, err)

	quit := make(chan os.Signal, 1)
	fileIn.WriteString(testData + "\n")
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(quit)
	}()
	err = mqbr.Run(quit)
	require.NoError(t, err)
	time.Sleep(100 * time.Millisecond)
	// Fetch out and compare with in
	bytes, err := ioutil.ReadFile(fileOut.Name())
	require.NoError(t, err)
	require.Equal(t, testData+"\n", string(bytes))
}

func setupDB(dsn string) error {
	opts, err := engine.ParseURL(dsn)
	if err != nil {
		return err
	}
	db := engine.Connect(opts)
	for i := 0; i < pgMaxRetries; i++ {
		_, err = db.Exec(funcSQL)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	_ = db.Close()
	return err
}
