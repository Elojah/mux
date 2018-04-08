package udp

import (
	"net"
	"sync"
)

// Clients is a pool of UDP connections without TTL.
type Clients struct {
	sync.Map
}

// Get acts as a LoadOrStore for UDP connection to an address.
func (cs *Clients) Get(address string) (*net.UDPConn, error) {
	conn, ok := cs.Load(address)
	if !ok {
		conn, err := net.Dial("udp", address)
		if err != nil {
			return nil, err
		}
		cs.Store(address, conn)
		return conn.(*net.UDPConn), nil
	}
	return conn.(*net.UDPConn), nil
}
