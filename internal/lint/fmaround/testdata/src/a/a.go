package a

type Real = float32

type Meters float32

type Vec2 struct{ X, Y float32 }

func positive(a, b, c float32, v Vec2, m, n Meters, s []float32) float32 {
	_ = a*b + c // want `float32 product feeds an addition or subtraction without explicit rounding`
	_ = c + a*b // want `float32 product feeds an addition`
	_ = a*b - c // want `float32 product feeds an addition`
	_ = c - a*b // want `float32 product feeds an addition`

	_ = (a * b) + c        // want `float32 product feeds an addition`
	_ = c + (a * b)        // want `float32 product feeds an addition`
	_ = -(a * b) + c       // want `float32 product feeds an addition`
	_ = c - -(a * b)       // want `float32 product feeds an addition`
	_ = a*b*c + a          // want `float32 product feeds an addition`
	_ = 2*a + c            // want `float32 product feeds an addition`
	_ = v.X*v.X + v.Y*v.Y  // want `float32 product feeds an addition` `float32 product feeds an addition`
	_ = a*b - c*a          // want `float32 product feeds an addition` `float32 product feeds an addition`
	_ = a*b + float32(c*a) // want `float32 product feeds an addition`
	_ = a + b + c*a        // want `float32 product feeds an addition`

	c += a * b    // want `float32 product feeds \+=`
	c -= a * b    // want `float32 product feeds -=`
	v.X += a * b  // want `float32 product feeds \+=`
	s[0] -= a * b // want `float32 product feeds -=`
	c += (a * b)  // want `float32 product feeds \+=`
	c += -(a * b) // want `float32 product feeds \+=`

	_ = m*n + m // want `float32 product feeds an addition or subtraction without explicit rounding; wrap it in Meters\(\.\.\.\)`

	var r Real = 1
	_ = r*r + r // want `wrap it in Real\(\.\.\.\)`

	return a*b + c // want `float32 product feeds an addition`
}

func inCall(a, b, c float32) {
	use(a*b + c) // want `float32 product feeds an addition`
	use(c + float32(a*b))
}

func use(float32) {}
