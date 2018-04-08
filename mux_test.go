package udp

import (
	"testing"

	"github.com/elojah/services"
)

func TestDial(t *testing.T) {
	message := []byte("To you rudy")
	t.Run("up", func(t *testing.T) {
		mux := Mux{}
		l := mux.NewLauncher(Namespaces{UDP: "server"}, "server")
		if err := l.Up(services.Configs{
			"server": map[string]interface{}{
				"address":     "localhost:8080",
				"packet_size": uint(1024),
			},
		}); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = l.Down(nil) }()

		cs := Clients{}
		conn, err := cs.Get("localhost:8080")
		if err != nil {
			t.Fatal(err)
		}
		if _, err = conn.Write(message); err != nil {
			t.Fatal(err)
		}
		rec := make([]byte, len(message))
		_, _, err = mux.ReadFromUDP(rec)
		if err != nil {
			t.Fatal(err)
		}
		if string(rec) != string(message) {
			t.Fatal("corrupted message")
		}
	})
}
