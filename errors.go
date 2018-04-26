package mux

import (
	"errors"
)

var (
	// ErrTooLargePacket is raised when a packet larger than size defined in config is received.
	ErrTooLargePacket = errors.New("packet too large")
)
