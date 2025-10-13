package main

import (
	"fmt"
	"go-raytracing/rt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strings"
)

func progressBar(done, total, width int) {
	p := float64(done) / float64(total)

	// updated this with the min function instead of a if else
	filled := min(int(p*float64(width)+0.5), width)
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	// happy  little accident enable to see each progress step
	//fmt.Fprintln(os.Stderr)
	//
	fmt.Fprintf(os.Stderr, "\r[%s] %3.0f%%  scanlines remaining: %d", bar, p*100, total-done)
}

func rayColor(r rt.Ray, world rt.Hittable) rt.Color {
	rec := &rt.HitRecord{}

	// Check if ray hits any object in the world
	if world.Hit(r, 0, math.Inf(1), rec) {
		// Color based on the surface normal
		return rt.Color{X: rec.Normal.X + 1, Y: rec.Normal.Y + 1, Z: rec.Normal.Z + 1}.Scale(0.5)
	}

	// Otherwise, render the sky gradient
	unitDirection := r.Direction().Unit()
	a := 0.5 * (unitDirection.Y + 1.0)

	// Use the original normalized color values
	white := rt.Color{X: 1.0, Y: 1.0, Z: 1.0}
	blue := rt.Color{X: 0.5, Y: 0.7, Z: 1.0}
	return white.Scale(1.0 - a).Add(blue.Scale(a))
}

func main() {
	// Image
	aspectRatio := 16.0 / 9.0
	imageWidth := 800

	// Calculate the image height, and ensure that it's at least 1.
	imageHeight := max(int(float64(imageWidth)/aspectRatio), 1)

	// Camera
	focalLength := 1.0
	viewportHeight := 2.0
	viewportWidth := viewportHeight * (float64(imageWidth) / float64(imageHeight))
	cameraCenter := rt.Point3{X: 0, Y: 0, Z: 0}

	// Calculate the vectors across the horizontal and down the vertical viewport edges.
	viewportU := rt.Vec3{X: viewportWidth, Y: 0, Z: 0}
	viewportV := rt.Vec3{X: 0, Y: -viewportHeight, Z: 0}

	// Calculate the horizontal and vertical delta vectors from pixel to pixel.
	pixelDeltaU := viewportU.Div(float64(imageWidth))
	pixelDeltaV := viewportV.Div(float64(imageHeight))

	// Calculate the location of the upper left pixel.
	viewportUpperLeft := cameraCenter.
		Sub(rt.Vec3{X: 0, Y: 0, Z: focalLength}).
		Sub(viewportU.Div(2)).
		Sub(viewportV.Div(2))
	pixel00Loc := viewportUpperLeft.Add(pixelDeltaU.Add(pixelDeltaV).Scale(0.5))

	// World - create objects to render
	world := rt.NewHittableList()
	world.Add(rt.NewSphere(rt.Point3{X: 0, Y: 0, Z: -1}, 0.5))
	world.Add(rt.NewSphere(rt.Point3{X: 0, Y: -100.5, Z: -1}, 100))

	// Create image for PNG output
	img := image.NewRGBA(image.Rect(0, 0, imageWidth, imageHeight))

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

	for j := range imageHeight {
		progressBar(j+1, imageHeight, barWidth)
		for i := range imageWidth {
			pixelCenter := pixel00Loc.
				Add(pixelDeltaU.Scale(float64(i))).
				Add(pixelDeltaV.Scale(float64(j)))
			rayDirection := pixelCenter.Sub(cameraCenter)
			r := rt.NewRay(cameraCenter, rayDirection)

			pixelColor := rayColor(r, world)
			rgb_r, rgb_g, rgb_b := pixelColor.ToRGB(1)

			// PNG output
			img.Set(i, j, color.RGBA{R: uint8(rgb_r), G: uint8(rgb_g), B: uint8(rgb_b), A: 255})

			// PPM output (commented out)
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
