package main

import (
	"go-raytracing/rt"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	world := rt.PrimitivesScene()
	camera := rt.PrimitivesCamera()

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
}
