package rt

import "math"

type Perlin struct {
	randfloat [256]float64
	permX     [256]int
	permY     [256]int
	permZ     [256]int
}

func NewPerlin() *Perlin {
	p := &Perlin{}

	for i := range 256 {
		p.randfloat[i] = RandomDouble()
	}

	perlinGeneratePerm(&p.permX)
	perlinGeneratePerm(&p.permY)
	perlinGeneratePerm(&p.permZ)

	return p
}

func (p *Perlin) Noise(pt Point3) float64 {
	u := pt.X - math.Floor(pt.X)
	v := pt.Y - math.Floor(pt.Y)
	w := pt.Z - math.Floor(pt.Z)

	u = u * u * (3 - 2*u)
	v = v * v * (3 - 2*v)
	w = w * w * (3 - 2*w)

	i := int(math.Floor(pt.X))
	j := int(math.Floor(pt.Y))
	k := int(math.Floor(pt.Z))

	var c [2][2][2]float64

	for di := range 2 {
		for dj := range 2 {
			for dk := range 2 {
				c[di][dj][dk] = p.randfloat[p.permX[(i+di)&255]^
					p.permY[(j+dj)&255]^
					p.permZ[(k+dk)&255]]
			}
		}
	}

	return trilinearInterp(c, u, v, w)
}
func perlinGeneratePerm(p *[256]int) {
	for i := range 256 {
		p[i] = i
	}
	permute(p, 256)
}

func permute(p *[256]int, n int) {
	for i := n - 1; i > 0; i-- {
		target := RandomInt(0, i)
		p[i], p[target] = p[target], p[i]
	}
}
func trilinearInterp(c [2][2][2]float64, u, v, w float64) float64 {
	accum := 0.0
	for i := range 2 {
		for j := range 2 {
			for k := range 2 {
				accum += (float64(i)*u + (1-float64(i))*(1-u)) *
					(float64(j)*v + (1-float64(j))*(1-v)) *
					(float64(k)*w + (1-float64(k))*(1-w)) *
					c[i][j][k]
			}
		}
	}
	return accum
}
