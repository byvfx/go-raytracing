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

// =============================================================================
// CAMERA STRUCT
// =============================================================================
type Camera struct {
	// caps mean public
	AspectRatio     float64
	ImageWidth      int
	ImageHeight     int
	SamplesPerPixel int
	MaxDepth        int
	Vfov            float64
	LookFrom        Point3
	LookAt          Point3
	Vup             Vec3
	DefocusAngle    float64
	FocusDist       float64
	LookFrom2       Point3
	LookAt2         Point3
	CameraMotion    bool

	pixelsSamplesScale float64
	center             Point3
	pixel00Loc         Point3
	pixelDeltaU        Vec3
	pixelDeltaV        Vec3
	u, v, w            Vec3
	defocusDiskU       Vec3
	defocusDiskV       Vec3
	centerMotion       Ray
	lookAtMotion       Ray
}

// =============================================================================
// CONSTRUCTOR
// =============================================================================

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
		DefocusAngle:    0.0,
		FocusDist:       1.0,
		LookFrom2:       Point3{0, 0, 0},
		LookAt2:         Point3{0, 0, 0},
		CameraMotion:    false,
	}
}

// =============================================================================
// CAMERA PRESETS
// =============================================================================

type CameraPreset struct {
	AspectRatio     float64
	ImageWidth      int
	SamplesPerPixel int
	MaxDepth        int
	Vfov            float64
	DefocusAngle    float64
	FocusDist       float64
	LookFrom        Point3
	LookAt          Point3
	Vup             Vec3
}

// camera presets
func QuickPreview() CameraPreset {
	return CameraPreset{
		AspectRatio:     16.0 / 9.0,
		ImageWidth:      400,
		SamplesPerPixel: 10,
		MaxDepth:        10,
		Vfov:            20,
		DefocusAngle:    0.0,
		FocusDist:       10.0,
		LookFrom:        Point3{X: 13, Y: 2, Z: 3},
		LookAt:          Point3{X: 0, Y: 0, Z: 0},
		Vup:             Vec3{X: 0, Y: 1, Z: 0},
	}
}

func StandardQuality() CameraPreset {
	return CameraPreset{
		AspectRatio:     16.0 / 9.0,
		ImageWidth:      600,
		SamplesPerPixel: 100,
		MaxDepth:        50,
		Vfov:            20,
		DefocusAngle:    0.6,
		FocusDist:       10.0,
		LookFrom:        Point3{X: 13, Y: 2, Z: 3},
		LookAt:          Point3{X: 0, Y: 0, Z: 0},
		Vup:             Vec3{X: 0, Y: 1, Z: 0},
	}
}

func HighQuality() CameraPreset {
	return CameraPreset{
		AspectRatio:     16.0 / 9.0,
		ImageWidth:      1200,
		SamplesPerPixel: 500,
		MaxDepth:        50,
		Vfov:            20,
		DefocusAngle:    0.6,
		FocusDist:       10.0,
		LookFrom:        Point3{X: 13, Y: 2, Z: 3},
		LookAt:          Point3{X: 0, Y: 0, Z: 0},
		Vup:             Vec3{X: 0, Y: 1, Z: 0},
	}
}
func (c *Camera) ApplyPreset(preset CameraPreset) {
	c.AspectRatio = preset.AspectRatio
	c.ImageWidth = preset.ImageWidth
	c.SamplesPerPixel = preset.SamplesPerPixel
	c.MaxDepth = preset.MaxDepth
	c.Vfov = preset.Vfov
	c.DefocusAngle = preset.DefocusAngle
	c.FocusDist = preset.FocusDist
	c.LookFrom = preset.LookFrom
	c.LookAt = preset.LookAt
	c.Vup = preset.Vup
}

// =============================================================================
// BUILDER PATTERN METHODS
// =============================================================================

func NewCameraBuilder() *Camera {
	return NewCamera()
}

func (c *Camera) SetResolution(width int, aspectRatio float64) *Camera {
	c.ImageWidth = width
	c.AspectRatio = aspectRatio
	return c
}

func (c *Camera) SetQuality(samples, maxDepth int) *Camera {
	c.SamplesPerPixel = samples
	c.MaxDepth = maxDepth
	return c
}

func (c *Camera) SetPosition(lookFrom, lookAt Point3, vup Vec3) *Camera {
	c.LookFrom = lookFrom
	c.LookAt = lookAt
	c.Vup = vup
	return c
}

func (c *Camera) SetLens(vfov, defocusAngle, focusDist float64) *Camera {
	c.Vfov = vfov
	c.DefocusAngle = defocusAngle
	c.FocusDist = focusDist
	return c
}

// =============================================================================
// INITIALIZATION
// =============================================================================

func (c *Camera) Initialize() {

	if c.CameraMotion {
		velocity := c.LookFrom2.Sub(c.LookFrom)
		c.centerMotion = NewRay(c.LookFrom, velocity, 0)

		lookAtVelocity := c.LookAt2.Sub(c.LookAt)
		c.lookAtMotion = NewRay(c.LookAt, lookAtVelocity, 0)

	} else {
		c.centerMotion = NewRay(c.LookFrom, Vec3{X: 0, Y: 0, Z: 0}, 0)
		c.lookAtMotion = NewRay(c.LookAt, Vec3{X: 0, Y: 0, Z: 0}, 0)
	}
	c.ImageHeight = max(int(float64(c.ImageWidth)/c.AspectRatio), 1)

	c.pixelsSamplesScale = 1.0 / float64(c.SamplesPerPixel)

	c.center = c.LookFrom

	theta := DegreesToRadians(c.Vfov)

	h := math.Tan(theta / 2)

	viewportHeight := 2 * h * c.FocusDist

	viewportWidth := viewportHeight * (float64(c.ImageWidth) / float64(c.ImageHeight))

	c.w = c.LookFrom.Sub(c.LookAt).Unit()

	c.u = Cross(c.Vup, c.w).Unit()

	c.v = Cross(c.w, c.u)

	viewportU := c.u.Scale(viewportWidth)

	viewportV := c.v.Neg().Scale(viewportHeight)

	c.pixelDeltaU = viewportU.Div(float64(c.ImageWidth))

	c.pixelDeltaV = viewportV.Div(float64(c.ImageHeight))

	viewportUpperLeft := c.center.
		Sub(c.w.Scale(c.FocusDist)).
		Sub(viewportU.Div(2)).
		Sub(viewportV.Div(2))

	c.pixel00Loc = viewportUpperLeft.Add(c.pixelDeltaU.Add(c.pixelDeltaV).Scale(0.5))

	defocusRadius := c.FocusDist * math.Tan(DegreesToRadians(c.DefocusAngle/2))
	c.defocusDiskU = c.u.Scale(defocusRadius)
	c.defocusDiskV = c.v.Scale(defocusRadius)
}

func (c *Camera) DefocusDiskSample() Point3 {
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

// =============================================================================
// RAY GENERATION
// =============================================================================

func (c *Camera) getRay(i, j int) Ray {
	offset := c.sampleSquare()
	rayTime := RandomDouble()

	currentCenter := c.centerMotion.At(rayTime)
	currentLookAt := c.lookAtMotion.At(rayTime)

	w := currentCenter.Sub(currentLookAt).Unit()
	u := Cross(c.Vup, w).Unit()
	v := Cross(w, u)

	theta := DegreesToRadians(c.Vfov)
	h := math.Tan(theta / 2)
	viewportHeight := 2 * h * c.FocusDist
	viewportWidth := viewportHeight * (float64(c.ImageWidth) / float64(c.ImageHeight))

	viewportU := u.Scale(viewportWidth)
	viewportV := v.Neg().Scale(viewportHeight)

	pixelDeltaU := viewportU.Div(float64(c.ImageWidth))
	pixelDeltaV := viewportV.Div(float64(c.ImageHeight))

	viewportUpperLeft := currentCenter.
		Sub(w.Scale(c.FocusDist)).
		Sub(viewportU.Div(2)).
		Sub(viewportV.Div(2))

	pixel00Loc := viewportUpperLeft.Add(pixelDeltaU.Add(pixelDeltaV).Scale(0.5))

	// Calculate pixel sample position
	pixelSample := pixel00Loc.
		Add(pixelDeltaU.Scale(float64(i) + offset.X)).
		Add(pixelDeltaV.Scale(float64(j) + offset.Y))

	// Apply defocus blur if enabled
	var rayOrigin Point3
	if c.DefocusAngle <= 0 {
		rayOrigin = currentCenter
	} else {
		// Defocus disk also moves with camera
		defocusRadius := c.FocusDist * math.Tan(DegreesToRadians(c.DefocusAngle/2))
		defocusDiskU := u.Scale(defocusRadius)
		defocusDiskV := v.Scale(defocusRadius)

		p := RandomInUnitDisk()
		rayOrigin = currentCenter.Add(defocusDiskU.Scale(p.X)).Add(defocusDiskV.Scale(p.Y))
	}

	rayDirection := pixelSample.Sub(rayOrigin)
	return NewRay(rayOrigin, rayDirection, rayTime)
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
	return c.skyColor(r)
}
func (c *Camera) skyColor(r Ray) Color {
	unitDirection := r.Direction().Unit()
	a := 0.5 * (unitDirection.Y + 1.0)
	white := Color{X: 1.0, Y: 1.0, Z: 1.0}
	blue := Color{X: 0.5, Y: 0.7, Z: 1.0}
	return white.Scale(1.0 - a).Add(blue.Scale(a))
}

// =============================================================================
// RENDERING
// =============================================================================

func (c *Camera) Render(world Hittable) {
	c.Initialize()

	img := image.NewRGBA(image.Rect(0, 0, c.ImageWidth, c.ImageHeight))

	const barWidth = 40

	for j := range c.ImageHeight {
		c.progressBar(j+1, c.ImageHeight, barWidth)
		for i := range c.ImageWidth {
			pixelColor := Color{X: 0, Y: 0, Z: 0}
			for sample := 0; sample < c.SamplesPerPixel; sample++ {
				r := c.getRay(i, j)
				pixelColor = pixelColor.Add(c.rayColor(r, c.MaxDepth, world))
			}
			c.writeColor(img, i, j, pixelColor)
		}
	}

	fmt.Fprintln(os.Stderr)
	c.saveImage(img, "image.png")
	fmt.Fprintln(os.Stdout, "Done. Image written to image.png")
}

// =============================================================================
// UTILITY FUNCTIONS
// =============================================================================
func (c *Camera) writeColor(img *image.RGBA, x, y int, pixelColor Color) {
	scale := c.pixelsSamplesScale
	r := pixelColor.X * scale
	g := pixelColor.Y * scale
	b := pixelColor.Z * scale

	// Apply gamma correction (gamma = 2.0)
	r = linearToGamma(r)
	g = linearToGamma(g)
	b = linearToGamma(b)

	// Clamp to [0, 1] and convert to [0, 255]
	intensity := NewInterval(0.0, 0.999)
	rByte := uint8(256 * intensity.Clamp(r))
	gByte := uint8(256 * intensity.Clamp(g))
	bByte := uint8(256 * intensity.Clamp(b))

	img.SetRGBA(x, y, color.RGBA{R: rByte, G: gByte, B: bByte, A: 255})
}

func linearToGamma(linearComponent float64) float64 {
	if linearComponent > 0 {
		return math.Sqrt(linearComponent)
	}
	return 0
}

func (c *Camera) saveImage(img *image.RGBA, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	if err := png.Encode(file, img); err != nil {
		panic(err)
	}

	fmt.Printf("Image saved to %s\n", filename)
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
