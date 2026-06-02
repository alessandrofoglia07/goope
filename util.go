package goope

func bytesToBits(b []byte) []bool {
	bits := make([]bool, 0, len(b)*8)
	for _, b := range b {
		for i := 7; i >= 0; i-- {
			bits = append(bits, (b>>i)&1 == 1)
		}
	}
	return bits
}
