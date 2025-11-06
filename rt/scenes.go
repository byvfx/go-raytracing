//TODO: add cameras that corresspond with each scene.

package rt

import (
	"math/rand"
)

type SceneConfig struct {
	GroundColor      Color
	SphereGridBounds struct{ MinA, MaxA, MinB, MaxB int }
	MovingSphereProb float64
	LambertProb      float64
	DielectricProb   float64
	MetalProb        float64
	LargeSpheresY    float64
}

func DefaultSceneConfig() SceneConfig {
	return SceneConfig{
		GroundColor: Color{X: 0.5, Y: 0.5, Z: 0.5},
		SphereGridBounds: struct {
			MinA int
			MaxA int
			MinB int
			MaxB int
		}{-10, 10, -10, 10},
		MovingSphereProb: 0,
		LambertProb:      0.3,
		DielectricProb:   0.3,
		MetalProb:        0.3,
		LargeSpheresY:    1.0,
	}
}

func RandomScene() *HittableList {
	return RandomSceneWithConfig(DefaultSceneConfig())
}

func RandomSceneWithConfig(config SceneConfig) *HittableList {
	world := NewHittableList()
	groundChecker := NewCheckerTextureFromColors(
		0.32,
		config.GroundColor,
		Color{X: 0.9, Y: 0.9, Z: 0.9},
	)
	groundMaterial := NewLambertianTexture(groundChecker)
	world.Add(NewPlane(Point3{X: 0, Y: 0, Z: -1}, Vec3{X: 0, Y: 1, Z: 0}, groundMaterial))

	for a := config.SphereGridBounds.MinA; a < config.SphereGridBounds.MaxA; a++ {
		for b := config.SphereGridBounds.MinB; b < config.SphereGridBounds.MaxB; b++ {
			chooseMat := rand.Float64()
			center := Point3{
				X: float64(a) + 0.9*rand.Float64(),
				Y: 0.2,
				Z: float64(b) + 0.9*rand.Float64(),
			}

			if center.Sub(Point3{X: 4, Y: 0.2, Z: 0}).Len() > 0.9 {
				addRandomSphere(world, center, chooseMat, config)
			}
		}
	}
	addLargeSpheres(world, config.LargeSpheresY)

	return world
}
func addRandomSphere(world *HittableList, center Point3, chooseMat float64, config SceneConfig) {
	var sphereMaterial Material

	lambertThreshold := config.LambertProb
	metalThreshold := config.MetalProb + lambertThreshold
	dielectricThreshold := config.DielectricProb + metalThreshold

	if chooseMat < lambertThreshold {
		albedo := Color{
			X: rand.Float64() * rand.Float64(),
			Y: rand.Float64() * rand.Float64(),
			Z: rand.Float64() * rand.Float64(),
		}
		sphereMaterial = NewLambertian(albedo)
		center2 := center.Add(Vec3{X: 0, Y: RandomDoubleRange(0, 0.5), Z: 0})
		world.Add(NewMovingSphere(center, center2, 0.2, sphereMaterial))
	} else if chooseMat < metalThreshold {

		albedo := Color{
			X: 0.5 + rand.Float64()*0.5,
			Y: 0.5 + rand.Float64()*0.5,
			Z: 0.5 + rand.Float64()*0.5,
		}
		fuzz := rand.Float64() * 0.5
		sphereMaterial = NewMetal(albedo, fuzz)
		world.Add(NewSphere(center, 0.2, sphereMaterial))
	} else if chooseMat < dielectricThreshold {

		sphereMaterial = NewDielectric(1.5)
		world.Add(NewSphere(center, 0.2, sphereMaterial))
	}
}

func addLargeSpheres(world *HittableList, y float64) {
	// Glass sphere (center)
	material1 := NewDielectric(1.5)
	world.Add(NewSphere(Point3{X: 0, Y: y, Z: 0}, 1.0, material1))

	// Diffuse sphere (left)
	material2 := NewLambertian(Color{X: 0.4, Y: 0.2, Z: 0.1})
	world.Add(NewSphere(Point3{X: -4, Y: y, Z: 0}, 1.0, material2))

	// Metal sphere (right)
	material3 := NewMetal(Color{X: 0.7, Y: 0.6, Z: 0.5}, 0.0)
	world.Add(NewSphere(Point3{X: 4, Y: y, Z: 0}, 1.0, material3))
}

func CheckeredSpheresScene() *HittableList {
	world := NewHittableList()

	checker := NewCheckerTextureFromColors(
		0.32,
		Color{X: 0.2, Y: 0.3, Z: 0.1},
		Color{X: 0.9, Y: 0.9, Z: 0.9},
	)

	checkerMaterial := NewLambertianTexture(checker)

	// Bottom sphere (at y=-10)
	world.Add(NewSphere(Point3{X: 0, Y: -10, Z: 0}, 10, checkerMaterial))

	// Top sphere (at y=10)
	world.Add(NewSphere(Point3{X: 0, Y: 10, Z: 0}, 10, checkerMaterial))

	return world
}

func SimpleScene() *HittableList {
	world := NewHittableList()

	materialGround := NewLambertian(Color{X: 0.8, Y: 0.8, Z: 0.0})
	materialCenter := NewLambertian(Color{X: 0.1, Y: 0.2, Z: 0.5})
	materialLeft := NewDielectric(1.5)
	materialBubble := NewDielectric(1.0 / 1.5)
	materialRight := NewMetal(Color{X: 0.8, Y: 0.6, Z: 0.2}, 0.0)

	world.Add(NewPlane(Point3{X: 0, Y: -0.5, Z: -1}, Vec3{X: 0, Y: 1, Z: 0}, materialGround))
	world.Add(NewSphere(Point3{X: 0, Y: 0, Z: -1}, 0.5, materialCenter))
	world.Add(NewSphere(Point3{X: -1, Y: 0, Z: -1}, 0.5, materialLeft))
	world.Add(NewSphere(Point3{X: -1, Y: 0, Z: -1}, 0.4, materialBubble))
	world.Add(NewSphere(Point3{X: 1, Y: 0, Z: -1}, 0.5, materialRight))

	return world
}
func EarthScene() *HittableList {
	world := NewHittableList()

	earthTexture := NewImageTexture("earthmap.jpg")
	earthSurface := NewLambertianTexture(earthTexture)
	globe := NewSphere(Point3{X: 0, Y: 0, Z: 0}, 2, earthSurface)

	world.Add(globe)
	return world
}
func EarthCamera() *Camera {
	camera := NewCamera()
	camera.AspectRatio = 16.0 / 9.0
	camera.ImageWidth = 800
	camera.SamplesPerPixel = 100
	camera.MaxDepth = 50
	camera.Vfov = 20
	camera.LookFrom = Point3{X: 0, Y: 0, Z: 12}
	camera.LookAt = Point3{X: 0, Y: 0, Z: 0}
	camera.Vup = Vec3{X: 0, Y: 1, Z: 0}
	camera.DefocusAngle = 0
	camera.Initialize()

	return camera
}
func PerlinSpheresScene() *HittableList {
	world := NewHittableList()

	pertext := NewNoiseTexture(4.0)

	world.Add(NewSphere(Point3{X: 0, Y: 2, Z: 0}, 2, NewLambertianTexture(pertext)))

	world.Add(NewPlane(Point3{X: 0, Y: 0, Z: -1}, Vec3{X: 0, Y: 1, Z: 0}, NewLambertianTexture(pertext)))

	return world
}

// PerlinSpheresCamera returns the camera configuration for the Perlin spheres scene
func PerlinSpheresCamera() *Camera {
	camera := NewCamera()
	camera.AspectRatio = 16.0 / 9.0
	camera.ImageWidth = 600
	camera.SamplesPerPixel = 100
	camera.MaxDepth = 50
	camera.Vfov = 20
	camera.LookFrom = Point3{X: 13, Y: 2, Z: -10}
	camera.LookAt = Point3{X: 0, Y: 1.5, Z: 0}
	camera.Vup = Vec3{X: 0, Y: 1, Z: 0}
	camera.DefocusAngle = 0
	camera.Initialize()

	return camera
}
