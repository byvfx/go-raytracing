package rt

import "math"

var intensity = NewInterval(0.000, 0.999)

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

	ri := int(256 * Clamp(r, intensity.Min, intensity.Max))
	gi := int(256 * Clamp(g, intensity.Min, intensity.Max))
	bi := int(256 * Clamp(b, intensity.Min, intensity.Max))
	return ri, gi, bi
}
