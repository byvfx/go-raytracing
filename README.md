# Go Raytracing

A Go implementation following the "Ray Tracing in One Weekend" tutorial by Peter Shirley.

## Overview

This project implements a raytracer that generates PNG image files. The current implementation features a simple sphere with ray intersection testing rendered against a sky gradient background.

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

## Output

The current implementation generates a 800x450 pixel image (16:9 aspect ratio):

- A sphere positioned at (0, 0, -1) with radius 0.5
- A sphere positioned at (-1, 0.5, -1.5) with radius 0.5
- A sphere positioned at (1, 0.5, -1.5) with radius 0.5
- A larger ground sphere at (0, -100.5, -1) with radius 100
- Sky gradient background:

The image is saved as `image.png` in PNG format for easy viewing.

## Progress

This implementation follows the "Ray Tracing in One Weekend" tutorial progression:

- [x] Basic image output (PPM format)
- [x] Progress indicator
- [x] Vector math utilities
- [x] Ray class
- [x] Simple sphere rendering
- [x] Surface normals and shading
- [ ] Anti-aliasing
- [ ] Diffuse materials
- [ ] Metal materials
- [ ] Dielectric materials
- [ ] Camera positioning
- [ ] Depth of field

## Resources

- [Ray Tracing in One Weekend](https://raytracing.github.io/books/RayTracingInOneWeekend.html) - Original tutorial

## License

This project is for educational purposes following the public domain tutorial "Ray Tracing in One Weekend"
