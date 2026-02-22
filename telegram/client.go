package telegram

import (
	"log"
	"sync"

	"github.com/gotd/contrib/bg"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type status string

const (
	statusIdle = "idle"
	statusBusy = "busy"
)

func (s status) IsIdle() bool {
	return s == statusIdle
}

func (s status) IsBusy() bool {
	return s == statusBusy
}

func (s status) SetBusy() {
	s = statusBusy
}

func (s status) SetIdle() {
	s = statusIdle
}

type Client struct {
	mu sync.Mutex

	uuid   string
	status status
	client *telegram.Client
	stop   bg.StopFunc
}

func NewClient(uuid string) *Client {
	client := &Client{
		uuid:   uuid,
		status: status(statusIdle),
	}

	invokeMiddleware := NewInvokeMiddleware(
		func() {
			client.mu.Lock()
			client.status.SetBusy()
		},
		func() {
			client.status.SetIdle()
			client.mu.Unlock()
		},
	)

	client.client = telegram.NewClient(AppID, AppHash, telegram.Options{
		SessionStorage: GetSessionStorage(uuid),
		Middlewares:    []telegram.Middleware{invokeMiddleware},
		// Resolver: dcs.Plain(dcs.PlainOptions{
		// 	Dial: dialer,
		// }),
	})

	stop, err := bg.Connect(client.client)
	if err != nil {
		log.Printf("failed to connect to telegram: %v", err)
		return nil
	}
	client.stop = stop

	return client
}

func (c *Client) Stop() {
	c.stop()
}

func (c *Client) IsIdle() bool {
	return c.status.IsIdle()
}

func (c *Client) IsBusy() bool {
	return c.status.IsBusy()
}

func (c *Client) API() *tg.Client {
	return c.client.API()
}
