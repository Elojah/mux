package client

import (
	"net"

	"github.com/rs/zerolog/log"

	"github.com/elojah/mux"
)

// C wraps an UDP connection.
type C struct {
	Middlewares []mux.Middleware
}

// Dial init the UDP client.
func (c *C) Dial(cfg Config) error {
	for _, mw := range cfg.Middlewares {
		switch mw {
		case "snappy":
			c.Middlewares = append(c.Middlewares, mux.MwSnappy{})
		}
	}
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
	for _, mw := range c.Middlewares {
		var err error
		raw, err = mw.Send(raw)
		if err != nil {
			log.Error().Err(err).Str("status", "invalid").Msg("packet rejected")
			return
		}
	}
	n, err := conn.Write(raw)
	if err != nil {
		log.Error().Err(err).Str("address", addr.String()).Msg("failed to write packet")
		return
	}
	log.Info().Int("bytes", n).Str("address", addr.String()).Msg("packet sent")
}
