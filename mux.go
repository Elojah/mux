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

// Service handles data and traffic parameters.
type Service struct {
	*logrus.Entry
	*Config

	Server
	Clients
	sync.Map
}

// NewMux returns a new clear Service.
func NewMux() *Service {
	return &Service{}
}

// Add adds a new handler identified by a string.
func (s *Service) Add(identifier string, f Handler) {
	s.Store(identifier, f)
}

// Get returns a previously added handler identified by a string.
func (s *Service) Get(identifier string) (Handler, error) {
	f, ok := s.Load(identifier)
	if !ok {
		return nil, fmt.Errorf("handler %s doesn't exist", identifier)
	}
	return f.(Handler), nil
}

// Read reads one packet from Conn and run it in identifier handler.
func (s *Service) Read() error {
	for {
		packet := make([]byte, s.PacketSize)
		_, err := s.Server.Read(packet)
		if err != nil {
			s.Logger.WithField("error", err).Error("failed to read")
			return err
		}
		go func(packet []byte) {
			s.Logger.WithFields(logrus.Fields{
				"type":   "packet",
				"status": "received",
				"data":   string(packet),
			}).Info("packet read successfully")

			fbs := make([]byte, s.PacketSize)
			if err := lz4.Uncompress(packet, fbs); err != nil {
				s.Logger.WithFields(logrus.Fields{
					"type":   "packet",
					"format": "lz4",
					"status": "received",
					"error":  err,
				}).Error("packet failed to uncompress")
				s.Logger.WithField("error", err).WithField("format", "lz4").Info("packet")
				return
			}
		}(packet)
	}
}

// Write writes one packet to conn opened previously in conn map.
func (s *Service) Write(packet []byte, identifier string) {
	if uint(len(packet)) > s.PacketSize {
		s.Logger.WithFields(logrus.Fields{
			"type":   "packet",
			"status": "failed",
			"error":  errors.New("packet too large"),
		}).Error("packet not sent")
		return
	}
	go func() {
		client, err := s.Clients.Get(identifier)
		if err != nil {
			s.Logger.WithFields(logrus.Fields{
				"type":       "connection",
				"status":     "unknown",
				"identifier": identifier,
				"error":      err,
			}).Error("packet not sent")
			return
		}
		n, err := client.Write(packet)
		if err != nil {
			s.Logger.WithFields(logrus.Fields{
				"type":       "packet",
				"status":     "failed",
				"identifier": identifier,
				"error":      err,
			}).Error("packet not sent")
			return
		}
		s.Logger.WithFields(logrus.Fields{
			"type":       "packet",
			"status":     "sent",
			"identifier": identifier,
			"size":       n,
		}).Info("packet sent")
	}()
}
