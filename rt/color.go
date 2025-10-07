package rt

import "math"

func Clamp(x, min, max float64) float64 {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

func (c Color) ToRGB(samplesPerPixel int) (int, int, int) {
	scale := 1.0 / float64(samplesPerPixel)
	r := math.Sqrt(c.X * scale)
	g := math.Sqrt(c.Y * scale)
	b := math.Sqrt(c.Z * scale)

	ri := int(256 * Clamp(r, 0.0, 0.999))
	gi := int(256 * Clamp(g, 0.0, 0.999))
	bi := int(256 * Clamp(b, 0.0, 0.999))
	return ri, gi, bi
}
