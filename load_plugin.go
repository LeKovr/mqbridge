//+build plugin

package mqbridge

import (
	"fmt"
	"plugin"
	"strings"

	"github.com/LeKovr/mqbridge/types"
)

// LoadEndPoint loads endpoint from plugin
func (srv *Service) LoadEndPoint(typ, connect string) (types.EndPoint, error) {
	pathItems := []interface{}{typ}
	if strings.Count(srv.cfg.PluginPathFormat, "%") == 2 {
		pathItems = append(pathItems, typ)
	}
	mod := fmt.Sprintf(srv.cfg.PluginPathFormat, pathItems...)
	srv.log.Info("Loading plugin", "plugin", mod)
	plug, err := plugin.Open(mod)
	if err != nil {
		return nil, err // errors.Wrap
	}

	symbol, err := plug.Lookup("New")
	if err != nil {
		return nil, err
	}
	NewEP, ok := symbol.(func(epa types.EndPointAttr, dsn string) (types.EndPoint, error))
	if !ok {
		return nil, ErrPluginBadNew
	}
	epa := types.EndPointAttr{Log: srv.log, WG: srv.wg, Abort: srv.abort, Ctx: srv.ctx}
	ep, err := NewEP(epa, connect)
	return ep, err
}
