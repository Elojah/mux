package client

import (
	"net"
	"testing"

	"github.com/elojah/services"
)

func TestSend(t *testing.T) {
	message := []byte("To you rudy")
	t.Run("up send", func(t *testing.T) {

		c := C{}
		lc := c.NewLauncher(Namespaces{Client: "client"}, "client")
		if err := lc.Up(services.Configs{
			"client": map[string]interface{}{
				"middlewares": []interface{}{},
				"packet_size": float64(1024),
			},
		}); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = lc.Down(nil) }()

		conn, err := net.ListenPacket("udp", "127.0.0.1:4244")
		if err != nil {
			t.Fatal(err)
		}
		listener, _ := net.ResolveUDPAddr("udp", "127.0.0.1:4244")
		c.Send(message, listener)
		raw := make([]byte, 100)
		n, _, err := conn.ReadFrom(raw)
		if err != nil {
			t.Fatal(err)
		}
		if n != 11 {
			t.Fatalf("expected %d bytes, received %d", 11, n)
		}
	})

}
