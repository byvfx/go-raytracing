package main

import "go-raytracing/rt"

func main() {
	//material time
	materialGround := rt.NewLambertian(rt.Color{X: 0.5, Y: 0.5, Z: 0.0})
	materialCenter := rt.NewLambertian(rt.Color{X: 1.0, Y: 0.5, Z: 0.0})
	materialLeft := rt.NewDielectric(1.5)
	materialBubble := rt.NewDielectric(1.0 / 1.5)
	materialRight := rt.NewMetal(rt.Color{X: 1.0, Y: 1.0, Z: 1.0}, 0.0)

	// Make the camera
	camera := rt.NewCamera()
	camera.AspectRatio = 16.0 / 9.0
	camera.ImageWidth = 400
	camera.SamplesPerPixel = 100
	camera.MaxDepth = 10
	camera.Vfov = 35
	camera.DefocusAngle = 10.0
	camera.FocusDist = 3.4

	//position camera

	camera.LookFrom = rt.Point3{X: -2, Y: 2, Z: 1}
	camera.LookAt = rt.Point3{X: 0, Y: 0, Z: -1}
	camera.Vup = rt.Vec3{X: 0, Y: 1, Z: 0}

	// Sphere time!
	world := rt.NewHittableList()

	//world.Add(rt.NewSphere(rt.Point3{X: 0, Y: -100.5, Z: -1}, 100, materialGround))
	world.Add(rt.NewPlane(rt.Point3{X: 0, Y: -0.5, Z: -1}, rt.Vec3{X: 0, Y: 1, Z: 0}, materialGround))
	world.Add(rt.NewSphere(rt.Point3{X: 0, Y: 0, Z: -1}, 0.5, materialCenter))
	world.Add(rt.NewSphere(rt.Point3{X: -1, Y: 0, Z: -1}, 0.5, materialLeft))
	world.Add(rt.NewSphere(rt.Point3{X: -1, Y: 0, Z: -1}, 0.4, materialBubble))
	world.Add(rt.NewSphere(rt.Point3{X: 1, Y: 0, Z: -1}, 0.5, materialRight))

	// Render time
	camera.Render(world)
}
