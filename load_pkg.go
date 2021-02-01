//+build !plugin

package mqbridge

import (
	"github.com/LeKovr/mqbridge/types"

	"github.com/LeKovr/mqbridge/plugins/example"
	"github.com/LeKovr/mqbridge/plugins/file"
	"github.com/LeKovr/mqbridge/plugins/nats"
	"github.com/LeKovr/mqbridge/plugins/pg"
)

// LoadEndPoint loads endpoint from plugin
func (srv *Service) LoadEndPoint(typ, connect string) (types.EndPoint, error) {
	var (
		ep  types.EndPoint
		err error
	)
	epa := types.EndPointAttr{Log: srv.log, WG: srv.wg, Abort: srv.abort, Quit: srv.quit}
	switch typ {
	case "example":
		ep, err = example.New(epa, connect)
	case EndPointTypeFile:
		ep, err = file.New(epa, connect)
	case "nats":
		ep, err = nats.New(epa, connect)
	case "pg":
		ep, err = pg.New(epa, connect)
	default:
		err = ErrPluginUnknown
	}
	return ep, err
}
