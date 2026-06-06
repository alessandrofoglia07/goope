package goope

import (
	"math/rand"
	"testing"
)

func makeCoinChan(pattern []int) <-chan bool {
	ch := make(chan bool, 256)
	go func() {
		for {
			for _, p := range pattern {
				ch <- p == 1
			}
		}
	}()
	return ch
}

// realCoinChan returns a channel of random bits
func realCoinChan(seed int64) <-chan bool {
	ch := make(chan bool, 256)
	go func() {
		r := rand.New(rand.NewSource(seed))
		for {
			ch <- r.Intn(2) == 1
		}
	}()
	return ch
}

func TestRhyper(t *testing.T) {
	tests := []struct {
		kk, nn1, nn2 int64
		pattern      []int
		want         int64
	}{
		// Small kk -> hypergeometricHYP path
		{kk: 1, nn1: 10, nn2: 10, pattern: []int{1, 0, 1, 0, 1, 0, 1, 0}, want: 1},
		{kk: 5, nn1: 20, nn2: 30, pattern: []int{1, 1, 0, 0, 1, 0, 1, 1}, want: 5},
		{kk: 10, nn1: 50, nn2: 50, pattern: []int{0, 1, 0, 1, 1, 0, 0, 1}, want: 0},
		// Large kk -> hypergeometricHRUA path
		{kk: 11, nn1: 50, nn2: 50, pattern: []int{1, 0, 1, 0, 1, 0, 1, 0}, want: 6},
		{kk: 50, nn1: 100, nn2: 200, pattern: []int{1, 1, 0, 0, 1, 1, 0, 0}, want: 19},
		{kk: 100, nn1: 500, nn2: 500, pattern: []int{0, 1, 1, 0, 0, 1, 1, 0}, want: 48},
	}

	for _, tt := range tests {
		coins := makeCoinChan(tt.pattern)
		got := Rhyper(tt.kk, tt.nn1, tt.nn2, coins)
		if got != tt.want {
			t.Errorf("Rhyper(%d, %d, %d) with pattern %v = %d; want %d", tt.kk, tt.nn1, tt.nn2, tt.pattern, got, tt.want)
		}
	}
}
