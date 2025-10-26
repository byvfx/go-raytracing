package rt

import "math"

type Texture interface {
	Value(u, v float64, p Point3) Color
}

type SolidColor struct {
	Albedo Color
}

type CheckerTexture struct {
	invScale float64
	even     Texture
	odd      Texture
}

func NewSolidColor(albedo Color) *SolidColor {
	return &SolidColor{
		Albedo: albedo,
	}
}

func NewSolidColorRGB(red, green, blue float64) *SolidColor {
	return &SolidColor{
		Albedo: Color{X: red, Y: green, Z: blue},
	}
}

func (s *SolidColor) Value(u, v float64, p Point3) Color {
	return s.Albedo
}
func NewCheckerTexture(scale float64, even, odd Texture) *CheckerTexture {
	return &CheckerTexture{
		invScale: 1.0 / scale,
		even:     even,
		odd:      odd,
	}
}
func NewCheckerTextureFromColors(scale float64, c1, c2 Color) *CheckerTexture {
	return NewCheckerTexture(scale, NewSolidColor(c1), NewSolidColor(c2))
}

func (c *CheckerTexture) Value(u, v float64, p Point3) Color {
	xInteger := int(math.Floor(c.invScale * p.X))
	yInteger := int(math.Floor(c.invScale * p.Y))
	zInteger := int(math.Floor(c.invScale * p.Z))

	isEven := (xInteger+yInteger+zInteger)%2 == 0

	if isEven {
		return c.even.Value(u, v, p)
	}
	return c.odd.Value(u, v, p)
}
