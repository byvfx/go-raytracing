package main

import (
	"bufio"
	"fmt"
	"go-raytracing/rt"
	"os"
	"strings"
)

func progressBar(done, total, width int) {
	p := float64(done) / float64(total)
	filled := int(p*float64(width) + 0.5)
	if filled > width {
		filled = width
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	// happy  little accident enable to see each progress step
	//fmt.Fprintln(os.Stderr)
	//
	fmt.Fprintf(os.Stderr, "\r[%s] %3.0f%%  scanlines remaining: %d", bar, p*100, total-done)
}

func main() {

	imageWidth := 256
	imageHeight := 256
	const barWidth = 40

	out, err := os.Create("image.ppm")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
	}
	w := bufio.NewWriter(out)
	defer w.Flush()

	fmt.Fprintf(w, "P3\n%d %d\n255\n", imageWidth, imageHeight)

	// image creation loop
	for j := range imageHeight {
		for i := range imageWidth {
			color := rt.Color{
				X: float64(i) / float64(imageWidth-1),
				Y: float64(j) / float64(imageHeight-1),
				Z: 0.25,
			}
			r, g, b := color.ToRGB(1)
			fmt.Fprintf(w, "%d %d %d\n", r, g, b)
		}
		progressBar(j+1, imageHeight, barWidth)
	}
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stdout, "image written to disk")
}
