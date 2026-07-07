package bitfield

// Bitfield 表示某个 peer 拥有哪些 pieces。
// 每个 bit 对应一个 piece：1 表示拥有，0 表示没有。
type Bitfield []byte

// HasPiece 判断 bitfield 中指定 piece 下标是否被设置。
func (bf Bitfield) HasPiece(index int) bool {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= len(bf) {
		return false
	}

	return bf[byteIndex]>>uint(7-offset)&1 != 0
}

// SetPiece 将 bitfield 中指定 piece 下标标记为已拥有。
func (bf Bitfield) SetPiece(index int) {
	byteIndex := index / 8
	offset := index % 8
	if byteIndex < 0 || byteIndex >= len(bf) {
		return
	}

	bf[byteIndex] |= 1 << uint(7-offset)
}
