package bitfield

import "testing"

func TestBitfieldHasAndSetPiece(t *testing.T) {
	bf := Bitfield{0b10000001, 0b01000000}

	for _, index := range []int{0, 7, 9} {
		if !bf.HasPiece(index) {
			t.Errorf("HasPiece(%d) = false, want true", index)
		}
	}
	if bf.HasPiece(1) {
		t.Error("HasPiece(1) = true, want false")
	}

	bf.SetPiece(14)
	if !bf.HasPiece(14) {
		t.Error("SetPiece(14) did not set piece")
	}
}

func TestBitfieldIgnoresOutOfRangeIndexes(t *testing.T) {
	bf := Bitfield{0}
	bf.SetPiece(-1)
	bf.SetPiece(8)

	if bf[0] != 0 {
		t.Errorf("bitfield = %08b, want unchanged", bf[0])
	}
	if bf.HasPiece(-1) || bf.HasPiece(8) {
		t.Error("out-of-range piece reported as present")
	}
}
