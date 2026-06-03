package goope

import "testing"

// TestValueRangeSize tests that Size() counts both endpoints as inclusive
func TestValueRangeSize(t *testing.T) {
	tests := []struct {
		start, end, want int
	}{
		{2, 1000, 999},          // normal range
		{0, 0, 1},               // single element
		{-10, 10, 21},           // straddles zero
		{-100, -50, 51},         // fully negative
		{0, 1<<15 - 1, 1 << 15}, // default in-range
	}
	for _, tt := range tests {
		r := ValueRange{tt.start, tt.end}
		if got := r.Size(); got != tt.want {
			t.Errorf("ValueRange{%d, %d}.Size() = %d, want %d", tt.start, tt.end, got, tt.want)
		}
	}
}

// TestValueRangeContains checks boundary conditions for Contains()
func TestValueRangeContains(t *testing.T) {
	r := ValueRange{2, 1000}
	for _, v := range []int{2, 3, 500, 999, 1000} {
		if !r.Contains(v) {
			t.Errorf("ValueRange{%d, %d}.Contains(%d) = false, want true", r.Start, r.End, v)
		}
	}
	for _, v := range []int{1, 0, -1, 1001, 10000} {
		if r.Contains(v) {
			t.Errorf("ValueRange{%d, %d}.Contains(%d) = true, want false", r.Start, r.End, v)
		}
	}
}

// TestValueRangeCopy checks that Copy() creates an independent copy of the ValueRange
func TestValueRangeCopy(t *testing.T) {
	original := ValueRange{10, 20}
	copy := original.Copy()
	if copy != original {
		t.Fatalf("Copy() did not produce an equal ValueRange: got %v, want %v", copy, original)
	}
	copy.Start = 99
	if original.Start == 99 {
		t.Fatalf("Copy() did not produce an independent copy: modifying copy changed original to %v", original)
	}
}

// TestNewValueRangeRejectsInvalidLmits check that start > end is rejected
func TestNewValueRangeRejectsInvalidLimits(t *testing.T) {
	_, err := NewValueRange(10, 5)
	if err != ErrInvalidRangeLimits {
		t.Fatalf("NewValueRange(10, 5) did not return ErrInvalidRangeLimits: got %v", err)
	}
	r, err := NewValueRange(7, 7)
	if err != nil {
		t.Fatalf("NewValueRange(7, 7) returned an error for a valid range: got %v", err)
	}
	if r.Size() != 1 {
		t.Fatalf("NewValueRange(7, 7) did not create a range of size 1: got size %d", r.Size())
	}
}

// makeBoolChan sends a fixed sequence of bits and then blocks forever, so a test that reads too many bits will hang rather than silently pass
func makeBoolChan(bits []int) <-chan bool {
	ch := make(chan bool, len(bits)+1)
	for _, b := range bits {
		ch <- b != 0
	}
	return ch
}

// TestSampleUniformUnitRange checks that a single-element range always returns that element, consuming no bits at all
func TestSampleUniformUnitRange(t *testing.T) {
	for _, v := range []int{0, 10, -5, 1 << 20} {
		r := ValueRange{v, v}
		got, err := sampleUniform(r, makeBoolChan(nil))
		if err != nil {
			t.Fatalf("sampleUniform(%v, ...) returned error: %v", r, err)
		}
		if got != v {
			t.Fatalf("sampleUniform(%v, ...) = %d, want %d", r, got, v)
		}
	}
}

// TestSampleUniformTwoBitRange checks that a two-element range correctly picks the lower element on bit=0 and the upper element on bit=1
func TestSampleUniformTwoBitRange(t *testing.T) {
	r := ValueRange{10, 11}
	got, _ := sampleUniform(r, makeBoolChan([]int{0}))
	if got != 10 {
		t.Fatalf("sampleUniform(%v, ...) with bit=0 = %d, want %d", r, got, 10)
	}
	got, _ = sampleUniform(r, makeBoolChan([]int{1}))
	if got != 11 {
		t.Fatalf("sampleUniform(%v, ...) with bit=1 = %d, want %d", r, got, 11)
	}
}

// TestSampleUniformMediumRange checks specific bit sequences against a 16-element range [20, 35]
func TestSampleUniformMediumRange(t *testing.T) {
	r := ValueRange{20, 35}

	tests := []struct {
		bits []int
		want int
	}{
		{[]int{0, 0, 0, 0}, 20}, // all-left  -> start
		{[]int{0, 0, 0, 1}, 21}, // one step right of start
		{[]int{1, 1, 1, 1}, 35}, // all-right -> end
	}
	for _, tt := range tests {
		got, err := sampleUniform(r, makeBoolChan(tt.bits))
		if err != nil {
			t.Fatalf("sampleUniform(%v, ...) with bits=%v returned error: %v", r, tt.bits, err)
		}
		if got != tt.want {
			t.Fatalf("sampleUniform(%v, ...) with bits=%v = %d, want %d", r, tt.bits, got, tt.want)
		}
	}
}

// TestSampleUniformNegativeRange checks that sampleUniform works correctly when both endpoints are negative
func TestSampleUniformNegativeRange(t *testing.T) {
	r := ValueRange{-32, -17} // size 16
	allZeros := make([]int, 10)
	allOnes := make([]int, 10)
	for i := range allOnes {
		allOnes[i] = 1
	}

	got, _ := sampleUniform(r, makeBoolChan(allZeros))
	if got != -32 {
		t.Fatalf("sampleUniform(%v, ...) with all zeros = %d, want %d", r, got, -32)
	}
	got, _ = sampleUniform(r, makeBoolChan(allOnes))
	if got != -17 {
		t.Fatalf("sampleUniform(%v, ...) with all ones = %d, want %d", r, got, -17)
	}
}

// TestSampleUniformMixedRange checks a range that straddles zero [-32, 31]
func TestSampleUniformMixedRange(t *testing.T) {
	r := ValueRange{-32, 31} // size 64, needs 6 bits
	allZeros := make([]int, 10)
	allOnes := make([]int, 10)
	for i := range allOnes {
		allOnes[i] = 1
	}

	got, _ := sampleUniform(r, makeBoolChan(allZeros))
	if got != -32 {
		t.Fatalf("sampleUniform(%v, ...) with all zeros = %d, want %d", r, got, -32)
	}
	got, _ = sampleUniform(r, makeBoolChan(allOnes))
	if got != 31 {
		t.Fatalf("sampleUniform(%v, ...) with all ones = %d, want %d", r, got, 31)
	}
}

// TestSampleUniformResultAlwaysInRange checks that no matter how many bits are read, the result always lands inside the range
func TestSampleUniformResultAlwaysInRange(t *testing.T) {
	r := ValueRange{100, 200}
	patterns := [][]int{
		{0, 0, 0, 0, 0, 0, 0},
		{1, 1, 1, 1, 1, 1, 1},
		{1, 0, 1, 0, 1, 0, 1},
		{0, 1, 0, 1, 0, 1, 0},
		{1, 1, 0, 0, 1, 0, 1},
	}
	for _, bits := range patterns {
		got, err := sampleUniform(r, makeBoolChan(bits))
		if err != nil {
			t.Fatalf("sampleUniform(%v, ...) with bits=%v returned error: %v", r, bits, err)
		}
		if !r.Contains(got) {
			t.Fatalf("sampleUniform(%v, ...) with bits=%v = %d, which is outside the range", r, bits, got)
		}
	}
}

// TestSampleHGDEqualRanges checks the fast path: when in_size == out_size, sampleHGD must return in_range.start + (n_sample - out_range.start)
func TestSampleHGDEqualRanges(t *testing.T) {
	inR := ValueRange{0, 9}
	outR := ValueRange{0, 9}
	for nsample := 0; nsample < 9; nsample++ {
		got, err := sampleHGD(inR, outR, nsample, makeCoinChan([]int{0}))
		if err != nil {
			t.Fatalf("sampleHGD(%v, %v, %d, ...) returned error: %v", inR, outR, nsample, err)
		}
		if got != nsample {
			t.Fatalf("sampleHGD(%v, %v, %d, ...) = %d, want %d", inR, outR, nsample, got, nsample)
		}
	}
}

// TestSampleHGDResultAlwaysInInputRange checks that sampleHGD always returns a value inside in_range, for a variety of inputs
func TestSampleHGDResultAlwaysInInputRange(t *testing.T) {
	inR := ValueRange{0, 9}
	outR := ValueRange{0, 99}
	for nsample := 0; nsample <= outR.End; nsample += 7 {
		got, err := sampleHGD(inR, outR, nsample, realCoinChan(int64(nsample)))
		if err != nil {
			t.Fatalf("sampleHGD(%v, %v, %d, ...) returned error: %v", inR, outR, nsample, err)
		}
		if !inR.Contains(got) {
			t.Fatalf("sampleHGD(%v, %v, %d, ...) = %d, which is outside the input range", inR, outR, nsample, got)
		}
	}
}

// TestSampleHGDNSampleAtBoundaries checks the edge case where n_sample is exactly at the start or end of the output range
func TestSampleHGDNSampleAtBoundaries(t *testing.T) {
	inR := ValueRange{0, 9}
	outR := ValueRange{0, 99}
	for _, nsample := range []int{outR.Start, outR.End} {
		got, err := sampleHGD(inR, outR, nsample, realCoinChan(int64(nsample)))
		if err != nil {
			t.Fatalf("sampleHGD(%v, %v, %d, ...) returned error: %v", inR, outR, nsample, err)
		}
		if !inR.Contains(got) {
			t.Fatalf("sampleHGD(%v, %v, %d, ...) = %d, which is outside the input range", inR, outR, nsample, got)
		}
	}
}
