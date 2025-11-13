package rt

import "math"

type Material interface {
	Scatter(rIn Ray, rec *HitRecord, attenuation *Color, scattered *Ray) bool
	Emitted(u, v float64, p Point3) Color
}

type Lambertian struct {
	tex Texture
}

func NewLambertian(albedo Color) *Lambertian {
	return &Lambertian{
		tex: NewSolidColor(albedo),
	}
}

func NewLambertianTexture(tex Texture) *Lambertian {
	return &Lambertian{
		tex: tex,
	}
}
func (l *Lambertian) Scatter(rIn Ray, rec *HitRecord, attenuation *Color, scattered *Ray) bool {
	scatterDirection := rec.Normal.Add(RandomUnitVector())

	if scatterDirection.NearZero() {
		scatterDirection = rec.Normal
	}

	*scattered = NewRay(rec.P, scatterDirection, rIn.Time())

	*attenuation = l.tex.Value(rec.U, rec.V, rec.P)

	return true
}
func (l *Lambertian) Emitted(u, v float64, p Point3) Color {
	return Color{X: 0, Y: 0, Z: 0}
}

type Metal struct {
	Albedo Color
	Fuzz   float64
}

func NewMetal(albedo Color, fuzz float64) *Metal {
	if fuzz > 1 {
		fuzz = 1
	}
	return &Metal{
		Albedo: albedo,
		Fuzz:   fuzz,
	}
}

func (m *Metal) Scatter(rIn Ray, rec *HitRecord, attenuation *Color, scattered *Ray) bool {
	reflected := Reflect(rIn.Direction(), rec.Normal)
	reflected = reflected.Unit().Add(RandomUnitVector().Scale(m.Fuzz))
	*scattered = NewRay(rec.P, reflected, rIn.Time())
	*attenuation = m.Albedo
	return Dot(scattered.Direction(), rec.Normal) > 0
}

func (m *Metal) Emitted(u, v float64, p Point3) Color {
	return Color{X: 0, Y: 0, Z: 0}
}

type Dielectric struct {
	RefractionIndex float64
}

func NewDielectric(refractionIndex float64) *Dielectric {
	return &Dielectric{
		RefractionIndex: refractionIndex,
	}
}

func reflectance(cosine, refractionIndex float64) float64 {
	r0 := (1 - refractionIndex) / (1 + refractionIndex)
	r0 = r0 * r0
	return r0 + (1-r0)*math.Pow(1-cosine, 5)
}

func (d *Dielectric) Scatter(rIn Ray, rec *HitRecord, attenuation *Color, scattered *Ray) bool {
	*attenuation = Color{X: 1.0, Y: 1.0, Z: 1.0}

	var ri float64
	if rec.FrontFace {
		ri = 1.0 / d.RefractionIndex
	} else {
		ri = d.RefractionIndex
	}
	unitDirection := rIn.Direction().Unit()
	cosTheta := math.Min(Dot(unitDirection.Neg(), rec.Normal), 1.0)
	sinTheta := math.Sqrt(1.0 - cosTheta*cosTheta)
	cannotRefract := ri*sinTheta > 1.0

	var direction Vec3

	if cannotRefract || reflectance(cosTheta, ri) > RandomDouble() {
		direction = Reflect(unitDirection, rec.Normal)
	} else {
		direction = Refract(unitDirection, rec.Normal, ri)
	}
	*scattered = NewRay(rec.P, direction, rIn.Time())

	return true
}
func (d *Dielectric) Emitted(u, v float64, p Point3) Color {
	return Color{X: 0, Y: 0, Z: 0}
}

// light material
type DiffuseLight struct {
	tex Texture
}

func NewDiffuseLight(tex Texture) *DiffuseLight {
	return &DiffuseLight{
		tex: tex,
	}
}
func NewDiffuseLightColor(emit Color) *DiffuseLight {
	return &DiffuseLight{
		tex: NewSolidColor(emit),
	}
}

func (dl *DiffuseLight) Scatter(rIn Ray, rec *HitRecord, attenuation *Color, scattered *Ray) bool {
	return false
}

func (dl *DiffuseLight) Emitted(u, v float64, p Point3) Color {
	return dl.tex.Value(u, v, p)
}
