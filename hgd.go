package goope

import "math"

// PRNG is a pseudo-random number generator that uses a channel of coin flips (bits) to produce random numbers.
type PRNG struct {
	coins <-chan bool
}

// Draw generates a random float64 in the range [0, 1] by consuming 32 bits from the coins channel and converting them into a uint32, which is then normalized to the range [0, 1].
func (p *PRNG) Draw() float64 {
	var out uint32
	for i := 0; i < 32; i++ {
		bit := <-p.coins
		out = (out << 1) | boolToUint32(bit)
	}
	return float64(out) / float64(uint32(0xFFFFFFFF)) // 2^32 - 1
}

// afc returns ln(i!), the natural log of i-factorial, using Stirling's approximation
func afc(i int) float64 {
	if i < 0 {
		panic("afc: i should not be < 0")
	}
	if i == 0 {
		return 0
	}
	frac12 := 1.0 / 12.0
	frac360 := 1.0 / 360.0
	frac_pi := 0.5 * math.Log(2*math.Pi)
	fi := float64(i)
	return (fi+0.5)*math.Log(fi) - fi + frac12/fi - frac360/(fi*fi*fi) + frac_pi
}

// loggam returns ln(gamma(x)), the natural log of the gamma function at x, using the algorithm from "Computation of Special Functions", 1996, John Wiley & Sons, Inc.
func loggam(x float64) float64 {
	a := [10]float64{
		8.333333333333333e-02, -2.777777777777778e-03,
		7.936507936507937e-04, -5.952380952380952e-04,
		8.417508417508418e-04, -1.917526917526918e-03,
		6.410256410256410e-03, -2.955065359477124e-02,
		1.796443723688307e-01, -1.39243221690590e+00,
	}
	x0 := x
	n := 0

	if x == 1.0 || x == 2.0 {
		return 0.0
	} else if x <= 7.0 {
		n = int(7.0 - x)
		x0 = x + float64(n)
	}

	x2 := 1.0 / (x0 * x0)
	xp := 2 * math.Pi
	gl0 := a[9]
	for k := 8; k >= 0; k-- {
		gl0 = gl0*x2 + a[k]
	}
	gl := gl0/x0 + 0.5*math.Log(xp) + (x0-0.5)*math.Log(x0) - x0
	if x <= 7.0 {
		for k := 1; k <= n; k++ {
			gl -= math.Log(x0 - float64(k))
			x0 -= 1.0
		}
	}
	return gl
}

// Rhyper returns a hypergeometric random variate: the number of "good" items drawn when drawing kk items from a pool of nn1 "good" items and nn2 "bad" items, without replacement.
func Rhyper(kk, nn1, nn2 int, coins <-chan bool) int {
	prng := &PRNG{coins: coins}
	if kk > 10 {
		return hypergeometricHRUA(prng, nn1, nn2, kk)
	}
	return hypergeometricHYP(prng, nn1, nn2, kk)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// hypergeometricHYP implements the hypergeometric distribution using the HYP algorithm, which is efficient for small sample sizes (kk <= 10)
func hypergeometricHYP(prng *PRNG, good, bad, sample int) int {
	d1 := bad + good - sample
	d2 := float64(min(bad, good))

	Y := d2
	K := sample
	for Y > 0.0 {
		U := prng.Draw()
		Y -= math.Floor(U + Y/float64(d1+K))
		K--
		if K == 0 {
			break
		}
	}
	Z := int(d2 - Y)
	if good > bad {
		Z = sample - Z
	}
	return Z
}

// hypergeometricHRUA implements the hypergeometric distribution using the HRUA (Hypergeometric Ratio of Uniforms with Aliasing) algorithm, which is efficient for larger sample sizes (kk > 10). The d4-d11 values are precomputed constants that make the accept/reject process more efficient.
func hypergeometricHRUA(prng *PRNG, good, bad, sample int) int {
	const D1 = 1.7155277699214135
	const D2 = 0.8989161620588988

	mingoodbad := min(good, bad)
	popsize := good + bad
	maxgoodbad := max(good, bad)
	m := min(sample, popsize-sample)

	d4 := float64(mingoodbad) / float64(popsize)
	d5 := 1.0 - d4
	d6 := float64(m)*d4 + 0.5
	d7 := math.Sqrt(float64(popsize-m)*float64(sample)*d4*d5/float64(popsize-1) + 0.5)
	d8 := D1*d7 + D2
	d9 := int(math.Floor(float64(m+1) * float64(mingoodbad+1) / float64(popsize+2)))
	d10 := loggam(float64(d9+1)) + loggam(float64(mingoodbad-d9+1)) +
		loggam(float64(m-d9+1)) + loggam(float64(maxgoodbad-m+d9+1))
	d11 := math.Min(float64(min(m, mingoodbad))+1.0, math.Floor(d6+16*d7))

	var Z int
	for {
		X := prng.Draw()
		Y := prng.Draw()
		W := d6 + d8*(Y-0.5)/X

		// fast rejection
		if W < 0.0 || W >= d11 {
			continue
		}

		Z = int(math.Floor(W))
		T := d10 - (loggam(float64(Z+1)) + loggam(float64(mingoodbad-Z+1)) +
			loggam(float64(m-Z+1)) + loggam(float64(maxgoodbad-m+Z+1)))

		// fast acceptance
		if (X*(4.0-X) - 3.0) <= T {
			break
		}

		// fast rejection
		if X*(X-T) >= 1.0 {
			continue
		}

		// acceptance
		if 2.0*math.Log(X) <= T {
			break
		}
	}

	if good > bad {
		Z = m - Z
	}
	if m < sample {
		Z = good - Z
	}
	return Z
}
