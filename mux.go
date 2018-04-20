package udp

import (
	"crypto/rand"
	"errors"
	"net"
	"sync"
	"time"

	"github.com/oklog/ulid"
	"github.com/sirupsen/logrus"
)

// Packet represents a network packet sent or received.
type Packet struct {
	ID     ulid.ULID
	Source *net.UDPAddr
	Data   []byte
}

// Handler is handle function responsible to process incoming data.
type Handler func(Packet) error

// Mux handles data and traffic parameters.
type Mux struct {
	*logrus.Entry
	*Config

	Server
	Clients

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
	m.Map = sync.Map{}
	return m.Server.Close()
}

// Listen reads one packet from Conn and run it in identifier handler.
func (m *Mux) Listen() {
	for {
		raw := make([]byte, m.PacketSize)
		n, addr, err := m.Server.ReadFromUDP(raw)
		if err != nil {
			m.Logger.WithField("error", err).Error("failed to read")
			break
		}
		if uint(n) > m.PacketSize {
			err := errors.New("packet too large")
			m.Logger.WithFields(logrus.Fields{
				"address": addr.String(),
				"type":    "packet",
				"status":  "intractable",
				"error":   err,
			}).Error("packet rejected")
			break
		}
		go func(packet Packet) {
			logger := m.Logger.WithFields(logrus.Fields{
				"id":      packet.ID.String(),
				"address": packet.Source.String(),
				"type":    "packet",
			})
			for _, mw := range m.Middlewares {
				packet.Data, err = mw.Receive(packet.Data)
				if err != nil {
					logger.WithFields(logrus.Fields{
						"status": "unreadable",
						"error":  err,
					}).Error("packet rejected")
					return
				}
			}
			if err := m.Handler(packet); err != nil {
				logger.WithFields(logrus.Fields{
					"status": "unprocessed",
					"error":  err,
				}).Error("packet rejected")
				return
			}
			logger.WithFields(logrus.Fields{
				"status": "processed",
			}).Info("packet processed")
		}(Packet{
			ID:     ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader),
			Source: addr,
			Data:   raw[:n],
		})
	}
}

// Send writes one packet to conn opened previously in conn map.
func (m *Mux) Send(packet Packet, address string) error {
	var err error
	for _, mw := range m.Middlewares {
		packet.Data, err = mw.Send(packet.Data)
		if err != nil {
			m.Logger.WithFields(logrus.Fields{
				"type":   "packet",
				"status": "failed",
				"error":  err,
			}).Error("packet not sent")
		}
	}
	go func(packet Packet, address string) {
		logger := m.Logger.WithFields(logrus.Fields{
			"id":      packet.ID.String(),
			"source":  packet.Source.String(),
			"type":    "packet",
			"address": address,
		})
		client, err := m.Clients.Get(address)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"status": "unassigned",
				"error":  err,
			}).Error("packet not sent")
			return
		}
		n, err := client.Write(packet.Data)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"status": "failed",
				"error":  err,
			}).Error("packet not sent")
			return
		}
		logger.WithFields(logrus.Fields{
			"status": "sent",
			"size":   n,
		}).Info("packet sent")
	}(packet, address)
	return nil
}
