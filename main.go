package main

import (
	"go-raytracing/rt"
	"math/rand"
)

func randomScene() *rt.HittableList {
	world := rt.NewHittableList()

	groundMaterial := rt.NewLambertian(rt.Color{X: 0.5, Y: 0.5, Z: 0.5})
	world.Add(rt.NewPlane(rt.Point3{X: 0, Y: 0, Z: -1}, rt.Vec3{X: 0, Y: 1, Z: 0}, groundMaterial))

	for a := -11; a < 11; a++ {
		for b := -20; b < 20; b++ {
			chooseMat := rand.Float64()
			center := rt.Point3{
				X: float64(a) + 0.9*rand.Float64(),
				Y: 0.2,
				Z: float64(b) + 0.9*rand.Float64(),
			}
			if center.Sub(rt.Point3{X: 4, Y: 0.2, Z: 0}).Len() > 0.9 {
				var sphereMaterial rt.Material

				if chooseMat < 0.8 {
					albedo := rt.Color{X: rand.Float64(), Y: rand.Float64(), Z: rand.Float64()}.
						Mult(rt.Color{X: rand.Float64(), Y: rand.Float64(), Z: rand.Float64()})
					sphereMaterial = rt.NewLambertian((albedo))
					world.Add(rt.NewSphere(center, 0.2, sphereMaterial))
				} else if chooseMat < 0.95 {
					albedo := rt.Color{
						X: 0.5 + rand.Float64()*0.5,
						Y: 0.5 + rand.Float64()*0.5,
						Z: 0.5 + rand.Float64()*0.5,
					}
					fuzz := rand.Float64() * 0.5
					sphereMaterial = rt.NewMetal(albedo, fuzz)
					world.Add(rt.NewSphere(center, 0.2, sphereMaterial))
				} else {
					sphereMaterial = rt.NewDielectric(1.5)
					world.Add(rt.NewSphere(center, 0.2, sphereMaterial))
				}
			}
		}
	}
	material1 := rt.NewDielectric(1.5)
	world.Add(rt.NewSphere(rt.Point3{X: 0, Y: 1, Z: 0}, 1.0, material1))

	material2 := rt.NewLambertian(rt.Color{X: 0.4, Y: 0.2, Z: 0.1})
	world.Add(rt.NewSphere(rt.Point3{X: -4, Y: 1, Z: 0}, 1.0, material2))

	material3 := rt.NewMetal(rt.Color{X: 0.7, Y: 0.6, Z: 0.5}, 0.0)
	world.Add(rt.NewSphere(rt.Point3{X: 4, Y: 1, Z: 0}, 1.0, material3))

	return world

}

func main() {
	world := randomScene()

	//material time
	// materialGround := rt.NewLambertian(rt.Color{X: 0.5, Y: 0.5, Z: 0.0})
	// materialCenter := rt.NewLambertian(rt.Color{X: 1.0, Y: 0.5, Z: 0.0})
	// materialLeft := rt.NewDielectric(1.5)
	// materialBubble := rt.NewDielectric(1.0 / 1.5)
	// materialRight := rt.NewMetal(rt.Color{X: 1.0, Y: 1.0, Z: 1.0}, 0.0)

	// Make the camera
	camera := rt.NewCamera()
	camera.AspectRatio = 16.0 / 9.0
	camera.ImageWidth = 800
	camera.SamplesPerPixel = 500
	camera.MaxDepth = 50
	camera.Vfov = 20
	camera.DefocusAngle = 0.75
	camera.FocusDist = 10.0

	//position camera
	camera.LookFrom = rt.Point3{X: 13, Y: 2, Z: 3}
	camera.LookAt = rt.Point3{X: 0, Y: 0, Z: 0}
	camera.Vup = rt.Vec3{X: 0, Y: 1, Z: 0}

	//camera.LookFrom = rt.Point3{X: -2, Y: 2, Z: 1}
	//camera.LookAt = rt.Point3{X: 0, Y: 0, Z: -1}
	//camera.Vup = rt.Vec3{X: 0, Y: 1, Z: 0}

	// Sphere time!

	// world := rt.NewHittableList()

	//world.Add(rt.NewSphere(rt.Point3{X: 0, Y: -100.5, Z: -1}, 100, materialGround))
	// world.Add(rt.NewPlane(rt.Point3{X: 0, Y: -0.5, Z: -1}, rt.Vec3{X: 0, Y: 1, Z: 0}, materialGround))
	// world.Add(rt.NewSphere(rt.Point3{X: 0, Y: 0, Z: -1}, 0.5, materialCenter))
	// world.Add(rt.NewSphere(rt.Point3{X: -1, Y: 0, Z: -1}, 0.5, materialLeft))
	// world.Add(rt.NewSphere(rt.Point3{X: -1, Y: 0, Z: -1}, 0.4, materialBubble))
	// world.Add(rt.NewSphere(rt.Point3{X: 1, Y: 0, Z: -1}, 0.5, materialRight))

	// Render time
	camera.Render(world)
}
