package main

import (
	"fmt"
	"go-raytracing/rt"
	"image"
	"image/color"
	"image/png"
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

func hitSphere(center rt.Point3, radius float64, r rt.Ray) bool {
	oc := center.Sub(r.Origin())
	a := rt.Dot(r.Direction(), r.Direction())
	b := -2.0 * rt.Dot(r.Direction(), oc)
	c := rt.Dot(oc, oc) - radius*radius
	discriminant := b*b - 4*a*c
	return discriminant >= 0
}

func rayColor(r rt.Ray) rt.Color {
	// Check if ray hits the sphere at (0, 0, -1) with radius 0.5
	if hitSphere(rt.Point3{X: 0, Y: 0, Z: -1}, 0.5, r) {
		return rt.Color{X: 1, Y: 0, Z: 1}
	}

	// Otherwise, render the sky gradient
	unitDirection := r.Direction().Unit()
	a := 0.5 * (unitDirection.Y + 1.0)
	// Blend white and blue based on Y coordinate
	// Using normalized RGB values (divide by 255)
	white := rt.Color{X: 255.0 / 255.0, Y: 255.0 / 255.0, Z: 255.0 / 255.0} // RGB(255, 255, 255)
	blue := rt.Color{X: 45.0 / 255.0, Y: 147.0 / 255.0, Z: 255.0 / 255.0}   // RGB(135, 206, 235) - sky blue
	return white.Scale(1.0 - a).Add(blue.Scale(a))

	// Original normalized values (commented out):
	// white := rt.Color{X: 1.0, Y: 1.0, Z: 1.0}
	// blue := rt.Color{X: 0.5, Y: 0.7, Z: 1.0}
	// return white.Scale(1.0 - a).Add(blue.Scale(a))
}

func main() {
	// Image
	aspectRatio := 16.0 / 9.0
	imageWidth := 400

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

	for j := 0; j < imageHeight; j++ {
		progressBar(j+1, imageHeight, barWidth)
		for i := range imageWidth {
			pixelCenter := pixel00Loc.
				Add(pixelDeltaU.Scale(float64(i))).
				Add(pixelDeltaV.Scale(float64(j)))
			rayDirection := pixelCenter.Sub(cameraCenter)
			r := rt.NewRay(cameraCenter, rayDirection)

			pixelColor := rayColor(r)
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
