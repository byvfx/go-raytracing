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
	AspectRatio     float64
	ImageWidth      int
	SamplesPerPixel int
	MaxDepth        int
	Vfov            float64
	LookFrom        Point3
	LookAt          Point3
	Vup             Vec3
	DefocusAngle    float64
	FocusDist       float64

	imageHeight        int
	pixelsSamplesScale float64
	center             Point3
	pixel00Loc         Point3
	pixelDeltaU        Vec3
	pixelDeltaV        Vec3
	u, v, w            Vec3
	defocusDiskU       Vec3
	defocusDiskV       Vec3
}

func NewCamera() *Camera {
	return &Camera{
		AspectRatio:     1.0,
		ImageWidth:      800,
		SamplesPerPixel: 10,
		MaxDepth:        50,
		Vfov:            90,
		LookFrom:        Point3{0, 0, 0},
		LookAt:          Point3{0, 0, -1},
		Vup:             Vec3{0, 1, 0},
	}
}

// init camera parameters
func (c *Camera) initialize() {
	c.imageHeight = max(int(float64(c.ImageWidth)/c.AspectRatio), 1)

	c.pixelsSamplesScale = 1.0 / float64(c.SamplesPerPixel)

	c.center = c.LookFrom

	theta := DegreesToRadians(c.Vfov)

	h := math.Tan(theta / 2)

	viewportHeight := 2 * h * c.FocusDist

	viewportWidth := viewportHeight * (float64(c.ImageWidth) / float64(c.imageHeight))

	c.w = c.LookFrom.Sub(c.LookAt).Unit()

	c.u = Cross(c.Vup, c.w).Unit()

	c.v = Cross(c.w, c.u)

	viewportU := c.u.Scale(viewportWidth)

	viewportV := c.v.Neg().Scale(viewportHeight)

	c.pixelDeltaU = viewportU.Div(float64(c.ImageWidth))

	c.pixelDeltaV = viewportV.Div(float64(c.imageHeight))

	viewportUpperLeft := c.center.
		Sub(c.w.Scale(c.FocusDist)).
		Sub(viewportU.Div(2)).
		Sub(viewportV.Div(2))

	c.pixel00Loc = viewportUpperLeft.Add(c.pixelDeltaU.Add(c.pixelDeltaV).Scale(0.5))

	defocusRadius := c.FocusDist * math.Tan(DegreesToRadians(c.DefocusAngle/2))
	c.defocusDiskU = c.u.Scale(defocusRadius)
	c.defocusDiskV = c.v.Scale(defocusRadius)
}

func (c *Camera) defocusDiskSample() Point3 {
	p := RandomInUnitDisk()
	return c.center.Add((c.defocusDiskU.Scale(p.X)).Add(c.defocusDiskV.Scale(p.Y)))
}

func (c *Camera) sampleSquare() Vec3 {
	return Vec3{
		X: RandomDouble() - 0.5,
		Y: RandomDouble() - 0.5,
		Z: 0,
	}
}

func (c *Camera) getRay(i, j int) Ray {
	offset := c.sampleSquare()
	pixelSample := c.pixel00Loc.
		Add(c.pixelDeltaU.Scale(float64(i) + offset.X)).
		Add(c.pixelDeltaV.Scale(float64(j) + offset.Y))

	var rayOrigin Point3
	if c.DefocusAngle <= 0 {
		rayOrigin = c.center

	} else {
		rayOrigin = c.defocusDiskSample()
	}

	rayDirection := pixelSample.Sub(rayOrigin)
	return NewRay(rayOrigin, rayDirection)
}

// sending out them color rays
func (c *Camera) rayColor(r Ray, depth int, world Hittable) Color {
	if depth <= 0 {
		return Color{X: 0, Y: 0, Z: 0}
	}

	rec := &HitRecord{}

	if world.Hit(r, NewInterval(0.001, math.Inf(1)), rec) {
		var attenuation Color
		var scattered Ray

		if rec.Mat.Scatter(r, rec, &attenuation, &scattered) {
			return attenuation.Mult(c.rayColor(scattered, depth-1, world))
		}
		return Color{X: 0, Y: 0, Z: 0}
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

	const barWidth = 40

	for j := range c.imageHeight {
		c.progressBar(j+1, c.imageHeight, barWidth)
		for i := range c.ImageWidth {
			pixelColor := Color{X: 0, Y: 0, Z: 0}
			for sample := 0; sample < c.SamplesPerPixel; sample++ {
				r := c.getRay(i, j)
				pixelColor = pixelColor.Add(c.rayColor(r, c.MaxDepth, world))
			}

			rgb_r, rgb_g, rgb_b := pixelColor.ToRGB(c.SamplesPerPixel)

			// PNG output
			img.Set(i, j, color.RGBA{R: uint8(rgb_r), G: uint8(rgb_g), B: uint8(rgb_b), A: 255})

		}
	}

	fmt.Fprintln(os.Stderr)

	//Write PNG file
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
