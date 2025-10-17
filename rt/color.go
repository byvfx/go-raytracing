package rt

import "math"

var intensity = NewInterval(0.000, 0.999)

func linearToGamma(linearComponent float64) float64 {
	if linearComponent > 0 {
		return math.Sqrt(linearComponent)
	}
	return 0
}

func (c Color) ToRGB(samplesPerPixel int) (int, int, int) {
	scale := 1.0 / float64(samplesPerPixel)
	r := c.X * scale
	g := c.Y * scale
	b := c.Z * scale

	r = linearToGamma(r)
	g = linearToGamma(g)
	b = linearToGamma(b)

	ri := int(256 * intensity.Clamp(r))
	gi := int(256 * intensity.Clamp(g))
	bi := int(256 * intensity.Clamp(b))
	return ri, gi, bi
}
