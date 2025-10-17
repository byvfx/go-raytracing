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

	imageHeight        int
	pixelsSamplesScale float64
	center             Point3
	pixel00Loc         Point3
	pixelDeltaU        Vec3
	pixelDeltaV        Vec3
	u, v, w            Vec3
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

	// Camera center is at LookFrom
	c.center = c.LookFrom

	// ✓ Calculate focal length FIRST
	// C++: auto focal_length = (lookfrom - lookat).length();
	focalLength := c.LookFrom.Sub(c.LookAt).Len()

	// Calculate viewport dimensions
	// C++: auto theta = degrees_to_radians(vfov);
	theta := DegreesToRadians(c.Vfov)

	// C++: auto h = std::tan(theta/2);
	h := math.Tan(theta / 2)

	// C++: auto viewport_height = 2 * h * focal_length;
	viewportHeight := 2 * h * focalLength

	// C++: auto viewport_width = viewport_height * (double(image_width)/image_height);
	viewportWidth := viewportHeight * (float64(c.ImageWidth) / float64(c.imageHeight))

	// ✓ Calculate camera basis vectors AFTER focal length
	// C++: w = unit_vector(lookfrom - lookat);
	c.w = c.LookFrom.Sub(c.LookAt).Unit()

	// C++: u = unit_vector(cross(vup, w));
	c.u = Cross(c.Vup, c.w).Unit()

	// C++: v = cross(w, u);
	c.v = Cross(c.w, c.u)

	// Calculate the vectors across the horizontal and down the vertical viewport edges
	// C++: vec3 viewport_u = viewport_width * u;
	viewportU := c.u.Scale(viewportWidth)

	// C++: vec3 viewport_v = viewport_height * -v;
	viewportV := c.v.Neg().Scale(viewportHeight)

	// Calculate the horizontal and vertical delta vectors from pixel to pixel
	// C++: pixel_delta_u = viewport_u / image_width;
	c.pixelDeltaU = viewportU.Div(float64(c.ImageWidth))

	// C++: pixel_delta_v = viewport_v / image_height;
	c.pixelDeltaV = viewportV.Div(float64(c.imageHeight))

	// Calculate the location of the upper left pixel
	// C++: auto viewport_upper_left = center - (focal_length * w) - viewport_u/2 - viewport_v/2;
	viewportUpperLeft := c.center.
		Sub(c.w.Scale(focalLength)).
		Sub(viewportU.Div(2)).
		Sub(viewportV.Div(2))

	// C++: pixel00_loc = viewport_upper_left + 0.5 * (pixel_delta_u + pixel_delta_v);
	c.pixel00Loc = viewportUpperLeft.Add(c.pixelDeltaU.Add(c.pixelDeltaV).Scale(0.5))
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

	rayOrigin := c.center

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
			pixelColor := Color{X: 0, Y: 0, Z: 0}
			for sample := 0; sample < c.SamplesPerPixel; sample++ {
				r := c.getRay(i, j)
				pixelColor = pixelColor.Add(c.rayColor(r, c.MaxDepth, world))
			}

			rgb_r, rgb_g, rgb_b := pixelColor.ToRGB(c.SamplesPerPixel)

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
