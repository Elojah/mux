package client

import (
	"net"

	"github.com/elojah/mux"
	"github.com/rs/zerolog/log"
)

// C wraps an UDP connection.
type C struct {
	Middlewares []mux.Middleware
}

// Dial init the UDP client.
func (c *C) Dial(cfg Config) error {
	for _, mw := range cfg.Middlewares {
		switch mw {
		case "lz4":
			lz4 := mux.Mwlz4{PacketSize: cfg.PacketSize}
			c.Middlewares = append(c.Middlewares, lz4)
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
	for _, mw := range c.Middlewares {
		var err error
		raw, err = mw.Send(raw)
		if err != nil {
			log.Error().Err(err).Str("status", "invalid").Msg("packet rejected")
			return
		}
	}

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
