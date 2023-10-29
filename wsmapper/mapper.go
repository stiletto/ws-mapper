package wsmapper

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	"github.com/stiletto/ws-mapper/forwarder"
	"github.com/stiletto/ws-mapper/router"
)

type WSMapper struct {
	listener net.Listener
	server   http.Server
	router   *router.Router
}

func NewWSMapper(config *Config, logger *slog.Logger) (*WSMapper, error) {
	wsm := &WSMapper{}
	var err error
	routes := make([]router.Route, len(config.Routes))
	for i, route := range config.Routes {
		routes[i] = router.Route{
			Match:   route.Match,
			Handler: forwarder.NewWSForwarder(route.Target),
		}
	}
	wsm.router, err = router.NewRouter(router.RouterOpts{Logger: logger, Routes: routes})
	if err != nil {
		return nil, err
	}

	wsm.server.Handler = wsm.router

	wsm.listener, err = net.Listen(config.Listen.Family, config.Listen.Address)
	if err != nil {
		return nil, err
	}
	return wsm, nil
}

func (wsm *WSMapper) Serve() error {
	return wsm.server.Serve(wsm.listener)
}

func (wsm *WSMapper) Close() {
	wsm.listener.Close()
	wsm.server.Close()
}

func (wsm *WSMapper) Shutdown(ctx context.Context) error {
	defer wsm.listener.Close()
	return wsm.server.Shutdown(ctx)
}
