# Rust Port Checklist

## Prerequisites

- [ ] Set up Rust project with cargo
- [ ] Add dependencies (see below)
- [ ] Decide on rendering architecture (offline vs interactive)

## Recommended Crates

```toml
[dependencies]
# Math and vectors
glam = "0.29"              # SIMD-optimized vector math (replaces Vec3)

# Parallelism
rayon = "1.10"             # Data parallelism (replaces worker pool)

# Image I/O
image = "0.25"             # PNG/JPG encoding and decoding

# 3D file loading
tobj = "4.0"               # OBJ file parsing (preferred over obj crate)

# Random numbers
rand = "0.8"               # RNG (replaces Go's math/rand)

# Optional for interactive rendering
winit = "0.30"             # Window creation
wgpu = "23.0"              # GPU rendering for display
pollster = "0.3"           # Async executor

# Utilities
anyhow = "1.0"             # Error handling
thiserror = "2.0"          # Custom error types
```

> **Note:** Versions updated to match `bif_poc_guide.md` workspace dependencies (as of Nov 2025).

## How to Use This Checklist

- Pair this file with `rust_port_learning_plan.md`. Each stage in the plan references the sections below so you always know *when* to tackle a block of work.
- Quick stage mapping:
    - **Stage 0** – Workspace bring-up: see `bif_poc_guide.md` Steps 1-7 (no checklist sections yet).
    - **Stage 1** – Core math types: Sections 1-4.
    - **Stage 2** – Traits & primitives: Sections 5-11.
    - **Stage 3** – Materials & textures: Sections 7, 13-16.
    - **Stage 4** – BVH, camera, CPU renderer: Sections 12, 17-19. (framebuffer crate deferred to post-PoC)
    - **Stage 5** – Scene graph, IO: Sections 18-21, 32.
    - **Stage 6** – USD/tooling/testing: Sections 20-26 + `usd_bridge` crate (`bif_poc_guide.md` Step 4).
    - **Stage 7+** – Optimizations, GPU, advanced features: Sections 22-31 plus Advanced Features.
- As you complete a stage, check off the corresponding items here and capture metrics/screenshots in `devlog/` to keep the documentation synchronized.

> **Crate Reference:** The PoC creates these crates: `app`, `scene`, `renderer`, `viewport`, `io_gltf`, `usd_bridge`. A `framebuffer` crate is added post-PoC during Stage 4 when progressive refinement is implemented.

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

### 11b. Compound Primitives

- [ ] `Box` primitive (6 quads)
- [ ] `Pyramid` primitive (base quad + 4 triangles)
- [ ] `Circle`/`Disk` primitive

### 11c. Transform System

- [ ] `Translate` wrapper with AABB update
- [ ] `RotateX`, `RotateY`, `RotateZ` wrappers
- [ ] `Scale` (non-uniform) with inverse factor caching
- [ ] `Transform` builder (SRT ordering: Scale → Rotate → Translate)

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
- [ ] `BVHLeaf` for multiple primitives per leaf (cache locality)
- [ ] Parallel BVH construction with `rayon` (see Go's bvhSemaphore pattern)
- [ ] Pre-compute centroids for faster sorting
- [ ] Configurable leaf size (Go uses `bvhLeafMaxSize = 4`)
- [ ] Parallel threshold (Go uses `bvhParallelThreshold = 8192`)

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

- [ ] Builder pattern for configuration (Go's `NewCameraBuilder()`)
- [ ] Lens parameters (FOV, focus, aperture)
- [ ] Motion blur support (camera position interpolation)
- [ ] Ray generation with DOF (defocus disk sampling)
- [ ] Free camera mode (forward vector instead of look-at)
- [ ] Camera presets: `QuickPreview`, `StandardQuality`, `HighQuality`
- [ ] Background color options (solid, sky gradient)
- [ ] Light collection for NEE (`AddLight()` method)

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
- [ ] Spiral bucket ordering (V-Ray style, center-out)
- [ ] Progressive multi-pass rendering (Preview → Refine → Final)
- [ ] Thread-safe framebuffer updates with `Mutex`
- [ ] Live viewport display with `egui`/`ebiten` equivalent
- [ ] Configurable pass quality (SPP/depth per pass)
- [ ] Per-bucket temp buffer to minimize lock contention
- [ ] Atomic counters for progress tracking

### 19. MIS Implementation

- [ ] Light sampling (NEE - Next Event Estimation)
- [ ] BRDF sampling
- [ ] Balance heuristic weighting: `w = pdf_light / (pdf_light + pdf_brdf)`
- [ ] PDF evaluation infrastructure
- [ ] `PDFEvaluator` trait: `fn pdf(wi, wo, normal) -> f64`
- [ ] `MaterialInfo` trait for material properties (is_specular, is_emissive, can_use_nee)
- [ ] Firefly clamping (max component = 20.0)

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

### 29. UI (Required for PoC)

#### egui Phase (PoC)

- [ ] Basic egui + wgpu window setup
- [ ] 3D viewport panel (wgpu texture display)
- [ ] Scene hierarchy tree view
- [ ] Properties panel (transform editor, material selector)
- [ ] Scatter controls panel (density, scale variance, seed)
- [ ] Render progress overlay (samples, time, pass)
- [ ] Camera controls (orbit, pan, zoom via mouse)

#### Node Editor (PoC or Early Production)

- [ ] Integrate `egui_node_graph` or similar crate
- [ ] Surface Input node (mesh selection)
- [ ] Point Sampler node (uniform, density-based)
- [ ] Filter node (slope, height, random cull)
- [ ] Instance Placer node (prototype, scale/rotation variance)
- [ ] Real-time preview of scatter points in viewport
- [ ] Node graph serialization (save/load presets)

#### Qt Migration (Production)

- [ ] Qt 6 / QML frontend via `cxx-qt`
- [ ] Native Qt viewport with embedded wgpu
- [ ] Qt-based node editor (QGraphicsScene or QtNodes)
- [ ] Dockable panels layout
- [ ] Keyboard shortcuts and professional UX
- [ ] Theme/styling system

### 29b. Viewport as Framebuffer

- [ ] Renderer writes to wgpu texture directly
- [ ] Viewport displays same texture (no copy)
- [ ] UI overlays render on top (gizmos, selection)
- [ ] Mode switching: interactive ↔ progressive ↔ final
- [ ] Snapshot/pin current render for comparison
- [ ] Progress bar and sample counter in viewport

### 29c. Scatter System (`crates/scatter/`)

- [ ] Surface point generation (uniform distribution)
- [ ] Density map support (texture-driven density)
- [ ] Point filtering (slope threshold, height range)
- [ ] Instance placement from point cloud
- [ ] Scale/rotation randomization with seed control
- [ ] Preview mode (show points before instancing)
- [ ] Batch processing for large point counts
- [ ] Integration with scene graph (creates instances)

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

```text
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
