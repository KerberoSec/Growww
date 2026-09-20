package bloom

import (
	"crypto/sha256"
	"encoding/binary"
	"math"
)

// CalculateOptimalParams computes the optimal bit size m and hash function count k
// given expected capacity n and target false positive rate p.
// m = - (n * ln(p)) / (ln(2)^2)
// k = (m / n) * ln(2)
func CalculateOptimalParams(expectedElements uint64, falsePositiveRate float64) (uint64, uint32) {
	if expectedElements == 0 {
		expectedElements = 1000
	}
	if falsePositiveRate <= 0.0 || falsePositiveRate >= 1.0 {
		falsePositiveRate = 0.001 // 0.1% default
	}

	ln2 := math.Ln2
	ln2Sq := ln2 * ln2

	mFloat := -float64(expectedElements) * math.Log(falsePositiveRate) / ln2Sq
	m := uint64(math.Ceil(mFloat))
	if m < 64 {
		m = 64
	}

	kFloat := (float64(m) / float64(expectedElements)) * ln2
	k := uint32(math.Ceil(kFloat))
	if k < 1 {
		k = 1
	}

	return m, k
}

// CalculateFalsePositiveRate calculates the theoretical false positive rate
// p = (1 - e^(-k * n / m))^k
func CalculateFalsePositiveRate(m uint64, k uint32, n uint64) float64 {
	if m == 0 || n == 0 || k == 0 {
		return 0.0
	}
	exponent := -float64(k) * float64(n) / float64(m)
	base := 1.0 - math.Exp(exponent)
	return math.Pow(base, float64(k))
}

// HashValues generates k hash indices for a given data slice using Kirsch-Mitzenmacher double hashing:
// g_i(x) = (h1(x) + i * h2(x) + i^2) mod m
func HashValues(data []byte, m uint64, k uint32) []uint64 {
	h := sha256.Sum256(data)
	h1 := binary.LittleEndian.Uint64(h[0:8])
	h2 := binary.LittleEndian.Uint64(h[8:16])

	// Ensure h2 is non-zero
	if h2 == 0 {
		h2 = 1
	}

	indices := make([]uint64, k)
	for i := uint32(0); i < k; i++ {
		idx := (h1 + uint64(i)*h2 + uint64(i*i)) % m
		indices[i] = idx
	}

	return indices
}
