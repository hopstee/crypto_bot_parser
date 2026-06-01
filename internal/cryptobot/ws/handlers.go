package ws

import (
	"context"
	"crypto_bot_parser/pkg/helpers/ws"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
)

type EventType string

const (
	ListUpdate EventType = "list:update"
)

type ListUpdateMessage struct {
	Op   string         `json:"op"`
	Data ListUpdateData `json:"data"`
}

type ListUpdateData struct {
	ID           string `json:"id"`
	Payload      string `json:"payload"`
	URL          string `json:"url"`
	BrandName    string `json:"brand_name"`
	InAsset      string `json:"in_asset"`
	OutAsset     string `json:"out_asset"`
	Provider     string `json:"provider"`
	InAmount     string `json:"in_amount"`
	OutAmount    string `json:"out_amount"`
	FeeAmount    string `json:"fee_amount"`
	ExchangeRate string `json:"exchange_rate"`
	ExpiresAt    string `json:"expires_at"`
}

type SocketEvent struct {
	Name EventType
	Data json.RawMessage
}

func (w *WSConn) handleFrame(ctx context.Context, msg []byte, upstreamBlockedUntil *atomic.Int64) {
	socketEvent, err := w.parseSocketMessage(msg, upstreamBlockedUntil)
	if err != nil {
		return
	}

	switch socketEvent.Name {
	case ListUpdate:
		w.handleP2CListUpdate(ctx, socketEvent.Data)
	default:
	}
}

func (w *WSConn) handleP2CListUpdate(ctx context.Context, data json.RawMessage) error {
	var events []ListUpdateMessage
	if err := json.Unmarshal(data, &events); err != nil {
		return fmt.Errorf("failed unmarshal websocket event: %w", err)
	}

	for _, e := range events {
		if e.Op == "add" {
			select {
			case w.addCh <- &e.Data:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}

func (w *WSConn) parseSocketMessage(msg []byte, upstreamBlockedUntil *atomic.Int64) (*SocketEvent, error) {
	msg, err := ws.DecodeIfBase64(msg)
	if err != nil {
		return &SocketEvent{}, fmt.Errorf("failed parse ws msg: %w", err)
	}

	if len(msg) == 0 {
		return &SocketEvent{}, nil
	}

	switch msg[0] {
	case '0':
		return w.handleOpenConnection(msg)
	case '2':
		return w.handlePing()
	case '4':
		return w.handleMessage(msg, upstreamBlockedUntil)
	}

	return &SocketEvent{}, nil
}

func (w *WSConn) handleOpenConnection(msg []byte) (*SocketEvent, error) {
	var open struct {
		SID          string `json:"sid"`
		PingInterval int    `json:"pingInterval"`
		PingTimeout  int    `json:"pingTimeout"`
	}
	if err := json.Unmarshal(msg[1:], &open); err != nil {
		return &SocketEvent{}, err
	}

	w.sid = open.SID
	w.pingInterval = time.Duration(open.PingInterval) * time.Millisecond

	_ = w.sendRaw("40")

	return &SocketEvent{}, nil
}

func (w *WSConn) handlePing() (*SocketEvent, error) {
	_ = w.sendRaw("3")
	return &SocketEvent{}, nil
}

func (w *WSConn) handleMessage(msg []byte, upstreamBlockedUntil *atomic.Int64) (*SocketEvent, error) {
	if len(msg) < 2 {
		return &SocketEvent{}, nil
	}

	switch msg[1] {
	case '0':
		_ = w.sendRaw(`42["list:initialize"]`)
	case '2':
		if !w.canHandleHandleMessage(upstreamBlockedUntil) {
			return &SocketEvent{}, nil
		}

		var arr []json.RawMessage
		if err := json.Unmarshal(msg[2:], &arr); err != nil {
			return &SocketEvent{}, err
		}

		if len(arr) != 2 {
			return &SocketEvent{}, errors.New("invalid socket.io event format")
		}

		if string(arr[0]) != `"list:update"` {
			return &SocketEvent{}, nil
		}

		return &SocketEvent{
			Name: ListUpdate,
			Data: arr[1],
		}, nil
	}

	return &SocketEvent{}, nil
}

func (w *WSConn) canHandleHandleMessage(upstreamBlockedUntil *atomic.Int64) bool {
	if until := upstreamBlockedUntil.Load(); until > 0 {
		sleep := time.Until(time.Unix(0, until))
		if sleep > 0 {
			slog.Info("message handling temporary stopped", slog.Duration("duration", sleep))
			return false
		}
	}
	return true
}
