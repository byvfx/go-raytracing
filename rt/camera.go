package rt

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strings"
)

type Camera struct {
	// caps mean public
	AspectRatio float64
	ImageWidth  int

	imageHeight int
	center      Point3
	pixel00Loc  Point3
	pixelDeltaU Vec3
	pixelDeltaV Vec3
}

func NewCamera() *Camera {
	return &Camera{
		AspectRatio: 1.0,
		ImageWidth:  800,
	}
}

// init camera parameters
func (c *Camera) initialize() {

	c.imageHeight = max(int(float64(c.ImageWidth)/c.AspectRatio), 1)
	c.center = Point3{X: 0, Y: 0, Z: 0}
	focalLength := 1.0
	viewportHeight := 2.0
	viewportWidth := viewportHeight * (float64(c.ImageWidth) / float64(c.imageHeight))

	viewportU := Vec3{X: viewportWidth, Y: 0, Z: 0}
	viewportV := Vec3{X: 0, Y: -viewportHeight, Z: 0}

	c.pixelDeltaU = viewportU.Div(float64(c.ImageWidth))
	c.pixelDeltaV = viewportV.Div(float64(c.imageHeight))

	viewportUpperLeft := c.center.
		Sub(Vec3{X: 0, Y: 0, Z: focalLength}).
		Sub(viewportU.Div(2)).
		Sub(viewportV.Div(2))
	c.pixel00Loc = viewportUpperLeft.Add(c.pixelDeltaU.Add(c.pixelDeltaV).Scale(0.5))
}

// sending out them rays
func (c *Camera) rayColor(r Ray, world Hittable) Color {
	rec := &HitRecord{}

	if world.Hit(r, 0, math.Inf(1), rec) {
		return Color{X: rec.Normal.X + 1, Y: rec.Normal.Y + 1, Z: rec.Normal.Z + 1}.Scale(0.5)
	}

	unitDirection := r.Direction().Unit()
	a := 0.5 * (unitDirection.Y + 1)

	white := Color{X: 1.0, Y: 1.0, Z: 1.0}
	blue := Color{X: 0.1, Y: 0.3, Z: 1.0}
	return white.Scale(1.0 - a).Add(blue.Scale(a))
}

func (c *Camera) progressBar(done, total, width int) {
	p := float64(done) / float64(total)
	filled := min(int(p*float64(width)+0.5), width)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	// happy  little accident enable to see each progress step
	//fmt.Fprintln(os.Stderr)
	//
	fmt.Fprintf(os.Stderr, "\r[%s] %3.0f%%  scanlines remaining: %d", bar, p*100, total-done)

}

func (c *Camera) Render(world Hittable) {
	c.initialize()

	img := image.NewRGBA(image.Rect(0, 0, c.ImageWidth, c.imageHeight))

	// Render (PPM - commented out)
	// out, err := os.Create("image.ppm")
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
	// 	return
	// }
	// w := bufio.NewWriter(out)
	// defer w.Flush()
	// fmt.Fprintf(w, "P3\n%d %d\n255\n", imageWidth, imageHeight)

	const barWidth = 40

	for j := range c.imageHeight {
		c.progressBar(j+1, c.imageHeight, barWidth)
		for i := range c.ImageWidth {
			pixelCenter := c.pixel00Loc.
				Add(c.pixelDeltaU.Scale(float64(i))).
				Add(c.pixelDeltaV.Scale(float64(j)))

			rayDirection := pixelCenter.Sub(c.center)
			r := NewRay(c.center, rayDirection)

			pixelColor := c.rayColor(r, world)
			rgb_r, rgb_g, rgb_b := pixelColor.ToRGB(1)

			// PNG output
			img.Set(i, j, color.RGBA{R: uint8(rgb_r), G: uint8(rgb_g), B: uint8(rgb_b), A: 255})

			// PPM output
			// fmt.Fprintf(w, "%d %d %d\n", rgb_r, rgb_g, rgb_b)
		}
	}

	fmt.Fprintln(os.Stderr)

	// Write PNG file
	outFile, err := os.Create("image.png")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating PNG file: %v\n", err)
		return
	}
	defer outFile.Close()

	err = png.Encode(outFile, img)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding PNG: %v\n", err)
		return
	}

	fmt.Fprintln(os.Stdout, "Done. Image written to image.png")
}
