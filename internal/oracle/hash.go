package oracle

import "math"

const (
	HashInit uint64 = 14695981039346656037

	hashPrime uint64 = 1099511628211
)

func HashBytes(h uint64, p []byte) uint64 {
	for _, b := range p {
		h ^= uint64(b)
		h *= hashPrime
	}
	return h
}

func HashU32(h uint64, v uint32) uint64 {
	return HashBytes(h, []byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)})
}

func HashU64(h uint64, v uint64) uint64 {
	h = HashU32(h, uint32(v))
	return HashU32(h, uint32(v>>32))
}

func HashF32(h uint64, v float32) uint64 { return HashU32(h, math.Float32bits(v)) }

func HashF64(h uint64, v float64) uint64 { return HashU64(h, math.Float64bits(v)) }

func HashPos(h uint64, x, y Real) uint64 {
	if DoublePrecision {
		h = HashU64(h, realBits(x))
		return HashU64(h, realBits(y))
	}
	h = HashU32(h, uint32(realBits(x)))
	return HashU32(h, uint32(realBits(y)))
}

func HashRot(h uint64, c, s float32) uint64 {
	h = HashF32(h, c)
	return HashF32(h, s)
}

func HashTransform(h uint64, px, py Real, qc, qs float32) uint64 {
	h = HashPos(h, px, py)
	return HashRot(h, qc, qs)
}
