package rt

import "math"

type AABB struct {
	X, Y, Z Interval
}

var (
	EmptyAABB    = NewAABB()
	UniverseAABB = NewAABBFromIntervals(
		UniverseInterval,
		UniverseInterval,
		UniverseInterval,
	)
)

func NewAABB() AABB {
	return AABB{
		X: NewEmptyInterval(),
		Y: NewEmptyInterval(),
		Z: NewEmptyInterval(),
	}
}

func NewAABBFromIntervals(x, y, z Interval) AABB {
	box := AABB{X: x, Y: y, Z: z}
	box.padToMinimums()
	return box
}

func NewAABBFromPoints(a, b Point3) AABB {
	box := AABB{
		X: NewInterval(math.Min(a.X, b.X), math.Max(a.X, b.X)),
		Y: NewInterval(math.Min(a.Y, b.Y), math.Max(a.Y, b.Y)),
		Z: NewInterval(math.Min(a.Z, b.Z), math.Max(a.Z, b.Z)),
	}
	box.padToMinimums()
	return box
}

func NewAABBFromBoxes(box0, box1 AABB) AABB {
	return AABB{
		X: NewIntervalFromIntervals(box0.X, box1.X),
		Y: NewIntervalFromIntervals(box0.Y, box1.Y),
		Z: NewIntervalFromIntervals(box0.Z, box1.Z),
	}
}
func (box AABB) AxisInterval(n int) Interval {
	if n == 1 {
		return box.Y
	}
	if n == 2 {
		return box.Z
	}
	return box.X
}

func (box AABB) Hit(r Ray, rayT Interval) bool {
	rayOrig := r.Origin()
	rayDir := r.Direction()

	for axis := 0; axis < 3; axis++ {
		ax := box.AxisInterval(axis)

		var axisOrig, axisDir float64
		switch axis {
		case 0:
			axisOrig = rayOrig.X
			axisDir = rayDir.X
		case 1:
			axisOrig = rayOrig.Y
			axisDir = rayDir.Y
		case 2:
			axisOrig = rayOrig.Z
			axisDir = rayDir.Z
		}

		adinv := 1.0 / axisDir

		t0 := (ax.Min - axisOrig) * adinv
		t1 := (ax.Max - axisOrig) * adinv

		if t0 < t1 {
			if t0 > rayT.Min {
				rayT.Min = t0
			}
			if t1 < rayT.Max {
				rayT.Max = t1
			}
		} else {
			if t1 > rayT.Min {
				rayT.Min = t1
			}
			if t0 < rayT.Max {
				rayT.Max = t0
			}
		}
		if rayT.Max <= rayT.Min {
			return false
		}
	}

	return true
}
func (box *AABB) padToMinimums() {
	delta := 0.0001
	if box.X.Size() < delta {
		box.X = box.X.Expand(delta)
	}
	if box.Y.Size() < delta {
		box.Y = box.Y.Expand(delta)
	}
	if box.Z.Size() < delta {
		box.Z = box.Z.Expand(delta)
	}
}

func (box AABB) Translate(offset Vec3) AABB {
	return NewAABBFromIntervals(
		box.X.Add(offset.X),
		box.Y.Add(offset.Y),
		box.Z.Add(offset.Z),
	)
}
