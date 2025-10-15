package rt

import "math"

type Interval struct {
	Min, Max float64
}

// the same as  the static const members
var (
	// C++: const interval interval::empty = interval(+infinity, -infinity);
	EmptyInterval = Interval{Min: math.Inf(1), Max: math.Inf(-1)}
	// C++: const interval interval::universe = interval(-infinity, +infinity);
	UniverseInterval = Interval{Min: math.Inf(-1), Max: math.Inf(1)}
)

// C++: interval(double min, double max) : min(min), max(max) {}
func NewInterval(min, max float64) Interval {
	return Interval{Min: min, Max: max}
}

// C++: interval() : min(+infinity), max(-infinity) {}
func NewEmptyInterval() Interval {
	return EmptyInterval
}

// C++: double size() const { return max - min; }
func (i Interval) Size() float64 {
	return i.Max - i.Min
}

//bool contains(double x) const { return min <= x && x <= max; }
func (i Interval) Contains(x float64) bool {
	return i.Min <= x && x <= i.Max
}

//bool surrounds(double x) const {
func (i Interval) Surrounds(x float64) bool {
	return i.Min < x && x < i.Max
}
