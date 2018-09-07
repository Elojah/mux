package client

import (
	"errors"
)

// Config is a UDP server config.
type Config struct {
	PacketSize uint `json:"packet_size"`
}

// Equal returns is both configs are equal.
func (c Config) Equal(rhs Config) bool {

	return c.PacketSize == rhs.PacketSize
}

// Dial set the config from a config namespace.
func (c *Config) Dial(fileconf interface{}) error {
	fconf, ok := fileconf.(map[string]interface{})
	if !ok {
		return errors.New("namespace empty")
	}
	cPacketSize, ok := fconf["packet_size"]
	if !ok {
		return errors.New("missing key packet_size")
	}
	cPacketSizeFloat, ok := cPacketSize.(float64)
	if !ok {
		return errors.New("key packet_size invalid. must be int")
	}
	c.PacketSize = uint(cPacketSizeFloat)
	return nil
}
