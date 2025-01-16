package shared

import "github.com/nats-io/nats.go"

type NatsConnection struct {
	Nc      *nats.Conn
	Js      *nats.JetStreamContext
	Kv		nats.KeyValue
	Channel string
}