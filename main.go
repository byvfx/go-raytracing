package main

import (
	"go-raytracing/rt"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	world := rt.CornellBoxScene()
	camera := rt.CornellBoxCamera()

	bvh := rt.NewBVHNodeFromList(world)

	rt.PrintRenderSettings(camera, len(world.Objects))

	renderer := rt.NewProgressiveRenderer(camera, bvh)

	ebiten.SetWindowSize(camera.ImageWidth, camera.ImageHeight)
	ebiten.SetWindowTitle("Go Raytracer")

	if err := ebiten.RunGame(renderer); err != nil {
		panic(err)
	}
}
