package mux

import "testing"

func TestSnappy(t *testing.T) {
	t.Run("send_receive", func(t *testing.T) {
		mwsnappy := MwSnappy{}
		raw := []byte("ice bow is well too OP plz motiontwins")
		sent, err := mwsnappy.Send(raw)
		if err != nil {
			t.Error(err)
			return
		}
		received, err := mwsnappy.Receive(sent)
		if err != nil {
			t.Error(err)
			return
		}
		if len(raw) > len(received) {
			t.Errorf("length invalid. expected min %d, actual %d", len(raw), len(received))
			return
		}
		for i := range raw {
			if raw[i] != received[i] {
				t.Errorf("byte invalid. expected %c, actual %c", raw[i], received[i])
				return
			}
		}
	})
}
