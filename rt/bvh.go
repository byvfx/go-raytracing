package rt

import (
	"sort"
	"sync"
)

// BVHNode represents a node in a Bounding Volume Hierarchy tree
// RUST PORT NOTE: Use Box<dyn Hittable> for left/right
// or consider enum-based dispatch for better performance:
// enum BVHNode { Leaf(Triangle), Branch { left: Box<BVH>, right: Box<BVH>, bbox: AABB } }
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

// Threshold for when to parallelize BVH construction
// Objects below this count use sequential construction
const bvhParallelThreshold = 1000

func NewBVHNode(objects []Hittable, start, end int) *BVHNode {
	return newBVHNodeParallel(objects, start, end, true)
}

func newBVHNodeParallel(objects []Hittable, start, end int, allowParallel bool) *BVHNode {
	node := &BVHNode{}

	objectSpan := end - start

	// Calculate bounding box for all objects in this range
	var bbox AABB
	for i := start; i < end; i++ {
		if i == start {
			bbox = objects[i].BoundingBox()
		} else {
			bbox = NewAABBFromBoxes(bbox, objects[i].BoundingBox())
		}
	}

	// Choose axis with longest extent (better than random for meshes)
	axis := bbox.LongestAxis()

	// Create comparator function for sorting
	comparator := boxCompare(axis)

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

		// Use parallel construction for large subtrees
		if allowParallel && objectSpan >= bvhParallelThreshold {
			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				node.left = newBVHNodeParallel(objects, start, mid, true)
			}()

			go func() {
				defer wg.Done()
				node.right = newBVHNodeParallel(objects, mid, end, true)
			}()

			wg.Wait()
		} else {
			// Sequential construction for smaller subtrees
			node.left = newBVHNodeParallel(objects, start, mid, false)
			node.right = newBVHNodeParallel(objects, mid, end, false)
		}
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
