package rt

import "math"

var intensity = NewInterval(0.000, 0.999)

func (c Color) ToRGB(samplesPerPixel int) (int, int, int) {
	scale := 1.0 / float64(samplesPerPixel)
	r := math.Sqrt(c.X * scale)
	g := math.Sqrt(c.Y * scale)
	b := math.Sqrt(c.Z * scale)

	ri := int(256 * intensity.Clamp(r))
	gi := int(256 * intensity.Clamp(g))
	bi := int(256 * intensity.Clamp(b))
	return ri, gi, bi
}
