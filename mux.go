package udp

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/cloudflare/golz4"
	"github.com/oklog/ulid"
	"github.com/sirupsen/logrus"
)

type randReader struct{}

func (randReader) Read(p []byte) (n int, err error) {
	return rand.Read(p)
}

// Handler is handle function responsible to process incoming data.
type Handler func([]byte) error

// Dispatcher is a dispatch function used to dispatch incoming packets.
type Dispatcher func([]byte) (string, error)

// Mux handles data and traffic parameters.
type Mux struct {
	*logrus.Entry
	*Config

	Server
	Clients
	Dispatcher Dispatcher

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
		id := ulid.MustNew(ulid.Timestamp(time.Now()), randReader{}).String()
		if uint(n) > m.PacketSize {
			err := errors.New("packet too large")
			m.Logger.WithFields(logrus.Fields{
				"id":      id,
				"address": addr.String(),
				"type":    "packet",
				"status":  "unparsable",
				"error":   err,
			}).Error("packet rejected")
			break
		}
		go func(raw []byte) {
			packet := make([]byte, m.PacketSize)
			if err := lz4.Uncompress(raw, packet); err != nil {
				m.Logger.WithFields(logrus.Fields{
					"id":      id,
					"address": addr.String(),
					"type":    "packet",
					"status":  "unreadable",
					"error":   err,
				}).Error("packet rejected")
				return
			}
			identifier, err := m.Dispatcher(packet)
			if err != nil {
				m.Logger.WithFields(logrus.Fields{
					"id":      id,
					"address": addr.String(),
					"type":    "packet",
					"status":  "unidentified",
					"error":   err,
				}).Error("packet rejected")
				return
			}
			handler, err := m.Get(identifier)
			if err != nil {
				m.Logger.WithFields(logrus.Fields{
					"id":      id,
					"address": addr.String(),
					"type":    "handler",
					"status":  "unknown",
					"error":   err,
				}).Error("packet rejected")
				return
			}
			m.Logger.WithFields(logrus.Fields{
				"id":         id,
				"address":    addr.String(),
				"type":       "packet",
				"status":     "read",
				"identifier": identifier,
				"error":      err,
			}).Error("packet read")
			if err := handler(packet); err != nil {
				m.Logger.WithFields(logrus.Fields{
					"id":         id,
					"address":    addr.String(),
					"type":       "packet",
					"status":     "processed",
					"identifier": identifier,
					"error":      err,
				}).Error("packet read but failed to be process")
				return
			}
			m.Logger.WithFields(logrus.Fields{
				"id":         id,
				"address":    addr.String(),
				"type":       "packet",
				"status":     "processed",
				"identifier": identifier,
			}).Error("packet processed")
		}(raw[:n])
	}
}

// Send writes one packet to conn opened previously in conn map.
func (m *Mux) Send(id string, raw []byte, identifier string) error {
	packet := make([]byte, lz4.CompressBound(raw))
	n, err := lz4.Compress(raw, packet)
	if err != nil {
		m.Logger.WithFields(logrus.Fields{
			"id":     id,
			"type":   "packet",
			"format": "lz4",
			"status": "failed",
			"error":  err,
		}).Error("packet not sent")
		return err
	}
	if uint(n) > m.PacketSize {
		err := errors.New("packet too large")
		m.Logger.WithFields(logrus.Fields{
			"id":     id,
			"type":   "packet",
			"status": "failed",
			"error":  err,
		}).Error("packet not sent")
		return err
	}
	go func(packet []byte) {
		client, err := m.Clients.Get(identifier)
		if err != nil {
			m.Logger.WithFields(logrus.Fields{
				"id":         id,
				"type":       "connection",
				"status":     "unknown",
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
	}(packet[:n])
	return nil
}
