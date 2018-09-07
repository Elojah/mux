package mux

import (
	"errors"
)

// Config is a UDP server config.
type Config struct {
	Addresses  []string `json:"addresses"`
	PacketSize uint     `json:"packet_size"`
}

// Equal returns is both configs are equal.
func (c Config) Equal(rhs Config) bool {

	if len(c.Addresses) != len(rhs.Addresses) {
		return false
	}
	for i := range c.Addresses {
		if c.Addresses[i] != rhs.Addresses[i] {
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
	cAddresses, ok := fconf["addresses"]
	if !ok {
		return errors.New("missing key addresses")
	}
	cAddressesSlice, ok := cAddresses.([]interface{})
	if !ok {
		return errors.New("key addresses invalid. must be slice")
	}
	c.Addresses = make([]string, len(cAddressesSlice))
	for i, adress := range cAddressesSlice {
		c.Addresses[i], ok = adress.(string)
		if !ok {
			return errors.New("value in addresses invalid. must be string")
		}
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
