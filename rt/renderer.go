package rt

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/basicfont"
)

type ProgressiveRenderer struct {
	framebuffer *image.RGBA
	camera      *Camera
	world       Hittable
	currentRow  int
	completed   bool
	renderStart time.Time
	renderEnd   time.Time
}

func NewProgressiveRenderer(camera *Camera, world Hittable) *ProgressiveRenderer {
	framebuffer := image.NewRGBA(image.Rect(0, 0, camera.ImageWidth, camera.ImageHeight))
	return &ProgressiveRenderer{
		framebuffer: framebuffer,
		camera:      camera,
		world:       world,
		currentRow:  0,
		completed:   false,
		renderStart: time.Now(), // Start timing when renderer is created
	}
}
func (r *ProgressiveRenderer) Update() error {
	if r.currentRow < r.camera.ImageHeight {
		r.renderScanline(r.currentRow)
		r.currentRow++
		if r.currentRow >= r.camera.ImageHeight && !r.completed {
			r.completed = true
			r.renderEnd = time.Now()
			r.drawStatsToFramebuffer()
			_ = r.SaveImage("image.png")

			// Print render stats with actual render time
			renderDuration := r.renderEnd.Sub(r.renderStart)
			PrintRenderStats(renderDuration, r.camera.ImageWidth, r.camera.ImageHeight)
		}
	}
	return nil
}

func (r *ProgressiveRenderer) Draw(screen *ebiten.Image) {
	screen.WritePixels(r.framebuffer.Pix)

	// Draw render settings in lower left corner
	r.drawRenderSettings(screen)
}

func (r *ProgressiveRenderer) drawRenderSettings(screen *ebiten.Image) {
	// Calculate progress
	progress := float64(r.currentRow) / float64(r.camera.ImageHeight) * 100.0
	if r.completed {
		progress = 100.0
	}

	// Calculate elapsed time
	var elapsed time.Duration
	if r.completed {
		elapsed = r.renderEnd.Sub(r.renderStart)
	} else {
		elapsed = time.Since(r.renderStart)
	}

	// Position text in lower left corner
	x := 10
	y := r.camera.ImageHeight - 95
	lineHeight := 12

	// Create semi-transparent background for better readability
	bgColor := color.RGBA{R: 0, G: 0, B: 0, A: 50}
	bgRect := image.Rect(x-5, y-5, x+150, y+85)
	for py := bgRect.Min.Y; py < bgRect.Max.Y; py++ {
		for px := bgRect.Min.X; px < bgRect.Max.X; px++ {
			if px >= 0 && px < r.camera.ImageWidth && py >= 0 && py < r.camera.ImageHeight {
				r.framebuffer.Set(px, py, bgColor)
			}
		}
	}
	// Display render settings
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Resolution: %dx%d", r.camera.ImageWidth, r.camera.ImageHeight), x, y)
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Samples/Pixel: %d", r.camera.SamplesPerPixel), x, y)
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Max Depth: %d", r.camera.MaxDepth), x, y)
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Progress: %.1f%%", progress), x, y)
	y += lineHeight
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Render Time: %s", FormatDuration(elapsed)), x, y)
	y += lineHeight

	if r.completed {
		ebitenutil.DebugPrintAt(screen, "Status: COMPLETED", x, y)
	} else {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Scanline: %d/%d", r.currentRow, r.camera.ImageHeight), x, y)
	}

}

// drawStatsToFramebuffer draws the render statistics directly onto the framebuffer for saving
func (r *ProgressiveRenderer) drawStatsToFramebuffer() {
	elapsed := r.renderEnd.Sub(r.renderStart)

	x := 10
	y := r.camera.ImageHeight - 95
	lineHeight := 12

	// Draw semi-transparent background
	bgColor := color.RGBA{R: 0, G: 0, B: 0, A: 50}
	bgRect := image.Rect(x-5, y-5, x+150, y+85)
	for py := bgRect.Min.Y; py < bgRect.Max.Y; py++ {
		for px := bgRect.Min.X; px < bgRect.Max.X; px++ {
			if px >= 0 && px < r.camera.ImageWidth && py >= 0 && py < r.camera.ImageHeight {
				// Blend background color with existing pixel
				existing := r.framebuffer.At(px, py)
				er, eg, eb, _ := existing.RGBA()
				// Simple alpha blending
				a := float64(bgColor.A) / 255.0
				finalR := uint8(float64(bgColor.R)*a + float64(er>>8)*(1-a))
				finalG := uint8(float64(bgColor.G)*a + float64(eg>>8)*(1-a))
				finalB := uint8(float64(bgColor.B)*a + float64(eb>>8)*(1-a))
				r.framebuffer.Set(px, py, color.RGBA{R: finalR, G: finalG, B: finalB, A: 255})
			}
		}
	}

	// Prepare text lines
	textColor := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	face := text.NewGoXFace(basicfont.Face7x13)

	lines := []string{
		fmt.Sprintf("Resolution: %dx%d", r.camera.ImageWidth, r.camera.ImageHeight),
		fmt.Sprintf("Samples/Pixel: %d", r.camera.SamplesPerPixel),
		fmt.Sprintf("Max Depth: %d", r.camera.MaxDepth),
		fmt.Sprintf("Progress: 100.0%%"),
		fmt.Sprintf("Render Time: %s", FormatDuration(elapsed)),
		"Status: COMPLETED",
	}

	// Create temporary ebiten.Image from framebuffer
	tempImg := ebiten.NewImageFromImage(r.framebuffer)

	fontHeight := 13
	for i, line := range lines {
		opts := &text.DrawOptions{}
		opts.GeoM.Translate(float64(x), float64(y+i*lineHeight+fontHeight-2))
		opts.ColorScale.ScaleWithColor(textColor)
		text.Draw(tempImg, line, face, opts)
	}

	// Copy the result back to framebuffer
	tempImg.ReadPixels(r.framebuffer.Pix)
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
			fmt.Fprintf(os.Stderr, "ERROR: Could not close file '%s': %v\n", filename, err)
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

// GetRenderDuration returns the actual render time
func (r *ProgressiveRenderer) GetRenderDuration() time.Duration {
	if r.completed {
		return r.renderEnd.Sub(r.renderStart)
	}
	return time.Since(r.renderStart)
}
