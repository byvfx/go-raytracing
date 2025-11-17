package rt

import "math"

type Quad struct {
	Q      Point3
	u      Vec3
	v      Vec3
	w      Vec3
	mat    Material
	bbox   AABB
	normal Vec3
	D      float64
}

func NewQuad(Q Point3, u, v Vec3, mat Material) *Quad {
	quad := &Quad{
		Q:   Q,
		u:   u,
		v:   v,
		mat: mat,
	}

	n := Cross(u, v)
	quad.normal = n.Unit()

	// Calculate the plane constant D
	quad.D = Dot(quad.normal, Q)
	quad.w = n.Scale(1.0 / Dot(n, n))

	quad.setBoundingBox()
	return quad
}
func (q *Quad) setBoundingBox() {
	bboxDiagonal1 := NewAABBFromPoints(q.Q, q.Q.Add(q.u).Add(q.v))
	bboxDiagonal2 := NewAABBFromPoints(q.Q.Add(q.u), q.Q.Add(q.v))
	q.bbox = NewAABBFromBoxes(bboxDiagonal1, bboxDiagonal2)
}

func (q *Quad) BoundingBox() AABB {
	return q.bbox
}

func (q *Quad) Hit(r Ray, rayT Interval, rec *HitRecord) bool {
	denom := Dot(q.normal, r.Direction())

	if math.Abs(denom) < 1e-8 {
		return false
	}

	// Return false if the hit point parameter t is outside the ray interval
	t := (q.D - Dot(q.normal, r.Origin())) / denom
	if !rayT.Contains(t) {
		return false
	}

	intersection := r.At(t)
	planarHitptVector := intersection.Sub(q.Q)
	alpha := Dot(q.w, Cross(planarHitptVector, q.v))
	beta := Dot(q.w, Cross(q.u, planarHitptVector))

	if !q.isInterior(alpha, beta, rec) {
		return false
	}
	rec.T = t
	rec.P = intersection
	rec.Mat = q.mat
	rec.SetFaceNormal(r, q.normal)

	return true
}
func (q *Quad) isInterior(a, b float64, rec *HitRecord) bool {
	unitInterval := NewInterval(0, 1)

	// Given the hit point in plane coordinates, return false if it is outside the
	// primitive, otherwise set the hit record UV coordinates and return true
	if !unitInterval.Contains(a) || !unitInterval.Contains(b) {
		return false
	}

	rec.U = a
	rec.V = b
	return true
}

func Box(a, b Point3, mat Material) Hittable {
	sides := NewHittableList()

	min := Point3{
		X: math.Min(a.X, b.X),
		Y: math.Min(a.Y, b.Y),
		Z: math.Min(a.Z, b.Z),
	}
	max := Point3{
		X: math.Max(a.X, b.X),
		Y: math.Max(a.Y, b.Y),
		Z: math.Max(a.Z, b.Z),
	}

	dx := Vec3{X: max.X - min.X, Y: 0, Z: 0}
	dy := Vec3{X: 0, Y: max.Y - min.Y, Z: 0}
	dz := Vec3{X: 0, Y: 0, Z: max.Z - min.Z}

	// Front face
	sides.Add(NewQuad(Point3{X: min.X, Y: min.Y, Z: max.Z}, dx, dy, mat))
	// Right face
	sides.Add(NewQuad(Point3{X: max.X, Y: min.Y, Z: max.Z}, dz.Neg(), dy, mat))
	// Back face
	sides.Add(NewQuad(Point3{X: max.X, Y: min.Y, Z: min.Z}, dx.Neg(), dy, mat))
	// Left face
	sides.Add(NewQuad(Point3{X: min.X, Y: min.Y, Z: min.Z}, dz, dy, mat))
	// Top face
	sides.Add(NewQuad(Point3{X: min.X, Y: max.Y, Z: max.Z}, dx, dz.Neg(), mat))
	// Bottom face
	sides.Add(NewQuad(Point3{X: min.X, Y: min.Y, Z: min.Z}, dx, dz, mat))

	return sides
}
