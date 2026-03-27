package ws

import (
	"context"
	"crypto_bot_parser/internal/constants"
	"crypto_bot_parser/pkg/helpers/backoff"
	"crypto_bot_parser/pkg/helpers/ws"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptrace"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
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

		time.Sleep(delay)
		<-reconnectLimiter
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
	}

	header := http.Header{}
	header.Set("User-Agent", w.userAgent)
	header.Set("Origin", w.origin)

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

	if w.addCh == nil {
		close(w.addCh)
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

			w.handleFrame(msg, upstreamBlockedUntil)
		}
	}
}
