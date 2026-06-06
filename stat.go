package goope

// ValueRange represents an inclusive range of integers [Start, End].
type ValueRange struct {
	Start int64
	End   int64
}

// NewValueRange creates a new ValueRange with the given start and end values.
func NewValueRange(start, end int64) (ValueRange, error) {
	if start > end {
		return ValueRange{}, ErrInvalidRangeLimits
	}
	return ValueRange{Start: start, End: end}, nil
}

// Size returns the number of integers in the range.
func (r ValueRange) Size() int64 {
	return r.End - r.Start + 1
}

// Contains checks if the given value is within the range [Start, End].
func (r ValueRange) Contains(value int64) bool {
	return value >= r.Start && value <= r.End
}

// Copy creates a new ValueRange with the same Start and End values.
func (r ValueRange) Copy() ValueRange {
	return ValueRange{Start: r.Start, End: r.End}
}

// sampleHGD maps a specific output-range position (nsample) back to an input-range value, using the hypergeometric distribution as the sampling method.
func sampleHGD(inRange, outRange ValueRange, nsample int64, coins <-chan bool) (int64, error) {
	inSize := inRange.Size()
	outSize := outRange.Size()

	if inSize < 0 || outSize < 0 || inSize > outSize || !outRange.Contains(nsample) {
		return 0, ErrInvalidRanges
	}

	nSampleIndex := nsample - outRange.Start + 1

	if inSize == outSize {
		return inRange.Start + nSampleIndex - 1, nil
	}

	inSampleNum := Rhyper(nSampleIndex, inSize, outSize-inSize, coins)
	if inSampleNum == 0 {
		return inRange.Start, nil
	}
	inSample := inRange.Start + inSampleNum - 1
	if !inRange.Contains(inSample) {
		return 0, ErrInvalidRanges
	}
	return inSample, nil
}

func floorDiv(a, b int64) int64 {
	q := a / b
	// if the result was negative and there's a remainder, round down instead of up
	if (a^b) < 0 && a*b != a {
		q--
	}
	return q
}

// sampleUniform picks a value from inRange using binary search guided by bits from coins. Each bit decides whether to take the lower or upper half of the current range, narrowing it until only one value remains.
func sampleUniform(inRange ValueRange, coins <-chan bool) (int64, error) {
	cur := inRange.Copy()
	if cur.Size() == 0 {
		return 0, ErrInvalidRanges
	}
	for cur.Size() > 1 {
		mid := floorDiv(cur.Start+cur.End, 2)
		bit := <-coins
		if bit == false {
			cur.End = mid
		} else {
			cur.Start = mid + 1
		}
	}
	return cur.Start, nil
}
