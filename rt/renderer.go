package rt

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type ProgressiveRenderer struct {
	framebuffer *image.RGBA
	camera      *Camera
	world       Hittable
	currentRow  int
	completed   bool
}

func NewProgressiveRenderer(camera *Camera, world Hittable) *ProgressiveRenderer {
	framebuffer := image.NewRGBA(image.Rect(0, 0, camera.ImageWidth, camera.ImageHeight))
	return &ProgressiveRenderer{
		framebuffer: framebuffer,
		camera:      camera,
		world:       world,
		currentRow:  0,
		completed:   false,
	}
}
func (r *ProgressiveRenderer) Update() error {
	if r.currentRow < r.camera.ImageHeight {
		r.renderScanline(r.currentRow)
		r.currentRow++
		if r.currentRow >= r.camera.ImageHeight && !r.completed {
			r.completed = true
			_ = r.SaveImage("image.png")
		}
	}
	return nil
}

func (r *ProgressiveRenderer) Draw(screen *ebiten.Image) {
	screen.WritePixels(r.framebuffer.Pix)
}

func (r *ProgressiveRenderer) Layout(w, h int) (int, int) {
	return r.camera.ImageWidth, r.camera.ImageHeight
}
func (r *ProgressiveRenderer) renderScanline(j int) {
	for i := 0; i < r.camera.ImageWidth; i++ {
		pixelColor := Color{X: 0, Y: 0, Z: 0}

		for sample := 0; sample < r.camera.SamplesPerPixel; sample++ {
			ray := r.camera.GetRay(i, j)
			pixelColor = pixelColor.Add(r.camera.RayColor(ray, r.camera.MaxDepth, r.world))
		}

		scale := 1.0 / float64(r.camera.SamplesPerPixel)
		pixelColor = pixelColor.Scale(scale)

		r.framebuffer.Set(i, j, color.RGBA{
			R: uint8(256 * Clamp(LinearToGamma(pixelColor.X), 0, 0.999)),
			G: uint8(256 * Clamp(LinearToGamma(pixelColor.Y), 0, 0.999)),
			B: uint8(256 * Clamp(LinearToGamma(pixelColor.Z), 0, 0.999)),
			A: 255,
		})
	}
}

func (r *ProgressiveRenderer) SaveImage(filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating image file: %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	if err := png.Encode(file, r.framebuffer); err != nil {
		return fmt.Errorf("error encoding PNG: %w", err)
	}

	fmt.Printf("\n✓ Image saved to %s\n", filename)
	return nil
}

func (r *ProgressiveRenderer) IsCompleted() bool {
	return r.completed
}
