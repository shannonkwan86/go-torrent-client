package torrentfile

import (
	"bytes"
	"fmt"
	"strconv"
)

const maxBencodeDepth = 1000

// findInfoBytes returns the exact bencoded value of the top-level info key.
// The info hash must be calculated from these original bytes, not from a
// decoded and re-encoded representation, which may omit unknown fields.
func findInfoBytes(data []byte) ([]byte, error) {
	if len(data) == 0 || data[0] != 'd' {
		return nil, fmt.Errorf("torrent root must be a dictionary")
	}

	position := 1
	infoStart := -1
	infoEnd := -1
	for {
		if position >= len(data) {
			return nil, fmt.Errorf("unterminated torrent dictionary")
		}
		if data[position] == 'e' {
			position++
			break
		}

		key, next, err := scanBencodeString(data, position)
		if err != nil {
			return nil, fmt.Errorf("read dictionary key at byte %d: %w", position, err)
		}
		valueStart := next
		valueEnd, err := scanBencodeValue(data, valueStart, 0)
		if err != nil {
			return nil, fmt.Errorf("read value for key %q: %w", key, err)
		}

		if bytes.Equal(key, []byte("info")) {
			if infoStart >= 0 {
				return nil, fmt.Errorf("duplicate info key")
			}
			if data[valueStart] != 'd' {
				return nil, fmt.Errorf("info value must be a dictionary")
			}
			infoStart = valueStart
			infoEnd = valueEnd
		}
		position = valueEnd
	}

	if position != len(data) {
		return nil, fmt.Errorf("trailing data after torrent dictionary")
	}
	if infoStart < 0 {
		return nil, fmt.Errorf("torrent dictionary has no info key")
	}
	return data[infoStart:infoEnd], nil
}

func scanBencodeValue(data []byte, position, depth int) (int, error) {
	if depth > maxBencodeDepth {
		return 0, fmt.Errorf("nesting exceeds maximum depth %d", maxBencodeDepth)
	}
	if position >= len(data) {
		return 0, fmt.Errorf("missing value at byte %d", position)
	}

	switch data[position] {
	case 'i':
		return scanBencodeInteger(data, position)
	case 'l':
		position++
		for {
			if position >= len(data) {
				return 0, fmt.Errorf("unterminated list")
			}
			if data[position] == 'e' {
				return position + 1, nil
			}
			var err error
			position, err = scanBencodeValue(data, position, depth+1)
			if err != nil {
				return 0, err
			}
		}
	case 'd':
		position++
		for {
			if position >= len(data) {
				return 0, fmt.Errorf("unterminated dictionary")
			}
			if data[position] == 'e' {
				return position + 1, nil
			}
			_, next, err := scanBencodeString(data, position)
			if err != nil {
				return 0, fmt.Errorf("read dictionary key at byte %d: %w", position, err)
			}
			position, err = scanBencodeValue(data, next, depth+1)
			if err != nil {
				return 0, err
			}
		}
	default:
		_, next, err := scanBencodeString(data, position)
		return next, err
	}
}

func scanBencodeString(data []byte, position int) ([]byte, int, error) {
	if position >= len(data) || data[position] < '0' || data[position] > '9' {
		return nil, 0, fmt.Errorf("expected byte string at byte %d", position)
	}

	colonOffset := bytes.IndexByte(data[position:], ':')
	if colonOffset < 0 {
		return nil, 0, fmt.Errorf("byte string at byte %d has no length separator", position)
	}
	colon := position + colonOffset
	lengthText := data[position:colon]
	if len(lengthText) > 1 && lengthText[0] == '0' {
		return nil, 0, fmt.Errorf("byte string length has a leading zero")
	}
	for _, digit := range lengthText {
		if digit < '0' || digit > '9' {
			return nil, 0, fmt.Errorf("invalid byte string length %q", lengthText)
		}
	}

	length, err := strconv.ParseUint(string(lengthText), 10, 64)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid byte string length %q: %w", lengthText, err)
	}
	start := colon + 1
	if length > uint64(len(data)-start) {
		return nil, 0, fmt.Errorf("byte string length %d exceeds remaining input", length)
	}
	end := start + int(length)
	return data[start:end], end, nil
}

func scanBencodeInteger(data []byte, position int) (int, error) {
	endOffset := bytes.IndexByte(data[position+1:], 'e')
	if endOffset < 0 {
		return 0, fmt.Errorf("unterminated integer at byte %d", position)
	}
	end := position + 1 + endOffset
	integer := data[position+1 : end]
	if len(integer) == 0 {
		return 0, fmt.Errorf("empty integer at byte %d", position)
	}

	digits := integer
	if digits[0] == '-' {
		digits = digits[1:]
		if len(digits) == 0 || (len(digits) == 1 && digits[0] == '0') {
			return 0, fmt.Errorf("invalid negative integer %q", integer)
		}
	}
	if len(digits) > 1 && digits[0] == '0' {
		return 0, fmt.Errorf("integer has a leading zero")
	}
	for _, digit := range digits {
		if digit < '0' || digit > '9' {
			return 0, fmt.Errorf("invalid integer %q", integer)
		}
	}
	return end + 1, nil
}
