package rt

import "math"

// Sphere represents a sphere object that can be hit by rays
type Sphere struct {
	Center Ray
	Radius float64
	Mat    Material
}

// NewSphere creates a new sphere with the given center and radius and material
func NewSphere(center Point3, radius float64, mat Material) *Sphere {
	return &Sphere{
		Center: NewRay(center, Vec3{X: 0, Y: 0, Z: 0}, 0),
		Radius: math.Max(0, radius),
		Mat:    mat,
	}
}

func NewMovingSphere(center1, center2 Point3, radius float64, mat Material) *Sphere {
	velocity := center2.Sub(center1)
	return &Sphere{
		Center: NewRay(center1, velocity, 0),
		Radius: math.Max(0, radius),
		Mat:    mat,
	}
}

func (s *Sphere) SphereCenter(time float64) Point3 {
	return s.Center.At(time)
}

func getSphereUV(p Point3) (u, v float64) {
	theta := math.Acos(-p.Y)
	phi := math.Atan2(-p.Z, p.X) + math.Pi
	u = phi / (2 * math.Pi)
	v = theta / math.Pi
	return u, v
}

// Hit implements the Hittable interface for Sphere

func (s *Sphere) Hit(r Ray, rayT Interval, rec *HitRecord) bool {

	sphereCenter := s.SphereCenter(r.Time())

	oc := sphereCenter.Sub(r.Origin())
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
	outwardNormal := rec.P.Sub(sphereCenter).Div(s.Radius)
	rec.SetFaceNormal(r, outwardNormal)
	rec.Mat = s.Mat
	return true
}
