package udp

import (
	"errors"
)

// Config is a UDP server config.
type Config struct {
	Address string
}

// Equal returns is both configs are equal.
func (c Config) Equal(rhs Config) bool {
	return c.Address == rhs.Address
}

// Dial set the config from a config namespace.
func (c *Config) Dial(fileconf interface{}) error {
	fconf, ok := fileconf.(map[string]interface{})
	if !ok {
		return errors.New("namespace empty")
	}
	cAddress, ok := fconf["address"]
	if !ok {
		return errors.New("missing key address")
	}
	if c.Address, ok = cAddress.(string); !ok {
		return errors.New("key address invalid. must be string")
	}
	return nil
}
