package forwarder

import (
	"log/slog"
	"net"
	"net/http"

	"github.com/stiletto/ws-mapper/contextids"
	"golang.org/x/net/websocket"
)

type Target struct {
	Address string `yaml:"address"`
	Family  string `yaml:"family"`
}

type WSForwarder struct {
	Config websocket.Config
	Target Target
}

func NewWSForwarder(target Target) *WSForwarder {
	fwd := &WSForwarder{Target: target}
	return fwd
}

func (h *WSForwarder) handshakeHandler(c *websocket.Config, r *http.Request) error {
	return nil
}

func (h *WSForwarder) connHandler(conn *websocket.Conn) {
	defer conn.Close()
	ctx := conn.Request().Context()
	logger, ok := ctx.Value(contextids.Logger).(*slog.Logger)
	if !ok {
		logger = slog.Default()
	}
	conn.PayloadType = websocket.BinaryFrame

	logger.Info("WebSocket established, connecting to target", "target", h.Target)
	targetConn, err := net.Dial(h.Target.Family, h.Target.Address)
	if err != nil {
		logger.Error("Failed to connect to target", "target", h.Target, "err", err)
		return
	}
	stop := make(chan bool, 2)
	// WebSocket protocol does not properly support half-open connections
	// so we don't bother either
	go func() {
		buf := make([]byte, 10<<10)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				logger.Error("client read error", "err", err)
				break
			}
			logger.Info("client read", "n", n, "data", buf[:n])
			n, err = targetConn.Write(buf[:n])
			if err != nil {
				logger.Error("target write error", "err", err)
				break
			}
			logger.Info("target write", "n", n)
		}
		stop <- true
	}()
	go func() {
		buf := make([]byte, 10<<10)
		for {
			n, err := targetConn.Read(buf)
			if err != nil {
				logger.Error("target read error", "err", err)
				break
			}
			logger.Info("target read", "n", n, "data", buf[:n])
			n, err = conn.Write(buf[:n])
			if err != nil {
				logger.Error("client write error", "err", err)
				break
			}
			logger.Info("client write", "n", n)
		}
		stop <- true
	}()
	<-stop

	defer targetConn.Close()

}

func (fwd *WSForwarder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	websocket.Server{
		Config:    fwd.Config,
		Handler:   fwd.connHandler,
		Handshake: fwd.handshakeHandler,
	}.ServeHTTP(w, r)
}

/*
func DialTCPorUnix(address string) (net.Conn, error) {
	if strings.HasPrefix(address, "unix:") {
		socketFileName := address[len("unix:"):]
		return net.Dial("unix", socketFileName)
	}
	return net.Dial("tcp", address)
}*/
