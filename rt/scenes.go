package rt

type SceneConfig struct {
	GroundColor      Color
	SphereGridBounds struct{ MinA, MaxA, MinB, MaxB int }
	MovingSphereProb float64
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
		}{-11, 11, -20, 20},
		MovingSphereProb: 0,
		MetalProb:        0.95,
		LargeSpheresY:    1.0,
	}
}

func RandomScene() *HittableList {
	return RandomSceneWithConfig(DefaultSceneConfig())
}

func RandomSceneWithConfig(config SceneConfig) *HittableList {
	world := NewHittableList()
	groundMaterial := NewLambertian(config.GroundColor)
	world.Add(NewPlane(Point3{X: 0, Y: 0, Z: -1}, Vec3{X: 0, Y: 1, Z: 0}, groundMaterial))

	return world
}
