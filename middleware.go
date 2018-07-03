package mux

import (
	"github.com/cloudflare/golz4"
)

// Middleware transforms and follow throught data when sending and/or receiving.
type Middleware interface {
	Send([]byte) ([]byte, error)
	Receive([]byte) ([]byte, error)
}

// Mwlz4 represents the lz4 compression middleware.
type Mwlz4 struct {
	PacketSize uint
}

// Send represents the lz4 compression message sending.
func (Mwlz4) Send(raw []byte) ([]byte, error) {
	out := make([]byte, lz4.CompressBound(raw))
	n, err := lz4.Compress(raw, out)
	return out[:n], err
}

// Receive represents the lz4 compression message receiving.
func (m Mwlz4) Receive(raw []byte) ([]byte, error) {
	out := make([]byte, m.PacketSize)
	err := lz4.Uncompress(raw, out)
	return out, err
}
