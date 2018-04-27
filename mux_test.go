package mux

import (
	"net"
	"testing"

	"github.com/elojah/services"
)

func TestDial(t *testing.T) {
	message := []byte("To you rudy")
	t.Run("up", func(t *testing.T) {
		mux := M{}
		l := mux.NewLauncher(Namespaces{M: "server"}, "server")
		if err := l.Up(services.Configs{
			"server": map[string]interface{}{
				"address":         "localhost:4242",
				"middlewares":     []interface{}{"lz4"},
				"packet_size":     float64(1024),
				"server_protocol": "udp",
			},
		}); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = l.Down(nil) }()

		conn, err := net.Dial("udp", "localhost:4242")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = conn.Write(message); err != nil {
			t.Fatal(err)
		}
	})
}
