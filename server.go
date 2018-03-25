package udp

import (
	"net"
)

// Server wraps an UDP connection.
type Server struct {
	*net.UDPConn
}

// Dial init the UDP server.
func (s *Server) Dial(c Config) error {
	var err error
	address, err := net.ResolveUDPAddr("udp", c.Address)
	if err != nil {
		return err
	}
	s.UDPConn, err = net.ListenUDP("udp", address)
	return err
}

// Healthcheck returns if database responds.
func (s *Server) Healthcheck() error {
	return nil
}
