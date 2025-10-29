package main

import (
	"fmt"
	"go-raytracing/rt"
	"image"
	"image/color"
	"image/png"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type RaytracerApp struct {
	framebuffer *image.RGBA
	camera      *rt.Camera
	world       rt.Hittable
	currentRow  int
	completed   bool
}

func (app *RaytracerApp) Update() error {
	// Render one scanline per frame for progressive display
	if app.currentRow < app.camera.ImageHeight {
		app.renderScanline(app.currentRow)
		app.currentRow++

		// Check if rendering is complete
		if app.currentRow >= app.camera.ImageHeight && !app.completed {
			app.completed = true
			app.saveImage()
		}
	}
	return nil
}

func (app *RaytracerApp) Draw(screen *ebiten.Image) {
	screen.WritePixels(app.framebuffer.Pix)
}

func (app *RaytracerApp) Layout(w, h int) (int, int) {
	return app.camera.ImageWidth, app.camera.ImageHeight
}

func (app *RaytracerApp) renderScanline(j int) {
	for i := 0; i < app.camera.ImageWidth; i++ {
		pixelColor := rt.Color{X: 0, Y: 0, Z: 0}

		for sample := 0; sample < app.camera.SamplesPerPixel; sample++ {
			r := app.camera.GetRay(i, j)
			pixelColor = pixelColor.Add(app.camera.RayColor(r, app.camera.MaxDepth, app.world))
		}

		scale := 1.0 / float64(app.camera.SamplesPerPixel)
		pixelColor = pixelColor.Scale(scale)

		app.framebuffer.Set(i, j, color.RGBA{
			R: uint8(256 * rt.Clamp(rt.LinearToGamma(pixelColor.X), 0, 0.999)),
			G: uint8(256 * rt.Clamp(rt.LinearToGamma(pixelColor.Y), 0, 0.999)),
			B: uint8(256 * rt.Clamp(rt.LinearToGamma(pixelColor.Z), 0, 0.999)),
			A: 255,
		})
	}
}

func (app *RaytracerApp) saveImage() {
	file, err := os.Create("image.png")
	if err != nil {
		fmt.Printf("Error creating image file: %v\n", err)
		return
	}
	defer file.Close()

	if err := png.Encode(file, app.framebuffer); err != nil {
		fmt.Printf("Error encoding PNG: %v\n", err)
		return
	}

	fmt.Println("\n✓ Image saved to image.png")
}

func main() {
	startTime := time.Now()

	// Create scene
	world := rt.RandomScene()
	bvh := rt.NewBVHNodeFromList(world)

	// Configure camera
	camera := rt.NewCamera()
	camera.ApplyPreset(rt.StandardQuality())
	camera.CameraMotion = false
	camera.LookFrom = rt.Point3{X: 12, Y: 2, Z: 3}
	camera.LookAt = rt.Point3{X: 0, Y: 0, Z: 0}
	camera.Initialize()

	rt.PrintRenderSettings(camera, len(world.Objects))

	// Create framebuffer
	framebuffer := image.NewRGBA(image.Rect(0, 0, camera.ImageWidth, camera.ImageHeight))

	// Create Ebiten app
	app := &RaytracerApp{
		framebuffer: framebuffer,
		camera:      camera,
		world:       bvh,
		currentRow:  0,
		completed:   false,
	}

	// Set window properties
	ebiten.SetWindowSize(camera.ImageWidth, camera.ImageHeight)
	ebiten.SetWindowTitle("Go Raytracer - Progressive Scanline Rendering")

	// Run the game loop
	if err := ebiten.RunGame(app); err != nil {
		panic(err)
	}

	elapsed := time.Since(startTime)
	rt.PrintRenderStats(elapsed, camera.ImageWidth, camera.ImageHeight)
}
