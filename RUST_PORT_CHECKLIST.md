# Rust Port Checklist

## Prerequisites
- [ ] Set up Rust project with cargo
- [ ] Add dependencies (see below)
- [ ] Decide on rendering architecture (offline vs interactive)

## Recommended Crates
```toml
[dependencies]
# Math and vectors
glam = "0.24"              # SIMD-optimized vector math (replaces Vec3)

# Parallelism
rayon = "1.8"              # Data parallelism (replaces worker pool)

# Image I/O
image = "0.24"             # PNG/JPG encoding and decoding

# 3D file loading
obj = "0.10"               # OBJ file parsing
# OR
tobj = "4.0"               # Alternative OBJ parser

# Random numbers
rand = "0.8"               # RNG (replaces Go's math/rand)

# Optional for interactive rendering
winit = "0.29"             # Window creation
wgpu = "0.18"              # GPU rendering for display
pollster = "0.3"           # Async executor

# Utilities
anyhow = "1.0"             # Error handling
```

## Core Types (High Priority)

### 1. Vector Math
- [ ] `Vec3` type using `glam::Vec3A` (SIMD-aligned)
- [ ] `Point3` type alias: `type Point3 = Vec3;`
- [ ] `Color` type alias: `type Color = Vec3;`
- [ ] Helper functions: `dot()`, `cross()`, `reflect()`, `refract()`

### 2. Ray
- [ ] `Ray` struct with origin, direction, time
- [ ] `at(t: f64) -> Point3` method

### 3. Interval
- [ ] `Interval` struct with min/max
- [ ] Methods: `contains()`, `clamp()`, `expand()`, `size()`

### 4. AABB (Axis-Aligned Bounding Box)
- [ ] `AABB` struct with x, y, z intervals
- [ ] `hit()` method for ray-box intersection
- [ ] `longest_axis()` method
- [ ] Combining AABBs

## Traits (Interfaces)

### 5. Hittable Trait
```rust
trait Hittable: Send + Sync {
    fn hit(&self, r: &Ray, ray_t: Interval, rec: &mut HitRecord) -> bool;
    fn bounding_box(&self) -> AABB;
}
```
- [ ] Define trait
- [ ] Add `Send + Sync` for thread safety

### 6. Material Trait
```rust
trait Material: Send + Sync {
    fn scatter(&self, r_in: &Ray, rec: &HitRecord) 
        -> Option<(Color, Ray)>;
    fn emitted(&self, u: f64, v: f64, p: Point3) -> Color {
        Color::ZERO
    }
}
```
- [ ] Define trait
- [ ] Add optional PDF methods for MIS

### 7. Texture Trait
```rust
trait Texture: Send + Sync {
    fn value(&self, u: f64, v: f64, p: Point3) -> Color;
}
```
- [ ] Define trait

## Primitives

### 8. Triangle
- [ ] Implement `Hittable` for `Triangle`
- [ ] Möller-Trumbore intersection
- [ ] Pre-compute bounding box
- [ ] Consider storing edge vectors

### 9. Sphere
- [ ] Implement `Hittable` for `Sphere`
- [ ] Moving sphere support
- [ ] UV mapping

### 10. Quad
- [ ] Implement `Hittable` for `Quad`
- [ ] Point sampling for area lights
- [ ] PDF evaluation

### 11. Plane
- [ ] Implement `Hittable` for `Plane`
- [ ] Infinite plane intersection

## Acceleration Structures

### 12. BVH
```rust
enum BVHNode {
    Leaf {
        object: Box<dyn Hittable>,
        bbox: AABB,
    },
    Branch {
        left: Box<BVHNode>,
        right: Box<BVHNode>,
        bbox: AABB,
    },
}
```
- [ ] Recursive construction with longest axis heuristic
- [ ] Consider SAH (Surface Area Heuristic) for production
- [ ] Implement `Hittable` for `BVHNode`

## Materials

### 13. Lambertian (Diffuse)
- [ ] Cosine-weighted hemisphere sampling
- [ ] Texture support
- [ ] PDF evaluation for MIS

### 14. Metal
- [ ] Perfect reflection
- [ ] Fuzz/roughness parameter
- [ ] PDF evaluation for MIS

### 15. Dielectric (Glass)
- [ ] Fresnel equations
- [ ] Refraction with Snell's law
- [ ] Total internal reflection

### 16. DiffuseLight
- [ ] Emission-only material
- [ ] Texture support

## Camera & Rendering

### 17. Camera
- [ ] Builder pattern for configuration
- [ ] Lens parameters (FOV, focus, aperture)
- [ ] Motion blur support
- [ ] Ray generation with DOF

### 18. Bucket Renderer
```rust
struct BucketRenderer {
    camera: Camera,
    world: Arc<dyn Hittable>,
    buckets: Vec<Bucket>,
    framebuffer: Arc<Mutex<RgbaImage>>,
}
```
- [ ] Parallel bucket rendering with `rayon`
- [ ] Spiral bucket ordering
- [ ] Progressive multi-pass rendering
- [ ] Thread-safe framebuffer updates

### 19. MIS Implementation
- [ ] Light sampling
- [ ] BRDF sampling
- [ ] Balance heuristic weighting
- [ ] PDF evaluation infrastructure

## File I/O

### 20. OBJ Loader
- [ ] Parse OBJ files with `obj` or `tobj` crate
- [ ] Build BVH from triangles
- [ ] Transform support
- [ ] Material assignment

### 21. Image Export
- [ ] PNG export with `image` crate
- [ ] Gamma correction
- [ ] Tone mapping

## Optimizations

### 22. SIMD
- [ ] Use `glam` for automatic SIMD vectorization
- [ ] Ensure proper alignment with `Vec3A`
- [ ] Profile hot paths

### 23. Memory Layout
- [ ] Use `#[repr(C)]` for predictable layout
- [ ] Consider struct-of-arrays for triangle meshes
- [ ] Cache-friendly data structures

### 24. Profiling
- [ ] Add `--release` profile optimizations
- [ ] Profile with `cargo flamegraph`
- [ ] Identify bottlenecks

## Testing

### 25. Unit Tests
- [ ] Vector math tests
- [ ] Ray-primitive intersection tests
- [ ] Material scattering tests
- [ ] BVH construction tests

### 26. Integration Tests
- [ ] Render test scenes
- [ ] Compare output with Go version
- [ ] Validate MIS implementation

## Advanced Features (Future)

### 27. GPU Acceleration (Optional)
- [ ] OptiX backend for NVIDIA
- [ ] Embree for CPU ray tracing
- [ ] Vulkan compute shaders

### 28. Scene Format
- [ ] JSON/YAML scene description
- [ ] Material library
- [ ] Camera presets

### 29. UI (Optional)
- [ ] Real-time preview with accumulation
- [ ] Interactive camera controls
- [ ] Render settings UI

## Code Quality

### 30. Documentation
- [ ] Doc comments on public APIs
- [ ] Examples in doc tests
- [ ] Architecture overview

### 31. Error Handling
- [ ] Use `anyhow` or `thiserror`
- [ ] Proper error propagation
- [ ] Meaningful error messages

### 32. Code Organization
```
src/
├── main.rs
├── lib.rs
├── types/
│   ├── mod.rs
│   ├── vec3.rs
│   ├── ray.rs
│   └── interval.rs
├── geometry/
│   ├── mod.rs
│   ├── triangle.rs
│   ├── sphere.rs
│   └── quad.rs
├── acceleration/
│   ├── mod.rs
│   └── bvh.rs
├── materials/
│   ├── mod.rs
│   ├── lambertian.rs
│   ├── metal.rs
│   └── dielectric.rs
├── camera.rs
├── renderer/
│   ├── mod.rs
│   └── bucket.rs
└── io/
    ├── mod.rs
    └── obj.rs
```

## Performance Targets

- [ ] Match or exceed Go version performance
- [ ] Target: <10ms per bucket (32x32) on modern CPU
- [ ] Target: 4-8x speedup with multi-threading
- [ ] Target: <1s preview pass for 600x600 image

## Migration Strategy

1. **Phase 1**: Core types and traits (Vec3, Ray, Interval, AABB)
2. **Phase 2**: Basic primitives (Sphere, Plane)
3. **Phase 3**: Materials (Lambertian, Metal)
4. **Phase 4**: BVH acceleration
5. **Phase 5**: Camera and basic rendering
6. **Phase 6**: Bucket renderer and parallelism
7. **Phase 7**: Triangle and OBJ loading
8. **Phase 8**: MIS implementation
9. **Phase 9**: Progressive rendering
10. **Phase 10**: Optimization and profiling

## Notes

- Prefer `f32` over `f64` for better SIMD performance (test both)
- Use `#[inline]` on hot path functions (Vec3 operations, ray intersection)
- Consider `unsafe` for critical sections if profiling shows benefit
- Keep unsafe code minimal and well-documented
- Use `cargo clippy` for linting
- Use `cargo fmt` for consistent formatting
