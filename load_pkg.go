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
	switch typ {
	case "example":
		ep, err = example.New(srv.log, srv.wg, srv.abort, srv.quit, connect)
	case "file":
		ep, err = file.New(srv.log, srv.wg, srv.abort, srv.quit, connect)
	case "nats":
		ep, err = nats.New(srv.log, srv.wg, srv.abort, srv.quit, connect)
	case "pg":
		ep, err = pg.New(srv.log, srv.wg, srv.abort, srv.quit, connect)
	default:
		err = ErrPluginUnknown
	}
	return ep, err
}
