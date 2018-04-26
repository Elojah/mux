package mux

import (
	"net"
	"testing"

	"github.com/elojah/services"
)

func TestDial(t *testing.T) {
	message := []byte("To you rudy")
	t.Run("up", func(t *testing.T) {
		mux := Mux{}
		l := mux.NewLauncher(Namespaces{Mux: "server"}, "server")
		if err := l.Up(services.Configs{
			"server": map[string]interface{}{
				"address":         "localhost:4242",
				"middlewares":     []interface{}{"lz4"},
				"packet_size":     float64(1024),
				"server_protocol": "tcp",
			},
		}); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = l.Down(nil) }()

		conn, err := net.Dial("tcp", "localhost:4242")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = conn.Write(message); err != nil {
			t.Fatal(err)
		}
	})
}
