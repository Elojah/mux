package mux

import (
	"net"
)

// Server wraps an UDP connection.
type Server struct {
	net.PacketConn
}

// Dial init the UDP server.
func (s *Server) Dial(c Config) error {
	var err error
	s.PacketConn, err = net.ListenPacket(c.ServerProtocol, c.Address)
	return err
}

// Healthcheck returns if database responds.
func (s *Server) Healthcheck() error {
	return nil
}
