package main

import (
	"go-raytracing/rt"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	startTime := time.Now()

	// Create quads scene
	world := rt.QuadsScene()
	camera := rt.QuadsCamera()

	bvh := rt.NewBVHNodeFromList(world)

	rt.PrintRenderSettings(camera, len(world.Objects))

	// Create progressive renderer
	renderer := rt.NewProgressiveRenderer(camera, bvh)

	// Set window properties
	ebiten.SetWindowSize(camera.ImageWidth, camera.ImageHeight)
	ebiten.SetWindowTitle("Go Raytracer")

	// Run the game loop
	if err := ebiten.RunGame(renderer); err != nil {
		panic(err)
	}

	elapsed := time.Since(startTime)
	rt.PrintRenderStats(elapsed, camera.ImageWidth, camera.ImageHeight)
}
