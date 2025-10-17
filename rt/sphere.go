package rt

import "math"

// Sphere represents a sphere object that can be hit by rays
type Sphere struct {
	Center Point3
	Radius float64
	Mat    Material
}

// NewSphere creates a new sphere with the given center and radius and material
func NewSphere(center Point3, radius float64, mat Material) *Sphere {
	return &Sphere{
		Center: center,
		Radius: math.Max(0, radius),
		Mat:    mat,
	}
}

// Hit implements the Hittable interface for Sphere

func (s *Sphere) Hit(r Ray, rayT Interval, rec *HitRecord) bool {

	oc := s.Center.Sub(r.Origin())
	a := r.Direction().Len2()
	h := Dot(r.Direction(), oc)
	c := oc.Len2() - s.Radius*s.Radius

	discriminant := h*h - a*c
	if discriminant < 0 {
		return false
	}

	sqrtd := math.Sqrt(discriminant)

	root := (h - sqrtd) / a
	if !rayT.Surrounds(root) {
		root = (h + sqrtd) / a
		if !rayT.Surrounds(root) {
			return false
		}
	}

	rec.T = root
	rec.P = r.At(rec.T)
	outwardNormal := rec.P.Sub(s.Center).Div(s.Radius)
	rec.SetFaceNormal(r, outwardNormal)
	rec.Mat = s.Mat
	return true
}
