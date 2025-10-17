package rt

import "math"

type Material interface {
	Scatter(rIn Ray, rec *HitRecord, attenuation *Color, scattered *Ray) bool
}

type Lambertian struct {
	Albedo Color
}

func NewLambertian(albedo Color) *Lambertian {
	return &Lambertian{
		Albedo: albedo,
	}
}

func (l *Lambertian) Scatter(rIn Ray, rec *HitRecord, attenuation *Color, scattered *Ray) bool {
	scatterDirection := rec.Normal.Add(RandomUnitVector())

	*scattered = NewRay(rec.P, scatterDirection)

	*attenuation = l.Albedo

	return true
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

	// Create the scattered ray
	// C++: scattered = ray(rec.p, reflected);
	*scattered = NewRay(rec.P, reflected)

	// Set the attenuation (color of the metal)
	// C++: attenuation = albedo;
	*attenuation = m.Albedo

	// Only scatter if reflected ray is above the surface (not absorbed into it)
	// C++: return (dot(scattered.direction(), rec.normal) > 0);
	return Dot(scattered.Direction(), rec.Normal) > 0
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
	*scattered = NewRay(rec.P, direction)

	return true
}
