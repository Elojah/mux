package mux

import (
	"net"
	"sync"

	"github.com/elojah/services"
	"github.com/rs/zerolog/log"
)

// Client wraps an UDP connection.
type Client struct {
	Middlewares []Middleware
}

// Dial init the UDP client.
func (c *Client) Dial(cfg Config) error {
	for _, mw := range cfg.Middlewares {
		switch mw {
		case "lz4":
			lz4 := mwlz4{PacketSize: cfg.PacketSize}
			c.Middlewares = append(c.Middlewares, lz4)
		}
	}
	return nil
}

// Close closes the conns pool.
func (c *Client) Close() error {
	return nil
}

// Healthcheck TODO returns if connections are still listening.
func (c *Client) Healthcheck() error {
	return nil
}

// Send sends a packet via a random connection picked.
// You should run it in a go call.
func (c *Client) Send(raw []byte, addr net.Addr) {
	for _, mw := range c.Middlewares {
		var err error
		raw, err = mw.Send(raw)
		if err != nil {
			log.Error().Err(err).Str("status", "invalid").Msg("packet rejected")
			return
		}
	}

	conn, err := net.DialUDP("udp", nil, addr.(*net.UDPAddr))
	if err != nil {
		log.Error().Err(err).Str("address", addr.String()).Msg("failed to initialize connection")
		return
	}
	n, err := conn.Write(raw)
	if err != nil {
		log.Error().Err(err).Str("address", addr.String()).Msg("failed to write packet")
		return
	}
	log.Info().Int("bytes", n).Str("address", addr.String()).Msg("packet sent")
}

// ClientNamespaces maps configs used for client service with config file namespaces.
type ClientNamespaces struct {
	Client services.Namespace
}

// ClientLauncher represents a client launcher.
type ClientLauncher struct {
	*services.Configs
	ns ClientNamespaces

	client *Client
	m      sync.Mutex
}

// NewLauncher returns a new client ClientLauncher.
func (c *Client) NewLauncher(ns ClientNamespaces, nsRead ...services.Namespace) *ClientLauncher {
	return &ClientLauncher{
		Configs: services.NewConfigs(nsRead...),
		client:  c,
		ns:      ns,
	}
}

// Up starts the client service with new configs.
func (l *ClientLauncher) Up(configs services.Configs) error {
	l.m.Lock()
	defer l.m.Unlock()

	cfg := Config{}
	if err := cfg.Dial(configs[l.ns.Client]); err != nil {
		return err
	}
	return l.client.Dial(cfg)
}

// Down stops the client service.
func (l *ClientLauncher) Down(configs services.Configs) error {
	l.m.Lock()
	defer l.m.Unlock()
	return l.client.Close()
}
