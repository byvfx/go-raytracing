package main

import (
	"go-raytracing/rt"
	"runtime"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	world, camera := rt.CornellBoxLucy()

	bvh := rt.NewBVHNodeFromList(world)

	rt.PrintRenderSettings(camera, len(world.Objects))

	bucketSize := 32
	numWorkers := runtime.NumCPU()

	renderer := rt.NewBucketRenderer(camera, bvh, bucketSize, numWorkers)

	// renderer := rt.NewProgressiveRenderer(camera, bvh)

	ebiten.SetWindowSize(camera.ImageWidth, camera.ImageHeight)
	ebiten.SetWindowTitle("Go Raytracer")

	if err := ebiten.RunGame(renderer); err != nil {
		panic(err)
	}
}
