package mux

import (
	"context"
	"crypto/rand"
	"net"
	"time"

	"github.com/oklog/ulid"
	"github.com/rs/zerolog/log"
)

// Key represents context keys.
type Key string

const (
	// Address is the context key for remote address of connection.
	Address Key = "address"
	// Packet is the context key for packet id assigned when received.
	Packet Key = "packet"
)

// Handler is handle function responsible to process incoming data.
type Handler func(context.Context, []byte) error

// Mux handles data and traffic parameters.
type Mux struct {
	*Config

	Server

	Middlewares []Middleware

	Handler Handler
}

// NewMux returns a new clear Mux.
func NewMux() *Mux {
	return &Mux{}
}

// Dial starts the mux server.
func (m *Mux) Dial(cfg Config) error {
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
func (m *Mux) Close() error {
	return m.Server.Close()
}

// Listen reads one packet from Conn and run it in identifier handler.
func (m *Mux) Listen() {
	for {
		conn, err := m.Server.Accept()
		ctx := context.Background()
		ctx = context.WithValue(ctx, Address, conn.RemoteAddr().String())
		if err != nil {
			log.Ctx(ctx).Error().Msg("connection refused")
		}

		go func(ctx context.Context, conn net.Conn) {
			defer func() { _ = conn.Close() }()
			log.Ctx(ctx).Info().Msg("connection accepted")
			raw := make([]byte, m.PacketSize)

			for {
				n, err := conn.Read(raw)
				if err != nil {
					log.Ctx(ctx).Error().Err(err).Msg("failed to read")
					continue
				}

				go func(ctx context.Context, raw []byte) {
					ctx = context.WithValue(ctx, Packet, ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader))
					if uint(n) > m.PacketSize {
						log.Ctx(ctx).Error().Err(ErrTooLargePacket).Str("status", "sizeable").Msg("packet rejected")
						return
					}

					for _, mw := range m.Middlewares {
						raw, err = mw.Receive(raw)
						if err != nil {
							log.Ctx(ctx).Error().Err(err).Str("status", "invalid").Msg("packet rejected")
							return
						}
					}
					if err := m.Handler(ctx, raw); err != nil {
						// Logging must be done inside handler.
						return
					}
					log.Ctx(ctx).Info().Str("status", "processed").Msg("packet processed")
				}(ctx, raw[:n])
			}
		}(ctx, conn)
	}
}
