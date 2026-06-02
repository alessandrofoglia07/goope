package goope

type ValueRange struct {
	Start int
	End   int
}

func NewValueRange(start, end int) (ValueRange, error) {
	if start > end {
		return ValueRange{}, ErrInvalidRangeLimits
	}
	return ValueRange{Start: start, End: end}, nil
}

func (r ValueRange) Size() int {
	return r.End - r.Start + 1
}

func (r ValueRange) Contains(value int) bool {
	return value >= r.Start && value <= r.End
}

func (r ValueRange) Copy() ValueRange {
	return ValueRange{Start: r.Start, End: r.End}
}

func sampleHGD(inRange, outRange ValueRange, nsample int, coins <-chan bool) int {
	inSize := inRange.Size()
	outSize := outRange.Size()
	nSampleIndex := nsample - outRange.Start + 1

	if inSize == outSize {
		return inRange.Start + nSampleIndex - 1
	}

	inSampleNum := Rhyper(nSampleIndex, inSize, outSize-inSize, coins)
	if inSampleNum == 0 {
		return inRange.Start
	}
	return inRange.Start + inSampleNum - 1
}

func sampleUniform(inRange ValueRange, coins <-chan bool) int {
	cur := inRange.Copy()
	for cur.Size() > 1 {
		mid := (cur.Start + cur.End) / 2
		bit := <-coins
		if bit == false {
			cur.End = mid
		} else {
			cur.Start = mid + 1
		}
	}
	return cur.Start
}
