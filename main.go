package main

import "go-raytracing/rt"

func main() {
	//material time
	materialGround := rt.NewLambertian(rt.Color{X: 0.8, Y: 0.8, Z: 0.0})
	materialCenter := rt.NewLambertian(rt.Color{X: 0.1, Y: 0.2, Z: 0.5})
	materialLeft := rt.NewMetal(rt.Color{X: 0.8, Y: 0.8, Z: 0.8}, 0.0)
	materialRight := rt.NewDielectric(1.25)

	// Make the camera
	camera := rt.NewCamera()
	camera.AspectRatio = 16.0 / 9.0
	camera.ImageWidth = 800
	camera.SamplesPerPixel = 100
	camera.MaxDepth = 50

	// Sphere time!
	world := rt.NewHittableList()
	world.Add(rt.NewSphere(rt.Point3{X: 0, Y: -100.5, Z: -1}, 100, materialGround))
	world.Add(rt.NewSphere(rt.Point3{X: 0, Y: 0, Z: -1}, 0.5, materialCenter))
	world.Add(rt.NewSphere(rt.Point3{X: -1, Y: 0, Z: -1}, 0.25, materialLeft)) // addtional sphere
	world.Add(rt.NewSphere(rt.Point3{X: 1, Y: 0, Z: -1}, 0.25, materialRight)) // addtional sphere

	// Render time
	camera.Render(world)
}
