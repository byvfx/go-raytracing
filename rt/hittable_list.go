package rt

// HittableList represents a collection of hittable objects
type HittableList struct {
	Objects []Hittable
}

// NewHittableList creates a new empty hittable list
func NewHittableList() *HittableList {
	return &HittableList{
		Objects: make([]Hittable, 0),
	}
}

// Add adds a hittable object to the list
func (hl *HittableList) Add(object Hittable) {
	hl.Objects = append(hl.Objects, object)
}

// Clear removes all objects from the list
func (hl *HittableList) Clear() {
	hl.Objects = hl.Objects[:0]
}

// Hit implements the Hittable interface for the list
// It finds the closest hit among all objects in the list
func (hl *HittableList) Hit(r Ray, rayTMin, rayTMax float64, rec *HitRecord) bool {
	tempRec := &HitRecord{}
	hitAnything := false
	closestSoFar := rayTMax

	for _, object := range hl.Objects {
		if object.Hit(r, rayTMin, closestSoFar, tempRec) {
			hitAnything = true
			closestSoFar = tempRec.T
			*rec = *tempRec
		}
	}

	return hitAnything
}
