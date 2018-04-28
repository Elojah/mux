package mux

import (
	"context"
	"crypto/rand"
	"net"
	"time"

	"github.com/oklog/ulid"
	"github.com/rs/zerolog/log"
)

// Key is a context key.
type Key string

// Handler is handle function responsible to process incoming data.
type Handler func(context.Context, []byte) error

// M handles data and traffic parameters.
type M struct {
	*Config

	Server

	Middlewares []Middleware

	Handler Handler
}

// NewM returns a new clear M.
func NewM() *M {
	return &M{}
}

// Dial starts the mux server.
func (m *M) Dial(cfg Config) error {
	m.Config = &cfg
	for _, mw := range cfg.Middlewares {
		switch mw {
		case "lz4":
			lz4 := mwlz4{PacketSize: cfg.PacketSize}
			m.Middlewares = append(m.Middlewares, lz4)
		}
	}
	return m.Server.Dial(cfg)
}

// Close resets the clients map and closes the server.
func (m *M) Close() error {
	return m.Server.Close()
}

// Listen reads one packet from Conn and run it in identifier handler.
func (m *M) Listen() {
	for {
		raw := make([]byte, m.PacketSize)
		n, addr, err := m.ReadFrom(raw)
		if err != nil {
			log.Error().Err(err).Msg("failed to read")
		}

		go func(addr net.Addr, raw []byte) {
			ctx := context.WithValue(context.Background(), Key("addr"), addr.String())
			ctx = context.WithValue(ctx, Key("packet"), ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader).String())
			logger := log.With().
				Str("packet", ctx.Value(Key("packet")).(string)).
				Str("addr", ctx.Value(Key("addr")).(string)).
				Logger()
			logger.Info().Str("status", "read").Msg("packet read")
			if uint(n) > m.PacketSize {
				logger.Error().Err(ErrTooLargePacket).Str("status", "sizeable").Msg("packet rejected")
				return
			}

			for _, mw := range m.Middlewares {
				raw, err = mw.Receive(raw)
				if err != nil {
					logger.Error().Err(err).Str("status", "invalid").Msg("packet rejected")
					return
				}
			}
			if err := m.Handler(ctx, raw); err != nil {
				// Logging must be done inside handler.
				return
			}
			logger.Info().Str("status", "processed").Msg("packet processed")
		}(addr, raw[:n])
	}
}
