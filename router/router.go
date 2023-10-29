package router

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"regexp"

	"github.com/stiletto/ws-mapper/contextids"
)

type Router struct {
	opts RouterOpts

	regexps []*regexp.Regexp
}

type RouterOpts struct {
	Logger *slog.Logger
	Routes []Route
}

type Route struct {
	Match   string
	Handler http.Handler
}

func NewRouter(opts RouterOpts) (*Router, error) {
	handler := &Router{
		opts:    opts,
		regexps: make([]*regexp.Regexp, len(opts.Routes)),
	}
	if handler.opts.Logger == nil {
		handler.opts.Logger = slog.Default()
	}
	for i, route := range opts.Routes {
		var err error
		handler.regexps[i], err = regexp.Compile(route.Match)
		if err != nil {
			return nil, fmt.Errorf("Route %d: %w", i, err)
		}
	}
	return handler, nil
}

func (h *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := h.opts.Logger.With("url", r.URL.Path)
	for i, route := range h.regexps {
		if route.Match([]byte(r.URL.Path)) {
			logger := logger.With("route", i)
			logger.Info("Matched")
			ctx := r.Context()
			ctx = context.WithValue(ctx, contextids.Logger, logger)
			r = r.WithContext(ctx)
			h.opts.Routes[i].Handler.ServeHTTP(w, r)
			logger.Info("Closing connection")
			return
		}
	}
	logger.Info("Not found")
	w.WriteHeader(404)
	w.Write([]byte("Route not found\n"))
}

var _ http.Handler = &Router{}
