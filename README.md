# Go Raytracing

A Go implementation following the "Ray Tracing in One Weekend" tutorial by Peter Shirley.

## Requirements

- Go 1.25.1 or later

## Building and Running

1. Clone or download this repository
2. Navigate to the project directory
3. Run the program:

```bash
go run main.go
```

The program will generate an `image.ppm` file in the same directory.

## Output

The current implementation generates a 256x256 pixel image with:

- Red channel: varies from 0 to 1 (left to right)
- Green channel: varies from 0 to 1 (bottom to top)  
- Blue channel: constant at 0

This creates a gradient from black (bottom-left) to yellow (top-right).

## Progress

This implementation follows the "Ray Tracing in One Weekend" tutorial progression:

- [x] Basic image output (PPM format)
- [x] Progress indicator
- [ ] Vector math utilities
- [ ] Ray class
- [ ] Simple sphere rendering
- [ ] Surface normals and shading
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
