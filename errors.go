package goope

import "errors"

var (
	ErrInvalidCiphertext  = errors.New("invalid ciphertext")
	ErrInvalidRangeLimits = errors.New("invalid range limits")
	ErrOutOfRange         = errors.New("value out of range")
	ErrInvalidRanges      = errors.New("invalid ranges")
)
