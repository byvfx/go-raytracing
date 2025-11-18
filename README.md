# Go Raytracing

Go raytracer following "Ray Tracing in One Weekend" series by Peter Shirley.

## Requirements

- Go 1.25.1+
- Ebiten v2 for progressive display

## Running

```bash
go run main.go
```

## Lastest Render

![Latest Render](image.png)

Renders to window with progressive scanline display. Saves final image as `image.png`.

## Features

### Rendering

- Progressive scanline rendering with live preview (Ebiten)
- Anti-aliasing via multi-sampling (configurable samples/pixel)
- Gamma correction (gamma 2.0)
- Max ray depth control for indirect lighting

### Camera

- Positionable in 3D space (`LookFrom`, `LookAt`)
- Adjustable field of view (`Vfov`)
- Depth of field (defocus blur via `DefocusAngle`, `FocusDist`)
- Camera motion blur support

**Presets:**

- `QuickPreview()` - 400x225, 10 samples, 10 depth
- `StandardQuality()` - 600x338, 100 samples, 50 depth, DOF enabled
- `HighQuality()` - 1200x675, 500 samples, 50 depth, DOF enabled

**Builder API:**

```go
camera := rt.NewCameraBuilder().
    SetResolution(800, 16.0/9.0).
    SetQuality(100, 50).
    SetPosition(
        rt.Point3{X: 13, Y: 2, Z: 3},
        rt.Point3{X: 0, Y: 0, Z: 0},
        rt.Vec3{X: 0, Y: 1, Z: 0},
    ).
    SetLens(20, 0.6, 10.0).
    AddLight(areaLight).
    Build()
```

### Materials

- **Lambertian** - Diffuse/matte surfaces
- **Metal** - Reflective surfaces w/ adjustable fuzz
- **Dielectric** - Glass/transparent materials w/ refraction, Fresnel effects (Schlick approximation), hollow sphere support
- **DiffuseLight** - Emissive surfaces for area lights

### Textures

- **SolidColor** - Uniform color
- **CheckerTexture** - 3D procedural checkerboard
- **ImageTexture** - Image-based textures (PNG/JPEG support)
- **NoiseTexture** - Perlin noise-based procedural texture

### Acceleration

**BVH (Bounding Volume Hierarchy):**

- Automatic construction from scene
- Recursive axis-aligned subdivision
- Ray culling via bounding box tests
- 10-100x speedup for large scenes

```go
bvh := rt.NewBVHNodeFromList(world)
```

### Geometry

- **Sphere** - Static and moving spheres
- **Plane** - Infinite planes
- **Quad** - Axis-aligned quadrilaterals
- **Triangle** - Basic triangle primitive
- **Circle/Disk** - Flat circular surfaces
- **Box** - Compound primitive (6 quads)
- **Pyramid** - Compound primitive (4 triangles + base)
- **BVHNode** - Acceleration structure node
- All objects have axis-aligned bounding boxes

### Transforms

- **Translate** - Position offset
- **RotateX/Y/Z** - Axis-aligned rotation
- **Scale** - Uniform and non-uniform scaling
- **Transform builder** - Chainable API with SRT ordering (Scale-Rotate-Translate)

### Lighting

- **Next Event Estimation (NEE)** - Direct light sampling for reduced noise
- **Area lights** - Quad-based emissive surfaces
- **Light registration** - Camera tracks lights for importance sampling
- **Shadow rays** - Visibility testing with proper PDF weighting

### Scenes

Predefined scenes:

- `RandomScene()` - Configurable random sphere distribution
- `CheckeredSpheresScene()` - Two checkered spheres
- `SimpleScene()` - Basic test scene
- `PerlinSpheresScene()` - Spheres with Perlin noise textures
- `EarthScene()` - Textured Earth sphere
- `QuadsScene()` - Box-like room made of quads
- `CornellBoxScene()` - Classic Cornell Box setup
- `PrimitivesScene()` - Scene showcasing various primitives

`SceneConfig` allows control over material probabilities, motion blur per material, grid bounds, etc.

## Usage

```go
// Basic setup
world := rt.RandomScene()
bvh := rt.NewBVHNodeFromList(world)

camera := rt.NewCamera()
camera.ApplyPreset(rt.StandardQuality())
camera.LookFrom = rt.Point3{X: 13, Y: 2, Z: 3}
camera.LookAt = rt.Point3{X: 0, Y: 0, Z: 0}
camera.Initialize()

renderer := rt.NewProgressiveRenderer(camera, bvh)
ebiten.RunGame(renderer)
```

```go
// Camera Builder API
camera := rt.NewCameraBuilder().
    SetResolution(800, 16.0/9.0).
    SetQuality(100, 50).
    SetPosition(
        rt.Point3{X: 13, Y: 2, Z: 3},
        rt.Point3{X: 0, Y: 0, Z: 0},
        rt.Vec3{X: 0, Y: 1, Z: 0},
    ).
    SetLens(20, 0.6, 10.0).
    AddLight(areaLight).
    Build()
```

```go
// Custom Random Scene
config := rt.DefaultSceneConfig()
config.LambertProb = 0.5
config.MetalProb = 0.3
config.DielectricProb = 0.2
config.SphereGridBounds.MinA = -5
config.SphereGridBounds.MaxA = 5
world := rt.RandomSceneWithConfig(config)
```

## Implementation Status

**Ray Tracing in One Weekend:**

- [x] PNG output
- [x] Vec3/ray math
- [x] Sphere rendering
- [x] Surface normals
- [x] Anti-aliasing
- [x] Diffuse materials
- [x] Metal materials
- [x] Dielectric materials
- [x] Positionable camera
- [x] Depth of field

**Ray Tracing: The Next Week:**

- [x] Motion blur (object + camera)
- [x] BVH acceleration
- [x] Texture system (solid, checker, image)
- [x] Perlin noise
- [x] Quadrilaterals
- [x] Lights (emissive materials + NEE)
- [x] Cornell Box scene
- [x] Instances (translation/rotation/scale)
- [ ] Volumes (fog/smoke)

**Additional:**

- [x] Progressive rendering w/ Ebiten
- [x] Infinite Plane primitive
- [x] Scene configuration system
- [x] Builder pattern API for camera
- [x] Triangle primitive
- [x] Circle/Disk primitive
- [x] Compound primitives (Box, Pyramid)
- [x] Transform system with SRT ordering
- [x] Next Event Estimation (NEE) for direct lighting
- [x] Preset scenes with cameras

## Resources

- [Ray Tracing in One Weekend](https://raytracing.github.io/books/RayTracingInOneWeekend.html)
- [Ray Tracing: The Next Week](https://raytracing.github.io/books/RayTracingTheNextWeek.html)

## License

Educational purposes, following public domain tutorial series.
