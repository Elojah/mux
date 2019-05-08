package mux

import (
	"github.com/golang/snappy"
)

// Middleware transforms and follow through data when sending and/or receiving.
type Middleware interface {
	Send([]byte) ([]byte, error)
	Receive([]byte) ([]byte, error)
}

// MwSnappy represents the snappy compression middleware.
type MwSnappy struct {
}

// Send represents the snappy compression message sending.
func (MwSnappy) Send(raw []byte) ([]byte, error) {
	return snappy.Encode(nil, raw), nil
}

// Receive represents the snappy compression message receiving.
func (m MwSnappy) Receive(raw []byte) ([]byte, error) {
	return snappy.Decode(nil, raw)
}
