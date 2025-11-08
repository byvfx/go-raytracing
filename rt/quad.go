package rt

type Quad struct {
	Q    Point3
	u    Vec3
	v    Vec3
	mat  Material
	bbox AABB
}

func NewQuad(q Point3, u, v Vec3, mat Material) *Quad {
	quad := &Quad{
		Q:   q,
		u:   u,
		v:   v,
		mat: mat,
	}
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

func (q *Quad) Hit(r Ray, rayT Interval) (*HitRecord, bool) {
	return nil, false // todo implement quad ray intersection
}
