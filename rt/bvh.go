package rt

import (
	"sort"
)

type BVHNode struct {
	left  Hittable
	right Hittable
	bbox  AABB
}

func NewBVHNodeFromList(list *HittableList) *BVHNode {
	objects := make([]Hittable, len(list.Objects))
	copy(objects, list.Objects)
	return NewBVHNode(objects, 0, len(objects))
}
func NewBVHNode(objects []Hittable, start, end int) *BVHNode {
	node := &BVHNode{}

	// Choose a random axis to split on
	axis := RandomInt(0, 2)

	// Create comparator function for sorting
	comparator := boxCompare(axis)

	objectSpan := end - start

	switch objectSpan {
	case 1:
		// Only one object - both children point to same object
		node.left = objects[start]
		node.right = objects[start]
	case 2:
		// Two objects - assign one to each child
		if comparator(objects[start], objects[start+1]) {
			node.left = objects[start]
			node.right = objects[start+1]
		} else {
			node.left = objects[start+1]
			node.right = objects[start]
		}
	default:
		// More than two objects - sort and recursively split
		sort.Slice(objects[start:end], func(i, j int) bool {
			return comparator(objects[start+i], objects[start+j])
		})

		mid := start + objectSpan/2
		node.left = NewBVHNode(objects, start, mid)
		node.right = NewBVHNode(objects, mid, end)
	}

	// Compute bounding box that encompasses both children
	node.bbox = NewAABBFromBoxes(node.left.BoundingBox(), node.right.BoundingBox())

	return node
}

func (b *BVHNode) Hit(r Ray, rayT Interval, rec *HitRecord) bool {
	// First check if ray hits this node's bounding box
	if !b.bbox.Hit(r, rayT) {
		return false
	}

	// Check left child
	hitLeft := b.left.Hit(r, rayT, rec)

	// Check right child (only up to closest hit so far)
	hitRight := b.right.Hit(r, NewInterval(rayT.Min, func() float64 {
		if hitLeft {
			return rec.T
		}
		return rayT.Max
	}()), rec)

	return hitLeft || hitRight
}

// BoundingBox returns the bounding box for this BVH node
// C++: aabb bounding_box() const override { return bbox; }
func (b *BVHNode) BoundingBox() AABB {
	return b.bbox
}

// boxCompare returns a comparison function for sorting objects along a given axis
func boxCompare(axis int) func(a, b Hittable) bool {
	return func(a, b Hittable) bool {
		aInterval := a.BoundingBox().AxisInterval(axis)
		bInterval := b.BoundingBox().AxisInterval(axis)
		return aInterval.Min < bInterval.Min
	}
}

// Helper comparison functions for each axis
func boxXCompare(a, b Hittable) bool {
	return boxCompare(0)(a, b)
}

func boxYCompare(a, b Hittable) bool {
	return boxCompare(1)(a, b)
}

func boxZCompare(a, b Hittable) bool {
	return boxCompare(2)(a, b)
}
