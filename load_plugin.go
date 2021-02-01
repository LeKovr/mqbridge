//+build plugin

package mqbridge

import (
	"fmt"
	"plugin"
	"strings"
	"sync"

	"github.com/LeKovr/mqbridge/types"
	"github.com/go-logr/logr"
)

// LoadEndPoint loads endpoint from plugin
func (srv *Service) LoadEndPoint(typ, connect string) (types.EndPoint, error) {
	pathItems := []interface{}{typ}
	if strings.Count(srv.cfg.PluginPathFormat, "%") == 2 {
		pathItems = append(pathItems, typ)
	}
	mod := fmt.Sprintf(srv.cfg.PluginPathFormat, pathItems...)

	plug, err := plugin.Open(mod)
	if err != nil {
		return nil, err // errors.Wrap
	}

	symbol, err := plug.Lookup("New")
	if err != nil {
		return nil, err
	}
	NewEP, ok := symbol.(func(log logr.Logger, wg *sync.WaitGroup, abort chan string, quit chan struct{}, dsn string) (types.EndPoint, error))
	if !ok {
		return nil, ErrPluginBadNew
	}
	epa := types.EndPointAttr{srv.log, srv.wg, srv.abort, srv.quit}
	ep, err := NewEP(epa, connect)
	return ep, err
}
