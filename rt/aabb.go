package rt

type AABB struct {
	X, Y, Z Interval
}

func NewAABB() AABB {
	return AABB{
		X: NewEmptyInterval(),
		Y: NewEmptyInterval(),
		Z: NewEmptyInterval(),
	}
}

func NewAABBFromIntervals(x, y, z Interval) AABB {
	return AABB{X: x, Y: y, Z: z}
}

func NewAABBFromPoints(a, b Point3) AABB {
	var x, y, z Interval

	if a.X <= b.X {
		x = NewInterval(a.X, b.X)
	} else {
		x = NewInterval(b.X, a.X)
	}

	if a.Y <= b.Y {
		y = NewInterval(a.Y, b.Y)
	} else {
		y = NewInterval(b.Y, a.Y)
	}

	if a.Z <= b.Z {
		z = NewInterval(a.Z, b.Z)
	} else {
		z = NewInterval(b.Z, a.Z)
	}

	return AABB{X: x, Y: y, Z: z}
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
