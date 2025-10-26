package main

import (
	"go-raytracing/rt"
	"time"
)

func main() {
	startTime := time.Now()
	config := rt.DefaultSceneConfig()
	config.LambertProb = 1.0
	config.MetalProb = 0.1
	config.DielectricProb = 1.0
	world := rt.CheckeredSpheresScene()
	bvh := rt.NewBVHNodeFromList(world)

	//material time
	// materialGround := rt.NewLambertian(rt.Color{X: 0.5, Y: 0.5, Z: 0.0})
	// materialCenter := rt.NewLambertian(rt.Color{X: 1.0, Y: 0.5, Z: 0.0})
	// materialLeft := rt.NewDielectric(1.5)
	// materialBubble := rt.NewDielectric(1.0 / 1.5)
	// materialRight := rt.NewMetal(rt.Color{X: 1.0, Y: 1.0, Z: 1.0}, 0.0)

	// Make the camera
	camera := rt.NewCamera()
	camera.ApplyPreset(rt.QuickPreview())
	camera.CameraMotion = false
	camera.LookFrom2 = rt.Point3{X: 12, Y: 2, Z: 3}
	camera.LookAt2 = rt.Point3{X: 0, Y: 0, Z: 0}
	camera.Initialize()

	// camera.AspectRatio = 16.0 / 9.0
	// camera.ImageWidth = 600
	// camera.SamplesPerPixel = 200
	// camera.MaxDepth = 50
	// camera.Vfov = 20
	// camera.DefocusAngle = 0.75
	// camera.FocusDist = 10.0
	// camera.CameraMotion = true

	//position camera
	// camera.LookFrom = rt.Point3{X: 13, Y: 2, Z: 3}
	// camera.LookFrom2 = rt.Point3{X: 12, Y: 2, Z: 2.5} // Move forward
	// camera.LookAt = rt.Point3{X: 0, Y: 0, Z: 0}
	// camera.LookAt2 = rt.Point3{X: 0, Y: 0, Z: 0}

	//camera.LookFrom = rt.Point3{X: -2, Y: 2, Z: 1}
	//camera.LookAt = rt.Point3{X: 0, Y: 0, Z: -1}
	//camera.Vup = rt.Vec3{X: 0, Y: 1, Z: 0}

	// Sphere time!

	// world := rt.NewHittableList()

	//world.Add(rt.NewSphere(rt.Point3{X: 0, Y: -100.5, Z: -1}, 100, materialGround))
	// world.Add(rt.NewPlane(rt.Point3{X: 0, Y: -0.5, Z: -1}, rt.Vec3{X: 0, Y: 1, Z: 0}, materialGround))
	// world.Add(rt.NewSphere(rt.Point3{X: 0, Y: 0, Z: -1}, 0.5, materialCenter))
	// world.Add(rt.NewSphere(rt.Point3{X: -1, Y: 0, Z: -1}, 0.5, materialLeft))
	// world.Add(rt.NewSphere(rt.Point3{X: -1, Y: 0, Z: -1}, 0.4, materialBubble))
	// world.Add(rt.NewSphere(rt.Point3{X: 1, Y: 0, Z: -1}, 0.5, materialRight))

	// Render time

	rt.PrintRenderSettings(camera, len(world.Objects))

	camera.Render(bvh)
	elapsed := time.Since(startTime)

	rt.PrintRenderStats(elapsed, camera.ImageWidth, camera.ImageHeight)
}
