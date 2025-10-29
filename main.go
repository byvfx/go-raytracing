package main

import (
	"go-raytracing/rt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	startTime := time.Now()

	// Create scene
	world := rt.RandomScene()
	bvh := rt.NewBVHNodeFromList(world)

	// Configure camera
	camera := rt.NewCamera()
	camera.ApplyPreset(rt.QuickPreview())
	camera.ImageWidth = 1000
	camera.CameraMotion = false
	camera.LookFrom = rt.Point3{X: 12, Y: 2, Z: 3}
	camera.LookAt = rt.Point3{X: 0, Y: 0, Z: 0}
	camera.Initialize()

	rt.PrintRenderSettings(camera, len(world.Objects))

	// Create progressive renderer
	renderer := rt.NewProgressiveRenderer(camera, bvh)

	// Set window properties
	ebiten.SetWindowSize(camera.ImageWidth, camera.ImageHeight)
	ebiten.SetWindowTitle("Go Raytracer - Progressive Scanline Rendering")

	// Run the game loop
	if err := ebiten.RunGame(renderer); err != nil {
		panic(err)
	}

	elapsed := time.Since(startTime)
	rt.PrintRenderStats(elapsed, camera.ImageWidth, camera.ImageHeight)
}
