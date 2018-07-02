package mux

import (
	"math/big"
	"net"
)

// Server wraps an UDP connection.
type Server struct {
	Conns []net.PacketConn
	NConn *big.Int
}

// Dial init the UDP server.
func (s *Server) Dial(c Config) error {
	var err error
	s.Conns = make([]net.PacketConn, len(c.Addresses))
	for i, address := range c.Addresses {
		if s.Conns[i], err = net.ListenPacket("udp", address); err != nil {
			return err
		}
	}
	s.NConn = big.NewInt(int64(len(s.Conns)))
	return nil
}

// Close closes the conns pool.
func (s *Server) Close() error {
	for _, conn := range s.Conns {
		if err := conn.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Healthcheck TODO returns if connections are still listening.
func (s *Server) Healthcheck() error {
	return nil
}
