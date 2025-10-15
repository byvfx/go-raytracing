package rt

// HitRecord stores information about a ray-object intersection
type HitRecord struct {
	P      Point3  // Point of intersection
	Normal Vec3    // Surface normal at intersection
	T      float64 // Parameter t where intersection occurs
}

// Hittable interface for objects that can be hit by rays
type Hittable interface {
	Hit(r Ray, rayT Interval, rec *HitRecord) bool
}
