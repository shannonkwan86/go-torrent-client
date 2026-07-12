package peers

import "testing"

func TestUnmarshal(t *testing.T) {
	input := []byte{
		127, 0, 0, 1, 0x1A, 0xE1,
		192, 168, 1, 9, 0x00, 0x50,
	}

	got, err := Unmarshal(input)
	if err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got, want := len(got), 2; got != want {
		t.Fatalf("peer count = %d, want %d", got, want)
	}
	if got[0].String() != "127.0.0.1:6881" {
		t.Errorf("first peer = %q, want %q", got[0].String(), "127.0.0.1:6881")
	}
	if got[1].String() != "192.168.1.9:80" {
		t.Errorf("second peer = %q, want %q", got[1].String(), "192.168.1.9:80")
	}
}

func TestUnmarshalRejectsMalformedLength(t *testing.T) {
	_, err := Unmarshal([]byte{127, 0, 0, 1, 0})
	if err == nil {
		t.Fatal("Unmarshal() error = nil, want error")
	}
}
