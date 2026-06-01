package ws

import (
	"context"
	"crypto_bot_parser/internal/constants"
	"crypto_bot_parser/pkg/helpers/backoff"
	"crypto_bot_parser/pkg/helpers/ws"
	"errors"
	"log"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptrace"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	utls "github.com/refraction-networking/utls"
)

var reconnectLimiter = time.Tick(5 * time.Millisecond)

type WSConn struct {
	id      int
	attemts int

	baseDelay time.Duration
	maxDelay  time.Duration

	conn         *websocket.Conn
	sid          string
	pingInterval time.Duration

	addCh chan<- *ListUpdateData

	userAgent string
	origin    string
	client    *http.Client
	helloID   utls.ClientHelloID
}

func NewWSConn(
	id int,
	addCh chan<- *ListUpdateData,
	userAgent string,
	origin string,
	client *http.Client,
) *WSConn {
	return &WSConn{
		id:        id,
		baseDelay: time.Second * 3,
		maxDelay:  time.Minute * 2,
		addCh:     addCh,
		userAgent: userAgent,
		origin:    origin,
		client:    client,
		helloID:   getHelloIDByUA(userAgent),
	}
}

func (w *WSConn) Run(ctx context.Context, upstreamBlockedUntil *atomic.Int64) {
	attempt := 0

	for {
		if ctx.Err() != nil {
			return
		}

		connectedAt := time.Now()
		err := w.connectAndServe(ctx, upstreamBlockedUntil)

		if errors.Is(err, context.Canceled) || err == nil {
			return
		}

		if ws.IsNormalWSClose(err) {
			slog.Debug("ws closed", slog.Any("err", err))
		} else {
			slog.Warn("ws error", slog.Any("err", err))
		}

		uptime := time.Since(connectedAt)
		if uptime > time.Minute*5 {
			attempt = 0
		}

		delay := backoff.Backoff(attempt, w.baseDelay, w.maxDelay)
		attempt++

		slog.Error(
			"failed to connect and serve websocket",
			slog.Any("error", err.Error()),
			slog.Duration("sleep for", time.Duration(delay)),
		)

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		select {
		case <-ctx.Done():
			return
		case <-reconnectLimiter:
		}
	}
}

func (w *WSConn) connectAndServe(ctx context.Context, upstreamBlockedUntil *atomic.Int64) error {
	trace := &httptrace.ClientTrace{
		GotConn: func(connInfo httptrace.GotConnInfo) {
			log.Println("Got connection:", connInfo.Conn.RemoteAddr())
		},
		GotFirstResponseByte: func() {
			log.Println("First byte received")
		},
	}

	cctx := httptrace.WithClientTrace(ctx, trace)

	dialer := websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		Jar:               w.client.Jar,
		EnableCompression: false,
		HandshakeTimeout:  10 * time.Second,
		NetDial:           createCustomNetDial(w.helloID),
	}

	header := http.Header{}
	header.Set("User-Agent", w.userAgent)
	header.Set("Origin", w.origin)
	header.Set("Accept-Language", "en-US,en;q=0.9")
	header.Set("Cache-Control", "no-cache")
	header.Set("Pragma", "no-cache")

	conn, _, err := dialer.DialContext(
		cctx,
		constants.CryptBotWsP2cURL,
		header,
	)
	if err != nil {
		if cerr, ok := err.(*websocket.CloseError); ok {
			slog.Error("websocket closed by server", slog.Any("code", cerr.Code), slog.String("text", cerr.Text))
		} else {
			slog.Error("websocket dial error", slog.Any("err", err))
		}
		return err
	}
	defer w.closeConn()

	w.conn = conn
	conn.SetReadLimit(64 * 1024)

	slog.Info("ws connected")

	conn.SetPongHandler(func(appData string) error {
		_ = conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(5*time.Second))
		return nil
	})

	return w.readLoop(ctx, conn, upstreamBlockedUntil)
}

func (w *WSConn) sendRaw(payload string) error {
	if w.conn == nil {
		return errors.New("ws not connected")
	}

	return w.conn.WriteMessage(
		websocket.TextMessage,
		[]byte(payload),
	)
}

func (w *WSConn) closeConn() {
	if w.conn != nil {
		_ = w.conn.Close()
		w.conn = nil
	}
}

func (w *WSConn) readLoop(ctx context.Context, conn *websocket.Conn, upstreamBlockedUntil *atomic.Int64) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return err
			}

			w.handleFrame(ctx, msg, upstreamBlockedUntil)
		}
	}
}

func createCustomNetDial(helloId utls.ClientHelloID) func(network, addr string) (net.Conn, error) {
	return func(network, addr string) (net.Conn, error) {
		baseConn, err := net.DialTimeout(network, addr, 10*time.Second)
		if err != nil {
			return nil, err
		}

		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			return nil, err
		}

		uTLSConn := utls.UClient(baseConn, &utls.Config{
			ServerName: host,
		}, helloId)

		if err := uTLSConn.Handshake(); err != nil {
			baseConn.Close()
			return nil, err
		}

		return uTLSConn, nil
	}
}

func getHelloIDByUA(ua string) utls.ClientHelloID {
	uaLower := strings.ToLower(ua)

	if strings.Contains(uaLower, "edg/") {
		return utls.HelloEdge_Auto
	}
	if strings.Contains(uaLower, "firefox") {
		return utls.HelloFirefox_Auto
	}
	if strings.Contains(uaLower, "safari") && !strings.Contains(uaLower, "chrome") {
		return utls.HelloSafari_Auto
	}

	return utls.HelloChrome_Auto
}
