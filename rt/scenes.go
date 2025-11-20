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

func RandomScene() (*HittableList, *Camera) {
	return RandomSceneWithConfig(DefaultSceneConfig())
}

func RandomSceneWithConfig(config SceneConfig) (*HittableList, *Camera) {
	world := NewHittableList()

	// =============================================================================
	// MATERIALS
	// =============================================================================
	groundChecker := NewCheckerTextureFromColors(
		0.32,
		config.GroundColor,
		Color{X: 0.9, Y: 0.9, Z: 0.9},
	)
	groundMaterial := NewLambertianTexture(groundChecker)

	// =============================================================================
	// GEOMETRY
	// =============================================================================
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

	// =============================================================================
	// CAMERA
	// =============================================================================
	camera := NewCameraBuilder().
		SetResolution(1200, 16.0/9.0).
		SetQuality(500, 50).
		SetPosition(
			Point3{X: 13, Y: 2, Z: 3},
			Point3{X: 0, Y: 0, Z: 0},
			Vec3{X: 0, Y: 1, Z: 0},
		).
		SetLens(20, 0.6, 10.0).
		EnableSkyGradient(true).
		Build()

	return world, camera
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

func CheckeredSpheresScene() (*HittableList, *Camera) {
	world := NewHittableList()

	// =============================================================================
	// MATERIALS
	// =============================================================================
	checker := NewCheckerTextureFromColors(
		0.32,
		Color{X: 0.2, Y: 0.3, Z: 0.1},
		Color{X: 0.9, Y: 0.9, Z: 0.9},
	)
	checkerMaterial := NewLambertianTexture(checker)

	// =============================================================================
	// GEOMETRY
	// =============================================================================
	// Bottom sphere (at y=-10)
	world.Add(NewSphere(Point3{X: 0, Y: -10, Z: 0}, 10, checkerMaterial))

	// Top sphere (at y=10)
	world.Add(NewSphere(Point3{X: 0, Y: 10, Z: 0}, 10, checkerMaterial))

	// =============================================================================
	// CAMERA
	// =============================================================================
	camera := NewCameraBuilder().
		SetResolution(600, 16.0/9.0).
		SetQuality(100, 50).
		SetPosition(
			Point3{X: 13, Y: 2, Z: 3},
			Point3{X: 0, Y: 0, Z: 0},
			Vec3{X: 0, Y: 1, Z: 0},
		).
		SetLens(20, 0, 10).
		EnableSkyGradient(true).
		Build()

	return world, camera
}

func SimpleScene() (*HittableList, *Camera) {
	world := NewHittableList()

	// =============================================================================
	// MATERIALS
	// =============================================================================
	materialGround := NewLambertian(Color{X: 0.8, Y: 0.8, Z: 0.0})
	materialCenter := NewLambertian(Color{X: 0.1, Y: 0.2, Z: 0.5})
	materialLeft := NewDielectric(1.5)
	materialBubble := NewDielectric(1.0 / 1.5)
	materialRight := NewMetal(Color{X: 0.8, Y: 0.6, Z: 0.2}, 0.0)

	// =============================================================================
	// GEOMETRY
	// =============================================================================
	world.Add(NewPlane(Point3{X: 0, Y: -0.5, Z: -1}, Vec3{X: 0, Y: 1, Z: 0}, materialGround))
	world.Add(NewSphere(Point3{X: 0, Y: 0, Z: -1}, 0.5, materialCenter))
	world.Add(NewSphere(Point3{X: -1, Y: 0, Z: -1}, 0.5, materialLeft))
	world.Add(NewSphere(Point3{X: -1, Y: 0, Z: -1}, 0.4, materialBubble))
	world.Add(NewSphere(Point3{X: 1, Y: 0, Z: -1}, 0.5, materialRight))

	// =============================================================================
	// CAMERA
	// =============================================================================
	camera := NewCameraBuilder().
		SetResolution(400, 16.0/9.0).
		SetQuality(100, 50).
		SetPosition(
			Point3{X: 0, Y: 0, Z: 2},
			Point3{X: 0, Y: 0, Z: -1},
			Vec3{X: 0, Y: 1, Z: 0},
		).
		SetLens(90, 0, 10).
		EnableSkyGradient(true).
		Build()

	return world, camera
}
func EarthScene() (*HittableList, *Camera) {
	world := NewHittableList()

	// =============================================================================
	// MATERIALS
	// =============================================================================
	earthTexture := NewImageTexture("earthmap.jpg")
	earthSurface := NewLambertianTexture(earthTexture)

	// =============================================================================
	// GEOMETRY
	// =============================================================================
	globe := NewSphere(Point3{X: 0, Y: 0, Z: 0}, 2, earthSurface)
	world.Add(globe)

	// =============================================================================
	// CAMERA
	// =============================================================================
	camera := NewCameraBuilder().
		SetResolution(800, 16.0/9.0).
		SetQuality(100, 50).
		SetPosition(
			Point3{X: 0, Y: 0, Z: 12},
			Point3{X: 0, Y: 0, Z: 0},
			Vec3{X: 0, Y: 1, Z: 0},
		).
		SetLens(20, 0, 10).
		EnableSkyGradient(true).
		Build()

	return world, camera
}
func PerlinSpheresScene() (*HittableList, *Camera) {
	world := NewHittableList()

	// =============================================================================
	// MATERIALS
	// =============================================================================
	pertext := NewNoiseTexture(4.0)
	perlMaterial := NewLambertianTexture(pertext)

	// =============================================================================
	// GEOMETRY
	// =============================================================================
	world.Add(NewSphere(Point3{X: 0, Y: 2, Z: 0}, 2, perlMaterial))
	world.Add(NewPlane(Point3{X: 0, Y: 0, Z: -1}, Vec3{X: 0, Y: 1, Z: 0}, perlMaterial))

	// =============================================================================
	// CAMERA
	// =============================================================================
	camera := NewCameraBuilder().
		SetResolution(600, 16.0/9.0).
		SetQuality(100, 50).
		SetPosition(
			Point3{X: 13, Y: 2, Z: -10},
			Point3{X: 0, Y: 1.5, Z: 0},
			Vec3{X: 0, Y: 1, Z: 0},
		).
		SetLens(20, 0, 10).
		EnableSkyGradient(true).
		Build()

	return world, camera
}
func QuadsScene() (*HittableList, *Camera) {
	world := NewHittableList()

	// =============================================================================
	// MATERIALS
	// =============================================================================
	leftRed := NewLambertian(Color{X: 1.0, Y: 0.2, Z: 0.2})
	backGreen := NewLambertian(Color{X: 0.2, Y: 1.0, Z: 0.2})
	rightBlue := NewLambertian(Color{X: 0.2, Y: 0.2, Z: 1.0})
	upperOrange := NewLambertian(Color{X: 1.0, Y: 0.5, Z: 0.0})
	lowerTeal := NewLambertian(Color{X: 0.2, Y: 0.8, Z: 0.8})

	// =============================================================================
	// GEOMETRY
	// =============================================================================
	world.Add(NewQuad(Point3{X: -3, Y: -2, Z: 5}, Vec3{X: 0, Y: 0, Z: -4}, Vec3{X: 0, Y: 4, Z: 0}, leftRed))
	world.Add(NewQuad(Point3{X: -2, Y: -2, Z: 0}, Vec3{X: 4, Y: 0, Z: 0}, Vec3{X: 0, Y: 4, Z: 0}, backGreen))
	world.Add(NewQuad(Point3{X: 3, Y: -2, Z: 1}, Vec3{X: 0, Y: 0, Z: 4}, Vec3{X: 0, Y: 4, Z: 0}, rightBlue))
	world.Add(NewQuad(Point3{X: -2, Y: 3, Z: 1}, Vec3{X: 4, Y: 0, Z: 0}, Vec3{X: 0, Y: 0, Z: 4}, upperOrange))
	world.Add(NewQuad(Point3{X: -2, Y: -3, Z: 5}, Vec3{X: 4, Y: 0, Z: 0}, Vec3{X: 0, Y: 0, Z: -4}, lowerTeal))

	// =============================================================================
	// CAMERA
	// =============================================================================
	camera := NewCameraBuilder().
		SetResolution(400, 1.0).
		SetQuality(100, 50).
		SetPosition(
			Point3{X: 0, Y: 0, Z: 9},
			Point3{X: 0, Y: 0, Z: 0},
			Vec3{X: 0, Y: 1, Z: 0},
		).
		SetLens(80, 0, 10).
		EnableSkyGradient(true).
		Build()

	return world, camera
}

// PrimitivesScene demonstrates all primitive types: sphere, circle, quad (as cube), triangle (as pyramid), and infinite plane
func PrimitivesScene() (*HittableList, *Camera) {
	world := NewHittableList()

	// =============================================================================
	// MATERIALS
	// =============================================================================
	redMat := NewLambertian(Color{X: 0.8, Y: 0.1, Z: 0.1})
	greenMat := NewLambertian(Color{X: 0.1, Y: 0.8, Z: 0.1})
	blueMat := NewLambertian(Color{X: 0.1, Y: 0.1, Z: 0.8})
	metalMat := NewMetal(Color{X: 0.7, Y: 0.7, Z: 0.7}, 0.1)
	lightMat := NewDiffuseLight(NewSolidColor(Color{X: 1, Y: 1, Z: 1}))

	checkerMat := NewLambertianTexture(NewCheckerTextureFromColors(1.0,
		Color{X: 0.0, Y: 0.0, Z: 0.0},
		Color{X: 0.9, Y: 0.9, Z: 0.9}))

	// =============================================================================
	// GROUND PLANE (Infinite Plane with Checker Pattern)
	// =============================================================================
	world.Add(NewPlane(Point3{X: 0, Y: -1, Z: 0}, Vec3{X: 0, Y: 1, Z: 0}, checkerMat))

	// =============================================================================
	// LEFT: Circle (Disk)
	// =============================================================================
	world.Add(NewCircle(
		Point3{X: -5, Y: 0, Z: 0},
		Vec3{X: 0, Y: 1, Z: 0}, // Normal pointing at camera
		0.9,
		redMat,
	))

	// =============================================================================
	// CENTER-LEFT: Triangle Pyramid
	// =============================================================================
	pyramidX := -2.5
	pyramidHeight := 1.8
	pyramidBase := 1.4

	world.Add(Pyramid(Point3{X: pyramidX, Y: -1, Z: 0}, pyramidBase, pyramidHeight, greenMat))

	// =============================================================================
	// CENTER: Glass Sphere
	// =============================================================================
	world.Add(NewSphere(Point3{X: 0, Y: 0.6, Z: 0}, 0.8, NewDielectric(1.5)))

	// =============================================================================
	// CENTER-RIGHT: Quad Cube
	// =============================================================================
	cubeX := 2.5
	cubeSize := 1.0

	world.Add(Box(
		Point3{X: cubeX - cubeSize/2, Y: -1, Z: -cubeSize / 2},
		Point3{X: cubeX + cubeSize/2, Y: -1 + cubeSize, Z: cubeSize / 2},
		blueMat,
	))

	// =============================================================================
	// OVERHEAD LIGHT SOURCE
	// =============================================================================
	areaLight := NewQuad(
		Point3{X: -2, Y: 5, Z: -2},
		Vec3{X: 4, Y: 0, Z: 0},
		Vec3{X: 0, Y: 0, Z: 4},
		lightMat,
	)
	world.Add(areaLight)

	// =============================================================================
	// RIGHT: Metal Sphere
	// =============================================================================
	world.Add(NewSphere(Point3{X: 5, Y: 0.6, Z: 0}, 0.8, metalMat))

	camera := NewCameraBuilder().
		SetResolution(800, 16.0/9.0).
		SetQuality(50, 50).
		SetPosition(
			Point3{X: 0, Y: 2, Z: 10},
			Point3{X: 0, Y: 0, Z: 0},
			Vec3{X: 0, Y: 1, Z: 0},
		).
		SetLens(45, 0, 10).
		SetBackground(Color{0, 0, 0}).
		EnableSkyGradient(false).
		AddLight(areaLight).
		Build()

	return world, camera
}

// ==================================================================================
// Cornell Box Scene
// ==================================================================================
func CornellBoxScene() (*HittableList, *Camera) {
	world := NewHittableList()

	whiteMat := NewLambertian(Color{X: 0.73, Y: 0.73, Z: 0.73})
	redMat := NewLambertian(Color{X: 0.65, Y: 0.05, Z: 0.05})
	greenMat := NewLambertian(Color{X: 0.12, Y: 0.45, Z: 0.15})
	lightMat := NewDiffuseLight(NewSolidColor(Color{X: 3, Y: 3, Z: 3}))

	areaLight := NewQuad(
		Point3{X: 213, Y: 554, Z: 227},
		Vec3{X: 130, Y: 0, Z: 0},
		Vec3{X: 0, Y: 0, Z: 105},
		lightMat,
	)
	world.Add(areaLight)

	// Walls
	world.Add(NewQuad(
		Point3{X: 555, Y: 0, Z: 0},
		Vec3{X: 0, Y: 555, Z: 0},
		Vec3{X: 0, Y: 0, Z: 555},
		greenMat,
	))
	world.Add(NewQuad(
		Point3{X: 0, Y: 0, Z: 0},
		Vec3{X: 0, Y: 555, Z: 0},
		Vec3{X: 0, Y: 0, Z: 555},
		redMat,
	))
	world.Add(NewQuad(
		Point3{X: 0, Y: 0, Z: 0},
		Vec3{X: 555, Y: 0, Z: 0},
		Vec3{X: 0, Y: 0, Z: 555},
		whiteMat,
	))
	world.Add(NewQuad(
		Point3{X: 555, Y: 555, Z: 555},
		Vec3{X: -555, Y: 0, Z: 0},
		Vec3{X: 0, Y: 0, Z: -555},
		whiteMat,
	))
	world.Add(NewQuad(
		Point3{X: 0, Y: 0, Z: 555},
		Vec3{X: 555, Y: 0, Z: 0},
		Vec3{X: 0, Y: 555, Z: 0},
		whiteMat,
	))
	// Light
	world.Add(NewQuad(
		Point3{X: 213, Y: 554, Z: 227},
		Vec3{X: 130, Y: 0, Z: 0},
		Vec3{X: 0, Y: 0, Z: 105},
		lightMat,
	))
	//Boxes
	box1 := Box(
		Point3{X: 0, Y: 0, Z: 0},
		Point3{X: 165, Y: 330, Z: 165},
		whiteMat,
	)
	box1Xform := NewTransform().
		SetScale(Vec3{X: 1.0, Y: 1.0, Z: 1.0}).
		SetRotationY(15).
		SetPosition(Vec3{X: 265, Y: 0, Z: 295}).
		Apply(box1)
	world.Add(box1Xform)

	box2 := Box(
		Point3{X: 0, Y: 0, Z: 0},
		Point3{X: 165, Y: 165, Z: 165},
		whiteMat,
	)
	box2Xform := NewTransform().
		SetScale(Vec3{X: 1.0, Y: 1.0, Z: 1.0}).
		SetRotationY(-18).
		SetPosition(Vec3{X: 130, Y: 0, Z: 65}).
		Apply(box2)
	world.Add(box2Xform)

	camera := NewCameraBuilder().
		SetResolution(600, 1.0).
		SetQuality(250, 10).
		SetPosition(
			Point3{X: 278, Y: 278, Z: -800},
			Point3{X: 278, Y: 278, Z: 0},
			Vec3{X: 0, Y: 1, Z: 0},
		).
		SetLens(40, 0, 10).
		SetBackground(Color{0, 0, 0}).
		AddLight(areaLight).
		Build()

	return world, camera
}
