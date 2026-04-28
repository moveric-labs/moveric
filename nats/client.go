package natsclient

import (
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
)

const (
	StreamName        = "MOVERIC"
	SubjectChunkReady = "moveric.chunk.ready"
	SubjectChunkAck   = "moveric.chunk.ack"
)

type Client struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func New(url string) (*Client, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("nats connect: %w", err)
	}
	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("nats jetstream: %w", err)
	}
	_, err = js.StreamInfo(StreamName)
	if err != nil {
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     StreamName,
			Subjects: []string{"moveric.>"},
		})
		if err != nil {
			return nil, fmt.Errorf("nats add stream: %w", err)
		}
	}
	return &Client{nc: nc, js: js}, nil
}

func (c *Client) Publish(subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = c.js.Publish(subject, data)
	return err
}

func (c *Client) Subscribe(subject, consumer string, handler func([]byte) error) error {
	_, err := c.js.Subscribe(subject, func(msg *nats.Msg) {
		if err := handler(msg.Data); err != nil {
			msg.Nak()
			return
		}
		msg.Ack()
	}, nats.Durable(consumer), nats.ManualAck())
	return err
}

func (c *Client) Close() {
	c.nc.Drain()
}
