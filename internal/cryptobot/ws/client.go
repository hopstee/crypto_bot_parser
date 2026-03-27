package ws

import (
	"context"
	"crypto_bot_parser/internal/constants"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

type TimedEvent[T any] struct {
	Data T

	ReceivedAt      time.Time
	CallbackAt      time.Time
	TakeDescisionAt time.Time
	TakeSendAt      time.Time
}

type Client struct {
	client *http.Client

	userAgent string
	origin    string

	connectsToOpen int

	addCh chan *ListUpdateData
	onAdd func(*ListUpdateData)

	ctx    context.Context
	cancel context.CancelFunc

	wg sync.WaitGroup
}

// NewClient creates a new WS instance.
// Parameters:
//
//	c: http client for serving requests
//	ua: user agent
//	h: callback function
func NewClient(
	client *http.Client,
	userAgent string,
	connectsToOpen int,
	addHandler func(*ListUpdateData),
) *Client {
	if connectsToOpen <= 0 {
		connectsToOpen = 1
	}

	c := &Client{
		client: client,

		userAgent: userAgent,
		origin:    constants.CryptBotOrigin,

		connectsToOpen: connectsToOpen,

		addCh: make(chan *ListUpdateData, 1024),
		onAdd: addHandler,
	}

	return c
}

func (c *Client) Run(ctx context.Context, upstreamBlockedUntil *atomic.Int64) {
	go c.readAddEvent()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for i := range c.connectsToOpen {
		<-ticker.C

		c.wg.Add(1)
		conn := NewWSConn(i, c.addCh, c.userAgent, c.origin, c.client)

		go func() {
			defer c.wg.Done()
			conn.Run(ctx, upstreamBlockedUntil)
		}()
	}

	c.wg.Wait()
}

func (c *Client) readAddEvent() {
	for data := range c.addCh {
		c.onAdd(data)
	}
}
