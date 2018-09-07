package client

import (
	"net"

	"github.com/rs/zerolog/log"
)

// C wraps an UDP connection.
type C struct {
}

// Dial init the UDP client.
func (c *C) Dial(cfg Config) error {
	return nil
}

// Close closes the conns pool.
func (c *C) Close() error {
	return nil
}

// Healthcheck TODO returns if connections are still listening.
func (c *C) Healthcheck() error {
	return nil
}

// Send sends a packet via a random connection picked.
// You should run it in a go call.
func (c *C) Send(raw []byte, addr net.Addr) {
	conn, err := net.DialUDP("udp", nil, addr.(*net.UDPAddr))
	if err != nil {
		log.Error().Err(err).Str("address", addr.String()).Msg("failed to initialize connection")
		return
	}
	n, err := conn.Write(raw)
	if err != nil {
		log.Error().Err(err).Str("address", addr.String()).Msg("failed to write packet")
		return
	}
	log.Info().Int("bytes", n).Str("address", addr.String()).Msg("packet sent")
}
