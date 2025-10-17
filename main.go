package main

import "go-raytracing/rt"

func main() {
	// Make the camera
	camera := rt.NewCamera()
	camera.AspectRatio = 16.0 / 9.0
	camera.ImageWidth = 800
	camera.SamplesPerPixel = 100

	// Sphere time!
	world := rt.NewHittableList()
	world.Add(rt.NewSphere(rt.Point3{X: 0, Y: 0, Z: -1}, 0.5))
	world.Add(rt.NewSphere(rt.Point3{X: -1, Y: 0, Z: -1.5}, .5)) // addtional sphere
	world.Add(rt.NewSphere(rt.Point3{X: 1, Y: 0, Z: -1.5}, .5))  // addtional sphere
	world.Add(rt.NewSphere(rt.Point3{X: 0, Y: -100.5, Z: -1}, 100))

	// Render time
	camera.Render(world)
}
