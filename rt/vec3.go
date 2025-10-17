package rt

import (
	"fmt"
	"math"
)

type Vec3 struct {
	X, Y, Z float64
}

// Vector method
func NewVec3(x, y, z float64) Vec3 { return Vec3{X: x, Y: y, Z: z} }

type Point3 = Vec3
type Color = Vec3

// debug statement if needed
func (v Vec3) String() string { return fmt.Sprintf("%g %g %g", v.X, v.Y, v.Z) }

// Basic Ops

func (v Vec3) Add(u Vec3) Vec3      { return Vec3{v.X + u.X, v.Y + u.Y, v.Z + u.Z} }
func (v Vec3) Sub(u Vec3) Vec3      { return Vec3{v.X - u.X, v.Y - u.Y, v.Z - u.Z} }
func (v Vec3) Mult(u Vec3) Vec3     { return Vec3{v.X * u.X, v.Y * u.Y, v.Z * u.Z} }
func (v Vec3) Scale(t float64) Vec3 { return Vec3{t * v.X, t * v.Y, t * v.Z} }
func (v Vec3) Div(t float64) Vec3   { return v.Scale(1 / t) }
func (v Vec3) Neg() Vec3            { return Vec3{-v.X, -v.Y, -v.Z} }

func (v Vec3) Len2() float64 { return v.X*v.X + v.Y*v.Y + v.Z*v.Z }
func (v Vec3) Len() float64  { return math.Sqrt(v.Len2()) }
func (v Vec3) Unit() Vec3 {
	l := v.Len()
	if l == 0 {
		return v
	}
	return v.Div(l)
}

func Dot(a, b Vec3) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func Cross(a, b Vec3) Vec3 {
	return Vec3{
		X: a.Y*b.Z - a.Z*b.Y,
		Y: a.Z*b.X - a.X*b.Z,
		Z: a.X*b.Y - a.Y*b.X,
	}
}
