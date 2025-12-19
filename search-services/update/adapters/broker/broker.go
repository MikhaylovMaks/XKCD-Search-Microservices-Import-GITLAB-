package broker

import (
	"log/slog"

	"github.com/nats-io/nats.go"
)

type Client struct {
	nc   *nats.Conn
	log  *slog.Logger
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	nc, err := nats.Connect(address)
	if err != nil {
		return nil, err
	}

	log.Info("connected to broker", "address", address)
	return &Client{
		nc:  nc,
		log: log,
	}, nil
}

func (c *Client) Close() error {
	if c.nc != nil {
		c.nc.Close()
	}
	return nil
}

func (c *Client) Publish(topic string, data []byte) error {
	return c.nc.Publish(topic, data)
}
