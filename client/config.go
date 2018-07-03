package client

import (
	"errors"
)

// Config is a UDP server config.
type Config struct {
	Middlewares []string `json:"middlewares"`
	PacketSize  uint     `json:"packet_size"`
}

// Equal returns is both configs are equal.
func (c Config) Equal(rhs Config) bool {

	if len(c.Middlewares) != len(rhs.Middlewares) {
		return false
	}
	for i := range c.Middlewares {
		if c.Middlewares[i] != rhs.Middlewares[i] {
			return false
		}
	}

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
	cMiddlewares, ok := fconf["middlewares"]
	if !ok {
		return errors.New("missing key middlewares")
	}
	cMiddlewaresSlice, ok := cMiddlewares.([]interface{})
	if !ok {
		return errors.New("key middlewares invalid. must be slice")
	}
	c.Middlewares = make([]string, len(cMiddlewaresSlice))
	for i, middleware := range cMiddlewaresSlice {
		c.Middlewares[i], ok = middleware.(string)
		if !ok {
			return errors.New("value in middlewares invalid. must be string")
		}
	}
	return nil
}
