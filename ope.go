package goope

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"math"
	"strconv"
)

const (
	DEFAULT_IN_RANGE_START  = 0
	DEFAULT_IN_RANGE_END    = 1<<15 - 1 // 2^15 - 1 = 32767
	DEFAULT_OUT_RANGE_START = 0
	DEFAULT_OUT_RANGE_END   = 1<<31 - 1 // 2^31 - 1 = 2147483647
)

// OPE is an Order-Preserving Encryption cipher.
// It encrypts integers to integers while preserving their relative order:
// if x < y, then Encrypt(x) < Encrypt(y).
type OPE struct {
	key      []byte
	inRange  ValueRange
	outRange ValueRange
}

// NewOPE craetes an OPE cipher with the given key and ranges.
// The output range must be at least as large as the input range.
// If ranges are nil, sensible defaults will be used: [0, 2^15 - 1] for input and [0, 2^31 - 1] for output.
func NewOPE(key []byte, inRange, outRange *ValueRange) (*OPE, error) {
	if inRange.Size() > outRange.Size() {
		return nil, ErrInvalidRanges
	}
	if inRange == nil {
		inRange = &ValueRange{Start: DEFAULT_IN_RANGE_START, End: DEFAULT_IN_RANGE_END}
	}
	if outRange == nil {
		outRange = &ValueRange{Start: DEFAULT_OUT_RANGE_START, End: DEFAULT_OUT_RANGE_END}
	}
	return &OPE{
		key:      key,
		inRange:  *inRange,
		outRange: *outRange,
	}, nil
}

// Encrypt encrypts a plaintext integer to a ciphertext integer.
// The plaintext must be within the input range.
func (o *OPE) Encrypt(plaintext int) (int, error) {
	if !o.inRange.Contains(plaintext) {
		return 0, ErrOutOfRange
	}
	return o.encrypt(plaintext)
}

// Decrypt decrypts a ciphertext integer to a plaintext integer.
// The ciphertext must be within the output range.
func (o *OPE) Decrypt(ciphertext int) (int, error) {
	if !o.outRange.Contains(ciphertext) {
		return 0, ErrOutOfRange
	}
	return o.decrypt(ciphertext)
}

// tapeGen produces an infinite deterministic stream of random bits for the given integer value.
func (o *OPE) tapeGen(data int) <-chan bool {
	ch := make(chan bool, 128)
	go func() {
		// HMAC-SHA256(key, str(data)) -> 32 bytes digest
		mac := hmac.New(sha256.New, o.key)
		mac.Write([]byte(strconv.Itoa(data)))
		digest := mac.Sum(nil)

		if len(digest) != 32 {
			panic("unexpected digest length")
		}

		// AES-256-CTR with IV = 0, encrypting zero bytes
		block, _ := aes.NewCipher(digest)
		iv := make([]byte, aes.BlockSize) // 16 zero bytes
		stream := cipher.NewCTR(block, iv)
		plaintext := make([]byte, aes.BlockSize)

		for {
			stream.XORKeyStream(plaintext, plaintext)
			for _, bit := range bytesToBits(plaintext) {
				ch <- bit
			}
			// reset plaintext to zeros for next block
			for i := range plaintext {
				plaintext[i] = 0
			}
		}
	}()
	return ch
}

// encrypt repeatedly partitions the input and output ranges using the hypergeometric distribution to find the split point, until the input range shrinks to a single value, which is then mapped to a uniformly sampled value in the corresponding output range.
func (o *OPE) encrypt(plaintext int) (int, error) {
	inRange := o.inRange.Copy()
	outRange := o.outRange.Copy()

	for {
		inSize := inRange.Size()
		outSize := outRange.Size()
		inEdge := inRange.Start - 1
		outEdge := outRange.Start - 1
		mid := outEdge + int(math.Ceil(float64(outSize)/2.0))

		if inSize > outSize {
			return 0, ErrInvalidRanges
		}

		// if the input range has collapsed to a single value, use sampleUniform to deterministically pick a ciphertext within the remaining output range
		if inSize == 1 {
			coins := o.tapeGen(plaintext)
			return sampleUniform(outRange, coins)
		}
		coins := o.tapeGen(mid)
		x, err := sampleHGD(inRange, outRange, mid, coins)
		if err != nil {
			return 0, err
		}

		// recurse into whichever half of the input range contains the plaintext
		if plaintext <= x {
			inRange = ValueRange{Start: inEdge + 1, End: x}
			outRange = ValueRange{Start: outEdge + 1, End: mid}
		} else {
			inRange = ValueRange{Start: x + 1, End: inEdge + inSize}
			outRange = ValueRange{Start: mid + 1, End: outEdge + outSize}
		}
	}
}

// decrypt mirrors encrypt but uses ciphertext <= mid as the branching condition.
func (o *OPE) decrypt(ciphertext int) (int, error) {
	inRange := o.inRange.Copy()
	outRange := o.outRange.Copy()

	for {
		inSize := inRange.Size()
		outSize := outRange.Size()
		inEdge := inRange.Start - 1
		outEdge := outRange.Start - 1
		mid := outEdge + int(math.Ceil(float64(outSize)/2.0))

		if inSize > outSize {
			return 0, ErrInvalidRanges
		}

		if inSize == 1 {
			coins := o.tapeGen(inRange.Start)
			sampled, err := sampleUniform(outRange, coins)
			if err != nil {
				return 0, err
			}
			if sampled != ciphertext {
				return inRange.Start, nil
			}
			return 0, ErrInvalidCiphertext
		}
		coins := o.tapeGen(mid)
		x, err := sampleHGD(inRange, outRange, mid, coins)
		if err != nil {
			return 0, err
		}

		if ciphertext <= mid {
			inRange = ValueRange{Start: inEdge + 1, End: x}
			outRange = ValueRange{Start: outEdge + 1, End: mid}
		} else {
			inRange = ValueRange{Start: x + 1, End: inEdge + inSize}
			outRange = ValueRange{Start: mid + 1, End: outEdge + outSize}
		}
	}
}
