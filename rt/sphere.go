package rt

import "math"

// Sphere represents a sphere object that can be hit by rays
type Sphere struct {
	Center Point3
	Radius float64
}

// NewSphere creates a new sphere with the given center and radius
func NewSphere(center Point3, radius float64) *Sphere {
	return &Sphere{
		Center: center,
		Radius: math.Max(0, radius), // Ensure radius is non-negative
	}
}

// Hit implements the Hittable interface for spheres
func (s *Sphere) Hit(r Ray, rayTMin, rayTMax float64, rec *HitRecord) bool {
	oc := s.Center.Sub(r.Origin())
	a := r.Direction().Len2()
	h := Dot(r.Direction(), oc)
	c := oc.Len2() - s.Radius*s.Radius

	discriminant := h*h - a*c
	if discriminant < 0 {
		return false
	}

	sqrtd := math.Sqrt(discriminant)

	// Find the nearest root that lies in the acceptable range
	root := (h - sqrtd) / a
	if root <= rayTMin || rayTMax <= root {
		root = (h + sqrtd) / a
		if root <= rayTMin || rayTMax <= root {
			return false
		}
	}

	rec.T = root
	rec.P = r.At(rec.T)
	rec.Normal = rec.P.Sub(s.Center).Div(s.Radius)

	return true
}
