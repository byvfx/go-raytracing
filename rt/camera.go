// TODO add option for Depth of Field, so we can set a global flag that will enable/disable defocus blur
// TODO add independent samples for area lights vs camera samples
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
	FreeCamera      bool
	Forward         Vec3
	Background      Color
	UseSkyGradient  bool
	Lights          []Hittable

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
		FreeCamera:      false,
		Forward:         Vec3{0, 0, -1},
		Background:      Color{X: 0.0, Y: 0.0, Z: 0.0},
		UseSkyGradient:  false,
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
	FreeCamera      bool
	Forward         Vec3
	Background      Color
	UseSkyGradient  bool
}

// camera presets
// TODO: add presets with blackground color
// TODO: add presets with free camera
// TODO: add skyColor a few presets
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
		Background:      Color{X: 0.5, Y: 0.7, Z: 1.0},
		UseSkyGradient:  true,
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
		Background:      Color{X: 0.5, Y: 0.7, Z: 1.0},
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
		Background:      Color{X: 0.5, Y: 0.7, Z: 1.0},
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
	c.FreeCamera = preset.FreeCamera
	c.Forward = preset.Forward
	c.Background = preset.Background
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
func (c *Camera) SetMotion(lookFrom2, lookAt2 Point3) *Camera {
	c.LookFrom2 = lookFrom2
	c.LookAt2 = lookAt2
	c.CameraMotion = true
	return c
}

func (c *Camera) SetVFOV(vfov float64) *Camera {
	c.Vfov = vfov
	return c
}

func (c *Camera) SetDefocus(angle, focusDist float64) *Camera {
	c.DefocusAngle = angle
	c.FocusDist = focusDist
	return c
}

func (c *Camera) DisableMotion() *Camera {
	c.CameraMotion = false
	return c
}
func (c *Camera) EnableFreeCamera(position Point3, forward Vec3, vup Vec3) *Camera {
	c.LookFrom = position
	c.Forward = forward.Unit()
	c.Vup = vup.Unit()
	c.FreeCamera = true
	return c
}
func (c *Camera) SetBackground(color Color) *Camera {
	c.Background = color
	return c
}

func (c *Camera) EnableSkyGradient(enable bool) *Camera {
	c.UseSkyGradient = enable
	return c
}
func (c *Camera) AddLight(light Hittable) *Camera {
	c.Lights = append(c.Lights, light)
	return c
}

func (c *Camera) Build() *Camera {
	c.Initialize()
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

	if c.FreeCamera {
		c.w = c.Forward.Neg()
	} else {
		c.w = c.center.Sub(c.LookAt).Unit()
	}

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

func (c *Camera) sampleSquare() Vec3 {
	return Vec3{
		X: RandomDouble() - 0.5,
		Y: RandomDouble() - 0.5,
		Z: 0,
	}
}

func (c *Camera) defocusDiskSample(center Point3, u, v Vec3) Point3 {
	p := RandomInUnitDisk()
	defocusRadius := c.FocusDist * math.Tan(DegreesToRadians(c.DefocusAngle/2))
	defocusDiskU := u.Scale(defocusRadius)
	defocusDiskV := v.Scale(defocusRadius)
	p = RandomInUnitDisk()

	return center.Add(defocusDiskU.Scale(p.X)).Add(defocusDiskV.Scale(p.Y))
}

// =============================================================================
// RAY GENERATION
// =============================================================================

func (c *Camera) GetRay(i, j int) Ray {
	offset := c.sampleSquare()
	rayTime := RandomDouble()

	currentCenter := c.centerMotion.At(rayTime)
	var u, v, w Vec3

	if c.FreeCamera {
		w = c.Forward.Neg()
		u = Cross(c.Vup, w).Unit()
		v = Cross(w, u)
	} else {
		currentLookAt := c.lookAtMotion.At(rayTime)
		w = currentCenter.Sub(currentLookAt).Unit()
		u = Cross(c.Vup, w).Unit()
		v = Cross(w, u)
	}

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
		rayOrigin = c.defocusDiskSample(currentCenter, u, v)
	}

	rayDirection := pixelSample.Sub(rayOrigin)
	return NewRay(rayOrigin, rayDirection, rayTime)
}

// sending out them color rays
func (c *Camera) RayColor(r Ray, depth int, world Hittable) Color {
	if depth <= 0 {
		return Color{X: 0, Y: 0, Z: 0}
	}

	rec := &HitRecord{}

	// If the ray hits nothing, return the background color
	if !world.Hit(r, NewInterval(0.001, math.Inf(1)), rec) {
		if c.UseSkyGradient {
			return c.SkyGradient(r)
		}
		return c.Background
	}

	var attenuation Color
	var scattered Ray

	colorFromEmission := rec.Mat.Emitted(rec.U, rec.V, rec.P)

	if !rec.Mat.Scatter(r, rec, &attenuation, &scattered) {
		// Hit a light source - return emission only
		return colorFromEmission
	}

	// For scattered rays, add direct light sampling (NEE)
	directLight := Color{X: 0, Y: 0, Z: 0}
	if len(c.Lights) > 0 {
		lightIdx := int(RandomDouble() * float64(len(c.Lights)))
		if lightIdx >= len(c.Lights) {
			lightIdx = len(c.Lights) - 1
		}
		directLight = c.sampleLight(rec.P, rec.Normal, world, lightIdx)
	}

	// Recursively trace scattered ray (indirect illumination)
	colorFromScatter := attenuation.Mult(c.RayColor(scattered, depth-1, world))

	// Combine: direct lighting + indirect lighting
	// Note: We don't add colorFromEmission here because this surface scatters (not a light)
	return directLight.Mult(attenuation).Add(colorFromScatter)
}

func (c *Camera) SkyGradient(r Ray) Color {
	unitDirection := r.Direction().Unit()
	a := 0.5 * (unitDirection.Y + 1.0)
	white := Color{X: 1.0, Y: 1.0, Z: 1.0}
	blue := Color{X: 0.5, Y: 0.7, Z: 1.0}
	return white.Scale(1.0 - a).Add(blue.Scale(a))
}

var (
	BackgroundSkyColor = Color{X: 0.5, Y: 0.7, Z: 1.0}
	BackgroundBlack    = Color{X: 0.0, Y: 0.0, Z: 0.0}
	BackgroundWhite    = Color{X: 1.0, Y: 1.0, Z: 1.0}
	BackgroundGray     = Color{X: 0.5, Y: 0.5, Z: 0.5}
	BackgroundSunset   = Color{X: 1.0, Y: 0.5, Z: 0.3}
	BackgroundNight    = Color{X: 0.05, Y: 0.05, Z: 0.2}
)

// In camera.go - replace the sampleLight method

func (c *Camera) sampleLight(hitPoint Point3, hitNormal Vec3, world Hittable, lightIdx int) Color {
	if lightIdx >= len(c.Lights) {
		return Color{X: 0, Y: 0, Z: 0}
	}

	light := c.Lights[lightIdx]
	lightQuad, ok := light.(*Quad)
	if !ok {
		return Color{X: 0, Y: 0, Z: 0}
	}

	// Sample random point on light surface
	lightPoint := lightQuad.SamplePoint()

	// Direction from hit point to light sample
	toLight := lightPoint.Sub(hitPoint)
	distanceToLight := toLight.Len()
	lightDir := toLight.Unit()

	// Check if light is on the same side as surface normal
	cosTheta := Dot(hitNormal, lightDir)
	if cosTheta <= 0 {
		return Color{X: 0, Y: 0, Z: 0}
	}

	// Shadow ray test
	shadowRay := NewRay(hitPoint, lightDir, 0)
	shadowRec := &HitRecord{}

	if world.Hit(shadowRay, NewInterval(0.001, distanceToLight-0.001), shadowRec) {
		// Something is blocking the light
		return Color{X: 0, Y: 0, Z: 0}
	}

	// Get light emission
	emission := lightQuad.mat.Emitted(0, 0, lightPoint)

	// Calculate PDF using solid angle conversion
	lightArea := lightQuad.Area()
	cosLightAngle := math.Abs(Dot(lightQuad.normal, lightDir.Neg()))

	if cosLightAngle < 0.001 {
		return Color{X: 0, Y: 0, Z: 0}
	}

	// PDF for uniform area sampling: 1/area
	// Converting to solid angle: pdf_solidAngle = (distance² / (cos_light * area))
	// Light contribution: emission * cos_surface / pdf_solidAngle
	// Simplifies to: emission * cos_surface * cos_light * area / distance²
	pdfSolidAngle := (distanceToLight * distanceToLight) / (cosLightAngle * lightArea)

	// Light contribution with proper PDF weighting
	contribution := emission.Scale(cosTheta / pdfSolidAngle)

	// Multiply by number of lights (probability of selecting this light is 1/numLights)
	contribution = contribution.Scale(float64(len(c.Lights)))

	// CLAMP to prevent fireflies - adjust this threshold as needed
	maxComponent := 20.0
	contribution.X = math.Min(contribution.X, maxComponent)
	contribution.Y = math.Min(contribution.Y, maxComponent)
	contribution.Z = math.Min(contribution.Z, maxComponent)

	return contribution
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
				r := c.GetRay(i, j)
				pixelColor = pixelColor.Add(c.RayColor(r, c.MaxDepth, world))
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
