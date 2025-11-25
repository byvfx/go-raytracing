# BIF Migration Guide: From Go Raytracer to Production Scene Assembler

## Current Implementation Review

Your Go raytracer has implemented the following features that will migrate to BIF:

### Core Math & Geometry
- Vec3/Point3/Color operations
- Ray class with time support (motion blur)
- Interval and AABB structures
- Transform system with SRT ordering

### Primitives (All with BVH support)
- Sphere (static and moving)
- Infinite plane
- Quad (axis-aligned)
- Triangle (Möller-Trumbore intersection)
- Circle/Disk
- Compound primitives (Box, Pyramid)

### Materials
- Lambertian (diffuse)
- Metal (reflective with fuzz)
- Dielectric (glass with Schlick approximation)
- DiffuseLight (emissive for area lights)

### Textures
- SolidColor
- CheckerTexture (3D procedural)
- ImageTexture (PNG/JPEG loading)
- NoiseTexture (Perlin noise with turbulence)

### Camera System
- Positionable camera with look-at
- Depth of field (defocus blur)
- Motion blur
- Builder pattern API
- Next Event Estimation (NEE) for direct lighting
- Multiple presets

### Rendering Features
- BVH acceleration structure
- Progressive scanline rendering
- Anti-aliasing via multi-sampling
- Gamma correction (gamma 2.0)
- Multiple pre-built scenes

## Migration Strategy

### Phase 1: Direct Port (Week 1-2)

Port your Go raytracer to Rust with minimal changes. This validates the Rust toolchain and gets you familiar with the language.

#### 1.1 Math Library Port

```rust
// crates/math/src/vec3.rs
use std::ops::{Add, Sub, Mul, Div, Neg};

#[derive(Debug, Clone, Copy, PartialEq)]
pub struct Vec3 {
    pub x: f32,
    pub y: f32,
    pub z: f32,
}

pub type Point3 = Vec3;
pub type Color = Vec3;

impl Vec3 {
    pub const ZERO: Self = Self { x: 0.0, y: 0.0, z: 0.0 };
    pub const ONE: Self = Self { x: 1.0, y: 1.0, z: 1.0 };
    
    pub fn new(x: f32, y: f32, z: f32) -> Self {
        Self { x, y, z }
    }
    
    pub fn dot(&self, other: Self) -> f32 {
        self.x * other.x + self.y * other.y + self.z * other.z
    }
    
    pub fn cross(&self, other: Self) -> Self {
        Self {
            x: self.y * other.z - self.z * other.y,
            y: self.z * other.x - self.x * other.z,
            z: self.x * other.y - self.y * other.x,
        }
    }
    
    pub fn length(&self) -> f32 {
        self.length_squared().sqrt()
    }
    
    pub fn length_squared(&self) -> f32 {
        self.x * self.x + self.y * self.y + self.z * self.z
    }
    
    pub fn unit(&self) -> Self {
        *self / self.length()
    }
    
    pub fn near_zero(&self) -> bool {
        const S: f32 = 1e-8;
        self.x.abs() < S && self.y.abs() < S && self.z.abs() < S
    }
}

// Implement ops traits for clean math
impl Add for Vec3 {
    type Output = Self;
    fn add(self, other: Self) -> Self {
        Self::new(self.x + other.x, self.y + other.y, self.z + other.z)
    }
}

impl Sub for Vec3 {
    type Output = Self;
    fn sub(self, other: Self) -> Self {
        Self::new(self.x - other.x, self.y - other.y, self.z - other.z)
    }
}

impl Mul<f32> for Vec3 {
    type Output = Self;
    fn mul(self, t: f32) -> Self {
        Self::new(self.x * t, self.y * t, self.z * t)
    }
}

impl Div<f32> for Vec3 {
    type Output = Self;
    fn div(self, t: f32) -> Self {
        self * (1.0 / t)
    }
}

impl Neg for Vec3 {
    type Output = Self;
    fn neg(self) -> Self {
        Self::new(-self.x, -self.y, -self.z)
    }
}
```

#### 1.2 Ray and Interval

```rust
// crates/math/src/ray.rs
use crate::vec3::{Vec3, Point3};

#[derive(Debug, Clone, Copy)]
pub struct Ray {
    pub origin: Point3,
    pub direction: Vec3,
    pub time: f32,
}

impl Ray {
    pub fn new(origin: Point3, direction: Vec3, time: f32) -> Self {
        Self { origin, direction, time }
    }
    
    pub fn at(&self, t: f32) -> Point3 {
        self.origin + self.direction * t
    }
}

// crates/math/src/interval.rs
#[derive(Debug, Clone, Copy)]
pub struct Interval {
    pub min: f32,
    pub max: f32,
}

impl Interval {
    pub const EMPTY: Self = Self { min: f32::INFINITY, max: f32::NEG_INFINITY };
    pub const UNIVERSE: Self = Self { min: f32::NEG_INFINITY, max: f32::INFINITY };
    
    pub fn new(min: f32, max: f32) -> Self {
        Self { min, max }
    }
    
    pub fn contains(&self, x: f32) -> bool {
        self.min <= x && x <= self.max
    }
    
    pub fn surrounds(&self, x: f32) -> bool {
        self.min < x && x < self.max
    }
    
    pub fn clamp(&self, x: f32) -> f32 {
        x.clamp(self.min, self.max)
    }
}
```

#### 1.3 Materials Trait System

```rust
// crates/renderer/src/material.rs
use crate::{Ray, HitRecord, Color, Vec3};

pub trait Material: Send + Sync {
    fn scatter(
        &self,
        ray_in: &Ray,
        rec: &HitRecord,
        attenuation: &mut Color,
        scattered: &mut Ray,
    ) -> bool;
    
    fn emitted(&self, u: f32, v: f32, p: Point3) -> Color {
        Color::ZERO  // Default: no emission
    }
}

pub struct Lambertian {
    pub albedo: Box<dyn Texture>,
}

impl Material for Lambertian {
    fn scatter(
        &self,
        ray_in: &Ray,
        rec: &HitRecord,
        attenuation: &mut Color,
        scattered: &mut Ray,
    ) -> bool {
        let mut scatter_direction = rec.normal + Vec3::random_unit_vector();
        
        if scatter_direction.near_zero() {
            scatter_direction = rec.normal;
        }
        
        *scattered = Ray::new(rec.p, scatter_direction, ray_in.time);
        *attenuation = self.albedo.value(rec.u, rec.v, rec.p);
        true
    }
}

pub struct Metal {
    pub albedo: Color,
    pub fuzz: f32,
}

impl Material for Metal {
    fn scatter(
        &self,
        ray_in: &Ray,
        rec: &HitRecord,
        attenuation: &mut Color,
        scattered: &mut Ray,
    ) -> bool {
        let reflected = Vec3::reflect(ray_in.direction.unit(), rec.normal);
        let reflected = reflected + Vec3::random_unit_vector() * self.fuzz;
        *scattered = Ray::new(rec.p, reflected, ray_in.time);
        *attenuation = self.albedo;
        scattered.direction.dot(rec.normal) > 0.0
    }
}

pub struct Dielectric {
    pub refraction_index: f32,
}

impl Material for Dielectric {
    fn scatter(
        &self,
        ray_in: &Ray,
        rec: &HitRecord,
        attenuation: &mut Color,
        scattered: &mut Ray,
    ) -> bool {
        *attenuation = Color::ONE;
        
        let ri = if rec.front_face {
            1.0 / self.refraction_index
        } else {
            self.refraction_index
        };
        
        let unit_direction = ray_in.direction.unit();
        let cos_theta = (-unit_direction).dot(rec.normal).min(1.0);
        let sin_theta = (1.0 - cos_theta * cos_theta).sqrt();
        
        let cannot_refract = ri * sin_theta > 1.0;
        let direction = if cannot_refract || reflectance(cos_theta, ri) > rand::random() {
            Vec3::reflect(unit_direction, rec.normal)
        } else {
            Vec3::refract(unit_direction, rec.normal, ri)
        };
        
        *scattered = Ray::new(rec.p, direction, ray_in.time);
        true
    }
}

fn reflectance(cosine: f32, ref_idx: f32) -> f32 {
    // Schlick's approximation
    let r0 = ((1.0 - ref_idx) / (1.0 + ref_idx)).powi(2);
    r0 + (1.0 - r0) * (1.0 - cosine).powi(5)
}
```

### Phase 2: Scene System Integration (Week 3-4)

Integrate your renderer with BIF's instance-based scene system.

#### 2.1 Hittable Trait

```rust
// crates/renderer/src/hittable.rs
use crate::{Ray, Interval, AABB, Material};
use std::sync::Arc;

pub struct HitRecord {
    pub p: Point3,
    pub normal: Vec3,
    pub mat: Arc<dyn Material>,
    pub t: f32,
    pub u: f32,
    pub v: f32,
    pub front_face: bool,
}

impl HitRecord {
    pub fn set_face_normal(&mut self, ray: &Ray, outward_normal: Vec3) {
        self.front_face = ray.direction.dot(outward_normal) < 0.0;
        self.normal = if self.front_face {
            outward_normal
        } else {
            -outward_normal
        };
    }
}

pub trait Hittable: Send + Sync {
    fn hit(&self, ray: &Ray, ray_t: Interval, rec: &mut HitRecord) -> bool;
    fn bounding_box(&self) -> AABB;
}
```

#### 2.2 BVH Implementation

```rust
// crates/renderer/src/bvh.rs
use crate::{Hittable, HitRecord, Ray, Interval, AABB};
use std::sync::Arc;

pub struct BvhNode {
    left: Arc<dyn Hittable>,
    right: Arc<dyn Hittable>,
    bbox: AABB,
}

impl BvhNode {
    pub fn new(objects: &mut [Arc<dyn Hittable>], time0: f32, time1: f32) -> Self {
        let axis = rand::random::<usize>() % 3;
        let comparator = match axis {
            0 => box_x_compare,
            1 => box_y_compare,
            _ => box_z_compare,
        };
        
        let (left, right) = match objects.len() {
            1 => (objects[0].clone(), objects[0].clone()),
            2 => {
                if comparator(&objects[0], &objects[1]) {
                    (objects[0].clone(), objects[1].clone())
                } else {
                    (objects[1].clone(), objects[0].clone())
                }
            }
            _ => {
                objects.sort_by(|a, b| {
                    if comparator(a, b) {
                        std::cmp::Ordering::Less
                    } else {
                        std::cmp::Ordering::Greater
                    }
                });
                
                let mid = objects.len() / 2;
                (
                    Arc::new(BvhNode::new(&mut objects[..mid], time0, time1)) as Arc<dyn Hittable>,
                    Arc::new(BvhNode::new(&mut objects[mid..], time0, time1)) as Arc<dyn Hittable>,
                )
            }
        };
        
        let box_left = left.bounding_box();
        let box_right = right.bounding_box();
        
        Self {
            left,
            right,
            bbox: AABB::surrounding_box(box_left, box_right),
        }
    }
}

impl Hittable for BvhNode {
    fn hit(&self, ray: &Ray, ray_t: Interval, rec: &mut HitRecord) -> bool {
        if !self.bbox.hit(ray, ray_t) {
            return false;
        }
        
        let hit_left = self.left.hit(ray, ray_t, rec);
        let hit_right = self.right.hit(
            ray,
            Interval::new(ray_t.min, if hit_left { rec.t } else { ray_t.max }),
            rec,
        );
        
        hit_left || hit_right
    }
    
    fn bounding_box(&self) -> AABB {
        self.bbox
    }
}

fn box_compare(a: &Arc<dyn Hittable>, b: &Arc<dyn Hittable>, axis: usize) -> bool {
    let box_a = a.bounding_box();
    let box_b = b.bounding_box();
    
    box_a.axis_interval(axis).min < box_b.axis_interval(axis).min
}

fn box_x_compare(a: &Arc<dyn Hittable>, b: &Arc<dyn Hittable>) -> bool { box_compare(a, b, 0) }
fn box_y_compare(a: &Arc<dyn Hittable>, b: &Arc<dyn Hittable>) -> bool { box_compare(a, b, 1) }
fn box_z_compare(a: &Arc<dyn Hittable>, b: &Arc<dyn Hittable>) -> bool { box_compare(a, b, 2) }
```

### Phase 3: BIF Integration (Week 5-6)

Connect your renderer to BIF's instance system and USD pipeline.

#### 3.1 Instance-Aware Renderer

```rust
// crates/renderer/src/instance_renderer.rs
use crate::{Ray, HitRecord, Scene, Camera};
use rayon::prelude::*;

pub struct InstanceRenderer {
    camera: Camera,
    samples_per_pixel: u32,
    max_depth: u32,
}

impl InstanceRenderer {
    pub fn render(&self, scene: &Scene, width: u32, height: u32) -> image::RgbImage {
        let mut img = image::RgbImage::new(width, height);
        
        img.enumerate_pixels_mut()
            .par_bridge()
            .for_each(|(x, y, pixel)| {
                let mut pixel_color = Color::ZERO;
                
                for _ in 0..self.samples_per_pixel {
                    let ray = self.camera.get_ray(x, y, width, height);
                    pixel_color += self.ray_color(&ray, scene, self.max_depth);
                }
                
                pixel_color /= self.samples_per_pixel as f32;
                
                // Apply gamma correction
                pixel_color = Color::new(
                    pixel_color.x.sqrt(),
                    pixel_color.y.sqrt(),
                    pixel_color.z.sqrt(),
                );
                
                *pixel = image::Rgb([
                    (pixel_color.x.clamp(0.0, 0.999) * 255.99) as u8,
                    (pixel_color.y.clamp(0.0, 0.999) * 255.99) as u8,
                    (pixel_color.z.clamp(0.0, 0.999) * 255.99) as u8,
                ]);
            });
        
        img
    }
    
    fn ray_color(&self, ray: &Ray, scene: &Scene, depth: u32) -> Color {
        if depth == 0 {
            return Color::ZERO;
        }
        
        let mut rec = HitRecord::default();
        
        // Check instances instead of raw geometry
        if let Some(hit) = scene.intersect_instances(ray, Interval::new(0.001, f32::INFINITY)) {
            let mut attenuation = Color::ZERO;
            let mut scattered = Ray::default();
            
            let emitted = hit.mat.emitted(hit.u, hit.v, hit.p);
            
            if !hit.mat.scatter(ray, &hit, &mut attenuation, &mut scattered) {
                return emitted;
            }
            
            // Next Event Estimation for lights
            let direct_light = if self.camera.lights.len() > 0 {
                self.sample_lights(&hit, scene)
            } else {
                Color::ZERO
            };
            
            let indirect = attenuation * self.ray_color(&scattered, scene, depth - 1);
            
            return emitted + direct_light + indirect;
        }
        
        // Sky gradient
        let unit_direction = ray.direction.unit();
        let t = 0.5 * (unit_direction.y + 1.0);
        Color::ONE * (1.0 - t) + Color::new(0.5, 0.7, 1.0) * t
    }
    
    fn sample_lights(&self, hit: &HitRecord, scene: &Scene) -> Color {
        // NEE implementation from your Go code
        // Sample area lights and compute contribution
        Color::ZERO  // Placeholder
    }
}
```

### Phase 4: Production Features (Week 7-8)

Add production features that differentiate BIF from your prototype.

#### 4.1 Progressive Rendering

```rust
// crates/renderer/src/progressive.rs
use std::sync::{Arc, Mutex};
use std::sync::atomic::{AtomicBool, AtomicU32, Ordering};

pub struct ProgressiveRenderer {
    accumulator: Arc<Mutex<Vec<Color>>>,
    sample_count: AtomicU32,
    dirty: AtomicBool,
    width: u32,
    height: u32,
}

impl ProgressiveRenderer {
    pub fn new(width: u32, height: u32) -> Self {
        let pixel_count = (width * height) as usize;
        Self {
            accumulator: Arc::new(Mutex::new(vec![Color::ZERO; pixel_count])),
            sample_count: AtomicU32::new(0),
            dirty: AtomicBool::new(true),
            width,
            height,
        }
    }
    
    pub fn add_sample(&self, scene: &Scene, camera: &Camera) {
        let samples = self.sample_count.fetch_add(1, Ordering::SeqCst) + 1;
        
        (0..self.height).into_par_iter().for_each(|y| {
            for x in 0..self.width {
                let ray = camera.get_ray(x, y, self.width, self.height);
                let color = self.ray_color(&ray, scene, 50);
                
                let idx = (y * self.width + x) as usize;
                let mut acc = self.accumulator.lock().unwrap();
                
                // Incremental average
                let old_avg = acc[idx];
                let new_avg = old_avg + (color - old_avg) / samples as f32;
                acc[idx] = new_avg;
            }
        });
        
        self.dirty.store(true, Ordering::SeqCst);
    }
    
    pub fn get_image(&self) -> image::RgbImage {
        let acc = self.accumulator.lock().unwrap();
        let mut img = image::RgbImage::new(self.width, self.height);
        
        for (i, pixel) in img.pixels_mut().enumerate() {
            let color = acc[i];
            // Apply gamma correction
            *pixel = image::Rgb([
                (color.x.sqrt().clamp(0.0, 0.999) * 255.99) as u8,
                (color.y.sqrt().clamp(0.0, 0.999) * 255.99) as u8,
                (color.z.sqrt().clamp(0.0, 0.999) * 255.99) as u8,
            ]);
        }
        
        img
    }
    
    pub fn reset(&self) {
        let mut acc = self.accumulator.lock().unwrap();
        acc.fill(Color::ZERO);
        self.sample_count.store(0, Ordering::SeqCst);
        self.dirty.store(true, Ordering::SeqCst);
    }
}
```

#### 4.2 Texture System

```rust
// crates/renderer/src/texture.rs
use image::GenericImageView;
use std::sync::Arc;

pub trait Texture: Send + Sync {
    fn value(&self, u: f32, v: f32, p: Point3) -> Color;
}

pub struct SolidColor {
    color: Color,
}

impl Texture for SolidColor {
    fn value(&self, _u: f32, _v: f32, _p: Point3) -> Color {
        self.color
    }
}

pub struct CheckerTexture {
    inv_scale: f32,
    even: Arc<dyn Texture>,
    odd: Arc<dyn Texture>,
}

impl Texture for CheckerTexture {
    fn value(&self, u: f32, v: f32, p: Point3) -> Color {
        let x_int = (self.inv_scale * p.x).floor() as i32;
        let y_int = (self.inv_scale * p.y).floor() as i32;
        let z_int = (self.inv_scale * p.z).floor() as i32;
        
        let is_even = (x_int + y_int + z_int) % 2 == 0;
        
        if is_even {
            self.even.value(u, v, p)
        } else {
            self.odd.value(u, v, p)
        }
    }
}

pub struct ImageTexture {
    data: Vec<u8>,
    width: u32,
    height: u32,
}

impl ImageTexture {
    pub fn new(filename: &str) -> Result<Self, image::ImageError> {
        let img = image::open(filename)?;
        let (width, height) = img.dimensions();
        let data = img.to_rgb8().into_raw();
        
        Ok(Self { data, width, height })
    }
}

impl Texture for ImageTexture {
    fn value(&self, u: f32, v: f32, _p: Point3) -> Color {
        let u = u.clamp(0.0, 1.0);
        let v = 1.0 - v.clamp(0.0, 1.0);  // Flip V
        
        let i = ((u * self.width as f32) as u32).min(self.width - 1);
        let j = ((v * self.height as f32) as u32).min(self.height - 1);
        
        let idx = ((j * self.width + i) * 3) as usize;
        
        Color::new(
            self.data[idx] as f32 / 255.0,
            self.data[idx + 1] as f32 / 255.0,
            self.data[idx + 2] as f32 / 255.0,
        )
    }
}

pub struct NoiseTexture {
    noise: Perlin,
    scale: f32,
}

impl Texture for NoiseTexture {
    fn value(&self, _u: f32, _v: f32, p: Point3) -> Color {
        // Turbulence pattern from your Go implementation
        let s = self.scale * p.z + 10.0 * self.noise.turb(p * self.scale, 7);
        let turb_value = 0.5 * (1.0 + s.sin());
        Color::ONE * turb_value
    }
}
```

## Migration Timeline

### Week 1-2: Core Port
- [ ] Port math library (Vec3, Ray, Interval, AABB)
- [ ] Port materials system
- [ ] Port primitive shapes
- [ ] Basic renderer working

### Week 3-4: Scene Integration
- [ ] Connect to BIF scene graph
- [ ] Instance-aware rendering
- [ ] BVH for instances
- [ ] Layer system support

### Week 5-6: USD Pipeline
- [ ] Export rendered scenes to USD
- [ ] Import USD for rendering
- [ ] Material binding from USD
- [ ] PointInstancer support

### Week 7-8: Production Features
- [ ] Progressive rendering
- [ ] Viewport integration
- [ ] Texture TX pipeline
- [ ] Memory-efficient instancing

### Week 9-10: GPU Acceleration
- [ ] Hybrid CPU/GPU rendering
- [ ] Wavefront path tracing
- [ ] Denoising integration
- [ ] Performance optimization

## Key Differences from Go Implementation

### Memory Management
- Rust uses `Arc<T>` instead of Go pointers for shared ownership
- No garbage collector - explicit memory management
- Use `Box<dyn Trait>` for trait objects

### Parallelism
- Use `rayon` for parallel iteration (simpler than Go routines)
- `std::sync` primitives for synchronization
- Fearless concurrency - compiler enforces thread safety

### Error Handling
- Use `Result<T, E>` instead of multiple return values
- `?` operator for error propagation
- No nil/null - use `Option<T>`

### Performance Optimizations
- Zero-cost abstractions
- SIMD via `packed_simd` or manual intrinsics
- Compile-time optimizations more aggressive than Go

## Integration with BIF Architecture

Your renderer becomes the core of BIF's rendering pipeline:

1. **Scene Graph**: BIF manages instances, your renderer traces them
2. **Materials**: Extend your material system with MaterialX support
3. **Textures**: Add TX/mipmap support via OIIO
4. **Camera**: Your camera system drives both viewport and production
5. **Acceleration**: Your BVH enhanced with instance-aware traversal

## Next Steps

1. Start with Phase 1 - direct port to Rust
2. Get comfortable with Rust's ownership model
3. Gradually integrate BIF's scene features
4. Add production features incrementally

The migration preserves all your hard work while building a production-ready foundation around it.
