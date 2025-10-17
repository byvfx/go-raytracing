package rt

// HitRecord stores information about a ray-object intersection
type HitRecord struct {
	P         Point3 // Point of intersection
	Normal    Vec3
	Mat       Material
	T         float64 // Parameter t where intersection occurs
	FrontFace bool
}

// Hittable interface for objects that can be hit by rays
type Hittable interface {
	Hit(r Ray, rayT Interval, rec *HitRecord) bool
}

func (rec *HitRecord) SetFaceNormal(r Ray, outwardNormal Vec3) {
	// Determine if ray is hitting from outside or inside
	rec.FrontFace = Dot(r.Direction(), outwardNormal) < 0

	// Normal always points against the ray direction
	if rec.FrontFace {
		rec.Normal = outwardNormal
	} else {
		rec.Normal = outwardNormal.Neg()
	}
}
