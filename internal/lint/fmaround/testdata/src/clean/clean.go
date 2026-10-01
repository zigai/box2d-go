package clean

type Vec2 struct{ X, Y float32 }

const k = 3.0

func rounded(a, b, c float32, v Vec2, s []float32) float32 {
	_ = float32(a*b) + c
	_ = c + float32(a*b)
	_ = float32(a*b) - c
	_ = c - float32(a*b)
	_ = float32(v.X*v.X) + float32(v.Y*v.Y)
	_ = (float32(a * b)) + c
	_ = -float32(a*b) + c
	c += float32(a * b)
	c -= float32(a * b)
	s[0] += float32(a * b)
	return c
}

func notFusable(a, b, c float32, x, y, z float64, i, j, n int) {
	// A product that does not feed an addition or subtraction.
	_ = a * b
	_ = a * b * c
	_ = a * b / c
	_ = a*b < c
	_ = a * (b + c)
	_ = (a + b) * c
	c *= a + b
	c /= a * b

	// Constant expressions are exact at compile time.
	_ = k*k + a
	_ = a + k*2

	// Integers and float64 are outside this analyzer.
	_ = i*j + n
	n += i * j
	_ = x*y + z
	z += x * y
	_ = float64(a)*float64(b) + float64(c)
	_ = float32(x*y) + c
}
