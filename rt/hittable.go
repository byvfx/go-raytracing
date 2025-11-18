package rt

import "math"

// HitRecord stores information about a ray-object intersection
type HitRecord struct {
	P         Point3
	Normal    Vec3
	Mat       Material
	U         float64
	V         float64
	T         float64 // Parameter t where intersection occurs
	FrontFace bool
}

// Hittable interface for objects that can be hit by rays
type Hittable interface {
	Hit(r Ray, rayT Interval, rec *HitRecord) bool
	BoundingBox() AABB
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

type Translate struct {
	Obj    Hittable
	Offset Vec3
	bbox   AABB
}

func NewTranslate(obj Hittable, offset Vec3) *Translate {
	bbox := obj.BoundingBox().Translate(offset)
	return &Translate{
		Obj:    obj,
		Offset: offset,
		bbox:   bbox,
	}
}

func (t *Translate) Hit(r Ray, rayT Interval, rec *HitRecord) bool {
	offsetRay := NewRay(
		r.Origin().Sub(t.Offset),
		r.Direction(),
		r.Time(),
	)

	if !t.Obj.Hit(offsetRay, rayT, rec) {
		return false
	}
	rec.P = rec.P.Add(t.Offset)
	return true
}

func (t *Translate) BoundingBox() AABB {
	return t.bbox
}

type RotateY struct {
	Obj      Hittable
	SinTheta float64
	CosTheta float64
	bbox     AABB
}

func Ry(obj Hittable, angle float64) *RotateY {
	radians := DegreesToRadians(angle)
	sinTheta := math.Sin(radians)
	cosTheta := math.Cos(radians)

	bbox := obj.BoundingBox()
	min := Point3{X: math.Inf(1), Y: math.Inf(1), Z: math.Inf(1)}
	max := Point3{X: math.Inf(-1), Y: math.Inf(-1), Z: math.Inf(-1)}

	for i := range 2 {
		for j := range 2 {
			for k := range 2 {
				x := float64(i)*bbox.X.Max + float64(1-i)*bbox.X.Min
				y := float64(j)*bbox.Y.Max + float64(1-j)*bbox.Y.Min
				z := float64(k)*bbox.Z.Max + float64(1-k)*bbox.Z.Min

				newX := cosTheta*x + sinTheta*z
				newZ := -sinTheta*x + cosTheta*z

				tester := Vec3{X: newX, Y: y, Z: newZ}

				min.X = math.Min(min.X, tester.X)
				max.X = math.Max(max.X, tester.X)
				min.Y = math.Min(min.Y, tester.Y)
				max.Y = math.Max(max.Y, tester.Y)
				min.Z = math.Min(min.Z, tester.Z)
				max.Z = math.Max(max.Z, tester.Z)
			}
		}
	}
	return &RotateY{
		Obj:      obj,
		SinTheta: sinTheta,
		CosTheta: cosTheta,
		bbox:     NewAABBFromPoints(min, max),
	}
}
func (ry *RotateY) Hit(r Ray, rayT Interval, rec *HitRecord) bool {
	// Transform ray from world space to object space
	origin := r.Origin()
	direction := r.Direction()

	origin.X = ry.CosTheta*r.Origin().X - ry.SinTheta*r.Origin().Z
	origin.Z = ry.SinTheta*r.Origin().X + ry.CosTheta*r.Origin().Z

	direction.X = ry.CosTheta*r.Direction().X - ry.SinTheta*r.Direction().Z
	direction.Z = ry.SinTheta*r.Direction().X + ry.CosTheta*r.Direction().Z

	rotatedRay := NewRay(origin, direction, r.Time())

	// Check for intersection in object space
	if !ry.Obj.Hit(rotatedRay, rayT, rec) {
		return false
	}

	// Transform intersection point and normal back to world space
	p := rec.P
	p.X = ry.CosTheta*rec.P.X + ry.SinTheta*rec.P.Z
	p.Z = -ry.SinTheta*rec.P.X + ry.CosTheta*rec.P.Z

	normal := rec.Normal
	normal.X = ry.CosTheta*rec.Normal.X + ry.SinTheta*rec.Normal.Z
	normal.Z = -ry.SinTheta*rec.Normal.X + ry.CosTheta*rec.Normal.Z

	rec.P = p
	rec.Normal = normal

	return true
}

func (ry *RotateY) BoundingBox() AABB {
	return ry.bbox
}
