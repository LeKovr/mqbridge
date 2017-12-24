package main

import (
	"errors"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"

	"github.com/LeKovr/go-base/log"
	"github.com/jessevdk/go-flags"

	"github.com/LeKovr/mqbridge/plugins/file"
	"github.com/LeKovr/mqbridge/plugins/nats"
	"github.com/LeKovr/mqbridge/plugins/pg"

	"github.com/LeKovr/mqbridge/types"
)

// -----------------------------------------------------------------------------

// Flags defines local application flags
type Flags struct {
	In      string   `long:"in"          description:"Side 'in' connect string"`
	Out     string   `long:"out"         description:"Side 'out' connect string"`
	Bridges []string `long:"bridge"      description:"Bridge in form 'in_channel[:out_channel]'"`
	Version bool     `long:"version"     description:"Show version and exit"`
}

// Config holds all config vars
type Config struct {
	Flags
	Log LogConfig `group:"Logging Options"`
}

func main() {

	cfg, lg := setUp()
	lg.Printf("info: mqbridge v %s. Bridge from one MQ to another", Version)
	lg.Print("info: Copyright (C) 2017, Alexey Kovrizhkin <lekovr+mqbridge@gmail.com>")

	wg := &sync.WaitGroup{}
	sideIn := newSide(lg, wg)
	sideOut := newSide(lg, wg)

	typeIn, connectIn, err := parseDSN(cfg.In)
	panicIfError(lg, err, "Source")

	typeOut, connectOut, err := parseDSN(cfg.Out)
	panicIfError(lg, err, "Destination")

	// This code may be rewritten for golang plugins

	funcsIn := map[string]types.ListenFunc{}
	funcsOut := map[string]types.NotifyFunc{}
	funcsIn["file"] = file.Listen
	funcsOut["file"] = file.Notify
	funcsIn["pg"] = pg.Listen
	funcsOut["pg"] = pg.Notify
	funcsIn["nats"] = nats.Listen
	funcsOut["nats"] = nats.Notify

	// End of plugins code block

	bridges, err := funcsIn[typeIn](sideIn, connectIn, cfg.Bridges)
	panicIfError(lg, err, "Source init")

	err = funcsOut[typeOut](sideOut, connectOut, bridges)
	panicIfError(lg, err, "Destination init")

	lg.Println("Service Ready")

	// Wait for interrupt signal to gracefully shutdown the server with
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	select {
	case <-quit:
		lg.Println("Interrupted")
		break
	case id := <-sideIn.Abort:
		lg.Printf("Bridge %d side 'in' aborted", id)
		break
	case id := <-sideOut.Abort:
		lg.Printf("Bridge %d side 'out' aborted", id)
		break
	}
	lg.Println("Exiting...")

	close(sideIn.Quit) //Shutdown producers
	sideIn.WG.Wait()   //Wait for producers

	close(sideOut.Quit) //Close messages without causing a panic
	sideOut.WG.Wait()   //Wait for consumers to finish processing messages and exit

	wg.Wait() // Wait for sides shutdown

	lg.Println("Server exit")
}

func newSide(lg log.Logger, wg *sync.WaitGroup) *types.Side {
	return &types.Side{
		Abort:     make(chan int),
		Log:       lg,
		WGControl: wg,
		WG:        &sync.WaitGroup{},
		Quit:      make(chan struct{}),
	}

}

func parseDSN(dsn string) (typ, connect string, err error) {
	if strings.HasPrefix(dsn, "postgres://") {
		typ = "pg"
		connect = dsn
	} else if strings.HasPrefix(dsn, "nats://") {
		typ = "nats"
		connect = dsn
	} else if strings.HasPrefix(dsn, "file://") {
		typ = "file"
		connect = strings.TrimPrefix(dsn, "file://")
	} else {
		err = errors.New("Unsupported connect string: " + dsn)
	}
	return
}

// -----------------------------------------------------------------------------

func setUp() (cfg *Config, lg log.Logger) {
	cfg = &Config{}
	p := flags.NewParser(cfg, flags.Default)

	_, err := p.Parse()
	if err != nil {
		os.Exit(1) // error message written already
	}
	if cfg.Version {
		// show version & exit
		fmt.Printf("%s\n%s\n%s", Version, Build, Commit)
		os.Exit(0)
	}

	// use all CPU cores for maximum performance
	runtime.GOMAXPROCS(runtime.NumCPU())

	lg, err = NewLog(cfg.Log)
	panicIfError(nil, err, "Parse loglevel")
	return
}

// -----------------------------------------------------------------------------

func panicIfError(lg log.Logger, err error, msg string) {
	if err != nil {
		if lg != nil {
			lg.Printf("error: %s (%s)", msg, err.Error())
		} else {
			fmt.Printf("error: %s (%s)", msg, err.Error())
		}
		os.Exit(1)
	}
}

// -----------------------------------------------------------------------------

func panicf(lg log.Logger, format string, v ...interface{}) {
	lg.Printf("error: Error: "+format, v...)
	os.Exit(1)
}
