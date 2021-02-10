package mqbridge

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"

	"github.com/LeKovr/mqbridge/types"

	"github.com/go-logr/logr"
)

// -----------------------------------------------------------------------------

// Config holds all package config vars
type Config struct {
	BridgeDelimiter  string `long:"delim"       default:","       description:"Bridge definition delimiter"`
	PluginPathFormat string `long:"path_format" default:"./%s.so" description:"Plugin path format string"`

	EndPoints []string `long:"point"  default:"io:file"               description:"Endpoints connect string in form 'tag[:plugin[://dsn]]'"`
	Bridges   []string `long:"bridge" default:"io:src.txt,io:dst.txt" description:"Bridge def in form 'in_tag:in_channel:out_tag[:out_channel]'"`
}

// Bridge holds bridge attributes
type Bridge struct {
	InTag      string
	InChannel  string
	OutTag     string
	OutChannel string
	Pipe       chan string
}

// Service holds MQBridge service
type Service struct {
	log       logr.Logger
	cfg       *Config
	wg        *sync.WaitGroup
	abort     chan string     // aborted worker sends its name through it
	ctx       context.Context // used to stop all of workers
	cancel    func()
	EndPoints map[string]types.EndPoint
	Bridges   []*Bridge
}

// Errors
var (
	ErrNoPrefixDSN   = errors.New("no prefix")
	ErrNoTagDSN      = errors.New("no tag")
	ErrNoDelimBridge = errors.New("no delim")
	ErrNoEndPoint    = errors.New("no endpoint")

	ErrPluginBadNew  = errors.New("plugin has no correct New signature")
	ErrPluginUnknown = errors.New("plugin type is unknown")
)

const (
	// EndPointTypeFile holds name for rndpoint type file
	EndPointTypeFile = "file"

	// DSNPartsMin holds minimal DSN parts count
	DSNPartsMin = 2
	// DSNPartsConnect holds DSN parts count when connect string given
	DSNPartsConnect = 3
	// BridgePartsCount holds count of bridge delimited parts
	BridgePartsCount = 2
)

// New creates MQBridge service
func New(log logr.Logger, cfg *Config) (*Service, error) {
	var wg sync.WaitGroup
	srv := &Service{cfg: cfg, log: log, wg: &wg,
		EndPoints: make(map[string]types.EndPoint, len(cfg.EndPoints)),
		Bridges:   make([]*Bridge, len(cfg.Bridges)),
		abort:     make(chan string),
	}
	srv.ctx, srv.cancel = context.WithCancel(context.Background())
	for _, dsn := range cfg.EndPoints {
		tag, ep, err := srv.NewEndPoint(dsn)
		if err != nil {
			return nil, err
		}
		srv.EndPoints[tag] = ep
	}
	var err error
	for i, bridge := range cfg.Bridges {
		var br *Bridge
		br, err = NewBridge(cfg.BridgeDelimiter, bridge)
		if err != nil {
			break
		}
		if _, ok := srv.EndPoints[br.InTag]; !ok {
			err = ErrNoEndPoint
			break
		}
		if _, ok := srv.EndPoints[br.OutTag]; !ok {
			err = ErrNoEndPoint
			break
		}
		srv.Bridges[i] = br
	}
	return srv, err
}

// Run runs bridges
func (srv *Service) Run(quit chan os.Signal) (err error) {
	defer func() {
		srv.cancel()
		srv.wg.Wait() // Wait for side controls shutdown
		srv.log.Info("Service Exit")
	}()
	for i, br := range srv.Bridges {
		in := srv.EndPoints[br.InTag]
		out := srv.EndPoints[br.OutTag]
		err = in.Listen(i, br.InChannel, br.Pipe)
		if err != nil {
			break
		}
		err = out.Notify(i, br.OutChannel, br.Pipe)
		if err != nil {
			break
		}
	}
	if err != nil {
		return err
	}
	srv.log.Info("Service Ready")
	select {
	case <-quit:
		srv.log.Info("Interrupted")
		break
	case channel := <-srv.abort:
		srv.log.Info("Bridge aborted", "channel", channel)
		break
	}
	srv.log.Info("Service Exiting...")
	return nil
}

// NewEndPoint creates endpoint via plugin
func (srv *Service) NewEndPoint(dsn string) (string, types.EndPoint, error) {
	var tag, typ, connect string
	parts := strings.SplitN(dsn, ":", 3) // tag:mod://config
	tag = parts[0]
	if len(parts) < DSNPartsMin {
		typ = EndPointTypeFile
	} else {
		typ = parts[1]
		if len(parts) == DSNPartsConnect {
			connect = parts[2]
		}
	}
	ep, err := srv.LoadEndPoint(typ, connect)
	return tag, ep, err
}

// NewBridge creates bridge from definition string
func NewBridge(delim, bridge string) (*Bridge, error) {
	def := strings.Split(bridge, delim)
	if len(def) != BridgePartsCount {
		return nil, ErrNoDelimBridge
	}
	br := Bridge{Pipe: make(chan string)}
	parts := strings.SplitN(def[0], ":", 2)
	br.InTag = parts[0]
	br.InChannel = parts[1]
	parts = strings.SplitN(def[1], ":", 2)
	br.OutTag = parts[0]
	br.OutChannel = parts[1]
	if br.OutChannel == "" {
		br.OutChannel = br.InChannel
	}
	return &br, nil
}
