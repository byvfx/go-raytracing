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

## Features

### Camera System

- **Positionable camera**: Place the camera anywhere in 3D space using `LookFrom`
- **Look-at targeting**: Point the camera at any location using `LookAt`
- **Adjustable field of view**: Control zoom level with `Vfov` (vertical field of view in degrees)
- **Camera orientation**: Define "up" direction with `Vup` vector
- **Aspect ratio control**: Set image dimensions with `AspectRatio` and `ImageWidth`

Example camera configuration:

```go
camera.Vfov = 20                                   // Telephoto lens (zoomed in)
camera.LookFrom = rt.Point3{X: -2, Y: 2, Z: 1}    // Camera position
camera.LookAt = rt.Point3{X: 0, Y: 0, Z: -1}      // Looking at origin
camera.Vup = rt.Vec3{X: 0, Y: 1, Z: 0}            // Y-axis is up
```

### Materials

- **Lambertian (Diffuse)**: Matte surfaces that scatter light randomly
- **Metal**: Reflective surfaces with adjustable fuzziness
- **Dielectric (Glass)**: Transparent materials with refraction and reflection
  - Supports realistic Fresnel effects (Schlick's approximation)
  - Can create hollow glass spheres (bubble effect)

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

You can easily modify the scene in `main.go`:

```go
// Create custom materials
materialGlass := rt.NewDielectric(1.5)                          // Glass (refractive index 1.5)
materialMetal := rt.NewMetal(rt.Color{X: 0.8, Y: 0.6, Z: 0.2}, 0.0)  // Shiny gold metal
materialMatte := rt.NewLambertian(rt.Color{X: 0.7, Y: 0.3, Z: 0.3})  // Red matte

// Add spheres to the scene
world.Add(rt.NewSphere(rt.Point3{X: 0, Y: 0, Z: -1}, 0.5, materialGlass))

// Adjust camera settings
camera.Vfov = 90                                    // Wide angle
camera.SamplesPerPixel = 500                        // High quality
camera.MaxDepth = 50                                // More light bounces
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
- [ ] Bounding Volume Hiearchies (BVH) 
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
