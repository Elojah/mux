package udp

import (
	"github.com/cloudflare/golz4"
)

// Middleware transforms and follow throught data when sending and/or receiving.
type Middleware interface {
	Send([]byte) ([]byte, error)
	Receive([]byte) ([]byte, error)
}

type mwlz4 struct {
	PacketSize uint
}

func (mwlz4) Send(raw []byte) ([]byte, error) {
	out := make([]byte, lz4.CompressBound(raw))
	n, err := lz4.Compress(raw, out)
	return out[:n], err
}

func (m mwlz4) Receive(raw []byte) ([]byte, error) {
	out := make([]byte, m.PacketSize)
	err := lz4.Uncompress(raw, out)
	return out, err
}
