package goope

// bytesToBits converts a byte slice into a slice of bits (booleans), where each byte expands into 8 bits. The bits are ordered from the most significant bit to the least significant bit for each byte.
func bytesToBits(b []byte) []bool {
	bits := make([]bool, 0, len(b)*8)
	for _, b := range b {
		for i := 7; i >= 0; i-- {
			bits = append(bits, (b>>i)&1 == 1)
		}
	}
	return bits
}

func boolToUint32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}
