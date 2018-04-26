package mux

import (
	"sync"

	"github.com/elojah/services"
)

// Namespaces maps configs used for mux service with config file namespaces.
type Namespaces struct {
	M services.Namespace
}

// Launcher represents a mux launcher.
type Launcher struct {
	*services.Configs
	ns Namespaces

	mux *M
	m   sync.Mutex
}

// NewLauncher returns a new mux Launcher.
func (mux *M) NewLauncher(ns Namespaces, nsRead ...services.Namespace) *Launcher {
	return &Launcher{
		Configs: services.NewConfigs(nsRead...),
		mux:     mux,
		ns:      ns,
	}
}

// Up starts the mux service with new configs.
func (l *Launcher) Up(configs services.Configs) error {
	l.m.Lock()
	defer l.m.Unlock()

	cfg := Config{}
	if err := cfg.Dial(configs[l.ns.M]); err != nil {
		// Add namespace key when returning error with logrus
		return err
	}
	return l.mux.Dial(cfg)
}

// Down stops the mux service.
func (l *Launcher) Down(configs services.Configs) error {
	l.m.Lock()
	defer l.m.Unlock()

	return l.mux.Close()
}
