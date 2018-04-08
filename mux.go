package udp

import (
	"errors"
	"fmt"
	"sync"

	"github.com/cloudflare/golz4"
	"github.com/sirupsen/logrus"
)

// Handler is handle function responsible to process incoming data.
type Handler func([]byte) error

// Mux handles data and traffic parameters.
type Mux struct {
	*logrus.Entry
	*Config

	Server
	Clients
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

// Read reads one packet from Conn and run it in identifier handler.
func (m *Mux) Read() error {
	for {
		packet := make([]byte, m.PacketSize)
		_, err := m.Server.Read(packet)
		if err != nil {
			m.Logger.WithField("error", err).Error("failed to read")
			return err
		}
		go func(packet []byte) {
			m.Logger.WithFields(logrus.Fields{
				"type":   "packet",
				"status": "received",
				"data":   string(packet),
			}).Info("packet read successfully")

			fbs := make([]byte, m.PacketSize)
			if err := lz4.Uncompress(packet, fbs); err != nil {
				m.Logger.WithFields(logrus.Fields{
					"type":   "packet",
					"format": "lz4",
					"status": "received",
					"error":  err,
				}).Error("packet failed to uncompress")
				m.Logger.WithField("error", err).WithField("format", "lz4").Info("packet")
				return
			}
		}(packet)
	}
}

// Write writes one packet to conn opened previously in conn map.
func (m *Mux) Write(packet []byte, identifier string) {
	if uint(len(packet)) > m.PacketSize {
		m.Logger.WithFields(logrus.Fields{
			"type":   "packet",
			"status": "failed",
			"error":  errors.New("packet too large"),
		}).Error("packet not sent")
		return
	}
	go func() {
		client, err := m.Clients.Get(identifier)
		if err != nil {
			m.Logger.WithFields(logrus.Fields{
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
				"type":       "packet",
				"status":     "failed",
				"identifier": identifier,
				"error":      err,
			}).Error("packet not sent")
			return
		}
		m.Logger.WithFields(logrus.Fields{
			"type":       "packet",
			"status":     "sent",
			"identifier": identifier,
			"size":       n,
		}).Info("packet sent")
	}()
}
