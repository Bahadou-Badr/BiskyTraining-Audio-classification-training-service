package nats

import (
	"time"

	"github.com/nats-io/nats.go"
)

func Connect(url string) (*nats.Conn, error) {
	opts := []nats.Option{
		nats.Name("audioml-api"),
		nats.MaxReconnects(10),
		nats.ReconnectWait(2 * time.Second),
	}
	nc, err := nats.Connect(url, opts...)
	if err != nil {
		return nil, err
	}
	return nc, nil
}
