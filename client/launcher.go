package client

import (
	"sync"

	"github.com/elojah/services"
)

// Namespaces maps configs used for client service with config file namespaces.
type Namespaces struct {
	Client services.Namespace
}

// Launcher represents a client launcher.
type Launcher struct {
	*services.Configs
	ns Namespaces

	client *C
	m      sync.Mutex
}

// NewLauncher returns a new client Launcher.
func (c *C) NewLauncher(ns Namespaces, nsRead ...services.Namespace) *Launcher {
	return &Launcher{
		Configs: services.NewConfigs(nsRead...),
		client:  c,
		ns:      ns,
	}
}

// Up starts the client service with new configs.
func (l *Launcher) Up(configs services.Configs) error {
	l.m.Lock()
	defer l.m.Unlock()

	cfg := Config{}
	if err := cfg.Dial(configs[l.ns.Client]); err != nil {
		return err
	}
	return l.client.Dial(cfg)
}

// Down stops the client service.
func (l *Launcher) Down(configs services.Configs) error {
	l.m.Lock()
	defer l.m.Unlock()
	return l.client.Close()
}
