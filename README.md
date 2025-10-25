# Go Raytracing

A Go implementation following the "Ray Tracing in One Weekend" tutorial by Peter Shirley.

## Overview

This project implements a raytracer that generates PNG image files. The implementation features a fully positionable camera with adjustable field of view, rendering spheres with different materials (diffuse, metallic, and glass) against a sky gradient background.

## Requirements

- Go 1.25.1 or later

## Building and Running

1. Clone or download this repository
2. Navigate to the project directory
3. Run the program:

```bash
go run main.go
```

The program will generate an `image.png` file in the same directory.

## Latest Render

![Rendered Scene](image.png)

## Features

### Camera System

#### Positionable Camera

- **Position control**: Place the camera anywhere in 3D space using `LookFrom`
- **Look-at targeting**: Point the camera at any location using `LookAt`
- **Adjustable field of view**: Control zoom level with `Vfov` (vertical field of view in degrees)
- **Camera orientation**: Define "up" direction with `Vup` vector
- **Aspect ratio control**: Set image dimensions with `AspectRatio` and `ImageWidth`
- **Depth of field**: Adjust focus distance and defocus angle for realistic lens effects
- **Motion blur**: Support for camera movement during exposure time

#### Camera Presets

Three quality presets are available for quick configuration:

```go
// Quick preview - low quality, fast render
camera.ApplyPreset(rt.QuickPreview())
// Resolution: 400x225, 10 samples, 10 max depth

// Standard quality - balanced settings
camera.ApplyPreset(rt.StandardQuality())
// Resolution: 600x338, 100 samples, 50 max depth, depth of field enabled

// High quality - production renders
camera.ApplyPreset(rt.HighQuality())
// Resolution: 1200x675, 500 samples, 50 max depth, depth of field enabled
```

#### Builder Pattern API

Configure cameras with a fluent interface:

```go
camera := rt.NewCamera().
    WithResolution(800, 16.0/9.0).
    WithQuality(100, 50).
    WithPosition(
        rt.Point3{X: 13, Y: 2, Z: 3},
        rt.Point3{X: 0, Y: 0, Z: 0},
        rt.Vec3{X: 0, Y: 1, Z: 0},
    ).
    WithLens(20, 0.6, 10.0).
    WithMotionBlur(
        rt.Point3{X: 13, Y: 2.5, Z: 3},
        rt.Point3{X: 0, Y: 0, Z: 0},
    )
```

### Materials

- **Lambertian (Diffuse)**: Matte surfaces that scatter light randomly
- **Metal**: Reflective surfaces with adjustable fuzziness
- **Dielectric (Glass)**: Transparent materials with refraction and reflection
  - Supports realistic Fresnel effects (Schlick's approximation)
  - Can create hollow glass spheres (bubble effect)
  
### Acceleration Structures

#### Bounding Volume Hierarchy (BVH)

The raytracer implements a BVH tree structure for efficient ray-object intersection testing:

- **Automatic construction**: Build BVH from any scene with `NewBVHNodeFromList()`
- **Recursive subdivision**: Objects are recursively partitioned along random axes
- **Tight bounding boxes**: Each node maintains minimal bounding volumes
- **Fast ray culling**: Skip entire branches of the tree when rays miss bounding boxes

Performance improvement: BVH acceleration typically provides 10-100x speedup for scenes with hundreds of objects.

```go
world := rt.RandomScene()
bvh := rt.NewBVHNodeFromList(world)
camera.Render(bvh)
```

#### Bounding Boxes

All objects implement bounding box calculations:

- **Spheres**: Tight axis-aligned bounding boxes
- **Moving spheres**: Boxes encompass entire motion from time=0 to time=1
- **Planes**: Infinite bounding boxes
- **Lists**: Combined boxes of all children

### Rendering Quality

- **Anti-aliasing**: Configurable samples per pixel (default: 100)
- **Ray bouncing**: Adjustable maximum ray depth for indirect lighting (default: 50)
- **Gamma correction**: Automatic gamma 2.0 correction for realistic color output
- **Progress indicator**: Real-time rendering progress bar

## Output

The current implementation generates configurable resolution images:

- Ground Plane: Large yellow-green diffuse plane
- Center sphere: Blue diffuse material
- Left sphere: Glass material with hollow bubble effect
- Right sphere: Shiny gold metal
- Spheres that have random materials added to them
- Random Lambert spheres with some velocity to test motion blur

The image is saved as `image.png` in PNG format for easy viewing.

## Customization

### Basic Scene Setup

```go
// Create custom materials
materialGlass := rt.NewDielectric(1.5)
materialMetal := rt.NewMetal(rt.Color{X: 0.8, Y: 0.6, Z: 0.2}, 0.0)
materialMatte := rt.NewLambertian(rt.Color{X: 0.7, Y: 0.3, Z: 0.3})

// Add spheres to the scene
world := rt.NewHittableList()
world.Add(rt.NewSphere(rt.Point3{X: 0, Y: 0, Z: -1}, 0.5, materialGlass))
world.Add(rt.NewMovingSphere(center1, center2, 0.2, materialMatte))

// Build BVH for acceleration
bvh := rt.NewBVHNodeFromList(world)
```

### Advanced Scene Configuration

```go
// Create custom scene with specific material distribution
config := rt.DefaultSceneConfig()
config.DiffuseProb = 0.50
config.MetalProb = 0.30
config.GlassProb = 0.20
config.SphereGridBounds.MinA = -5
config.SphereGridBounds.MaxA = 5
config.DiffuseMotionBlur = true
config.MetalMotionBlur = false
config.GlassMotionBlur = false

world := rt.RandomSceneWithConfig(config)
```

### Camera Configuration

```go
camera := rt.NewCamera()
camera.ApplyPreset(rt.StandardQuality())

// Override specific settings
camera.Vfov = 20
camera.LookFrom = rt.Point3{X: 13, Y: 2, Z: 3}
camera.LookAt = rt.Point3{X: 0, Y: 0, Z: 0}
camera.DefocusAngle = 0.6
camera.FocusDist = 10.0

camera.Initialize()
camera.Render(bvh)
```

## Progress

### Ray Tracing in One Weekend

- [x] Basic image output (PNG format)
- [x] Progress indicator
- [x] Vector math utilities
- [x] Ray class
- [x] Simple sphere rendering
- [x] Surface normals and shading
- [x] Anti-aliasing (multi-sampling)
- [x] Diffuse materials (Lambertian)
- [x] Metal materials (with fuzziness)
- [x] Dielectric materials (glass with refraction)
- [x] Camera positioning (positionable camera)
- [x] Adjustable field of view
- [x] Defocus blur (depth of field)

### Ray Tracing: The Next Week

- [x] Motion blur
- [x] Bounding Volume Hierarchies (BVH)
- [ ] Image texture mapping
- [ ] Perlin noise
- [ ] Quadrilaterals
- [ ] Lights
- [ ] Instances
- [ ] Cornell Box scene
- [ ] Volumes

## Performance Notes

Rendering time depends on:

- Image resolution (`ImageWidth` × calculated height)
- Samples per pixel (`SamplesPerPixel`)
- Maximum ray depth (`MaxDepth`)
- Scene complexity (number of objects)

## Resources

- [Ray Tracing in One Weekend](https://raytracing.github.io/books/RayTracingInOneWeekend.html)
- [Ray Tracing: The Next Week](https://raytracing.github.io/books/RayTracingTheNextWeek.html)

## License

This project is for educational purposes following the public domain tutorial "Ray Tracing in One Weekend" and "RayTracing: The Next Week"
