package udp

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/cloudflare/golz4"
	"github.com/oklog/ulid"
	"github.com/sirupsen/logrus"
)

type Middleware interface {
	Send([]byte) ([]byte, error)
	Receive([]byte) ([]byte, error)
}

type Packet struct {
	ID     ulid.ULID
	Source *net.UDPAddr
	Data   []byte
}

// Handler is handle function responsible to process incoming data.
type Handler func(Packet) error

// Dispatcher is a dispatch function used to dispatch incoming packets.
type Dispatcher func([]byte) (string, error)

// Mux handles data and traffic parameters.
type Mux struct {
	*logrus.Entry
	*Config

	Server
	Clients
	Dispatcher Dispatcher

	Middlewares []Middleware

	// Map stores the different handlers.
	sync.Map
}

// NewMux returns a new clear Mux.
func NewMux() *Mux {
	return &Mux{}
}

// Dial starts the mux server.
func (m *Mux) Dial(cfg Config) error {
	m.Config = &cfg
	return m.Server.Dial(cfg)
}

// Close resets the clients map and closes the server.
func (m *Mux) Close() error {
	m.Map = sync.Map{}
	return m.Server.Close()
}

// Add adds a new handler identified by a string.
func (m *Mux) Add(identifier string, f Handler) {
	m.Store(identifier, f)
}

// Get returns a previously added handler identified by a string.
func (m *Mux) Get(identifier string) (Handler, error) {
	f, ok := m.Load(identifier)
	if !ok {
		return nil, fmt.Errorf("handler %s doesn't exist", identifier)
	}
	return f.(Handler), nil
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
			identifier, err := m.Dispatcher(packet)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"status": "unidentified",
					"error":  err,
				}).Error("packet rejected")
				return
			}
			handler, err := m.Get(identifier)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"status": "unassigned",
					"error":  err,
				}).Error("packet rejected")
				return
			}
			logger.WithFields(logrus.Fields{
				"status":     "read",
				"identifier": identifier,
			}).Info("packet read")
			if err := handler(packet); err != nil {
				logger.WithFields(logrus.Fields{
					"status":     "processed",
					"identifier": identifier,
					"error":      err,
				}).Error("packet read but failed to be process")
				return
			}
			logger.WithFields(logrus.Fields{
				"status":     "processed",
				"identifier": identifier,
			}).Info("packet processed")
		}(Packet{
			ID:     ulid.MustNew(ulid.Timestamp(time.Now()), rand.Reader),
			Source: addr,
			Data:   raw[:n],
		})
	}
}

// Send writes one packet to conn opened previously in conn map.
func (m *Mux) Send(packet Packet, identifier string) error {
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
	go func(packet Packet) {
		logger := m.Logger.WithFields(logrus.Fields{
			"id":     packet.ID.String(),
			"source": packet.Source.String(),
			"type":   "packet",
		})
		client, err := m.Clients.Get(identifier)
		if err != nil {
			m.Logger.WithFields(logrus.Fields{
				"status":     "unassigned",
				"identifier": identifier,
				"error":      err,
			}).Error("packet not sent")
			return
		}
		n, err := client.Write(packet)
		if err != nil {
			m.Logger.WithFields(logrus.Fields{
				"id":         id,
				"type":       "packet",
				"status":     "failed",
				"identifier": identifier,
				"error":      err,
			}).Error("packet not sent")
			return
		}
		m.Logger.WithFields(logrus.Fields{
			"id":         id,
			"type":       "packet",
			"status":     "sent",
			"identifier": identifier,
			"size":       n,
		}).Info("packet sent")
	}(packet)
	return nil
}
