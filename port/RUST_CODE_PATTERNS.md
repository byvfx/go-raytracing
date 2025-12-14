# Rust Code Patterns for Raytracer Port

Quick-reference code snippets showing idiomatic Rust patterns for common raytracer operations. Copy-paste ready.

---

## Project Setup

### Cargo.toml

```toml
[package]
name = "bif-raytracer"
version = "0.1.0"
edition = "2021"

[dependencies]
# Math (SIMD-optimized vectors)
glam = { version = "0.29", features = ["fast-math"] }

# Parallelism
rayon = "1.10"

# Image I/O
image = "0.25"

# 3D file loading
tobj = "4.0"

# Random numbers
rand = "0.8"

# Error handling
anyhow = "1.0"
thiserror = "2.0"

# Optional: Interactive rendering
winit = "0.30"
wgpu = "23.0"
pollster = "0.3"
egui = "0.29"
egui-wgpu = "0.29"

[profile.release]
lto = true
codegen-units = 1
opt-level = 3
```

---

## Core Types

### Vec3 (using glam)

```rust
use glam::Vec3A;

pub type Vec3 = Vec3A;
pub type Point3 = Vec3A;
pub type Color = Vec3A;

// Constants
const ZERO: Vec3 = Vec3::ZERO;
const ONE: Vec3 = Vec3::ONE;

// Operations (already implemented by glam)
let a = Vec3::new(1.0, 2.0, 3.0);
let b = Vec3::new(4.0, 5.0, 6.0);

let sum = a + b;
let diff = a - b;
let scaled = a * 2.0;
let dot = a.dot(b);
let cross = a.cross(b);
let len = a.length();
let unit = a.normalize();
let near_zero = a.abs_diff_eq(Vec3::ZERO, 1e-8);
```

### Custom Vec3 (if not using glam)

```rust
use std::ops::{Add, Sub, Mul, Div, Neg};

#[derive(Debug, Clone, Copy, PartialEq, Default)]
pub struct Vec3 {
    pub x: f32,
    pub y: f32,
    pub z: f32,
}

impl Vec3 {
    pub const ZERO: Self = Self { x: 0.0, y: 0.0, z: 0.0 };
    pub const ONE: Self = Self { x: 1.0, y: 1.0, z: 1.0 };

    #[inline]
    pub fn new(x: f32, y: f32, z: f32) -> Self {
        Self { x, y, z }
    }

    #[inline]
    pub fn dot(self, other: Self) -> f32 {
        self.x * other.x + self.y * other.y + self.z * other.z
    }

    #[inline]
    pub fn cross(self, other: Self) -> Self {
        Self {
            x: self.y * other.z - self.z * other.y,
            y: self.z * other.x - self.x * other.z,
            z: self.x * other.y - self.y * other.x,
        }
    }

    #[inline]
    pub fn length_squared(self) -> f32 {
        self.dot(self)
    }

    #[inline]
    pub fn length(self) -> f32 {
        self.length_squared().sqrt()
    }

    #[inline]
    pub fn normalize(self) -> Self {
        self / self.length()
    }

    #[inline]
    pub fn near_zero(self) -> bool {
        const S: f32 = 1e-8;
        self.x.abs() < S && self.y.abs() < S && self.z.abs() < S
    }
}

impl Add for Vec3 {
    type Output = Self;
    #[inline]
    fn add(self, rhs: Self) -> Self {
        Self::new(self.x + rhs.x, self.y + rhs.y, self.z + rhs.z)
    }
}

impl Sub for Vec3 {
    type Output = Self;
    #[inline]
    fn sub(self, rhs: Self) -> Self {
        Self::new(self.x - rhs.x, self.y - rhs.y, self.z - rhs.z)
    }
}

impl Mul<f32> for Vec3 {
    type Output = Self;
    #[inline]
    fn mul(self, t: f32) -> Self {
        Self::new(self.x * t, self.y * t, self.z * t)
    }
}

impl Mul<Vec3> for f32 {
    type Output = Vec3;
    #[inline]
    fn mul(self, v: Vec3) -> Vec3 {
        v * self
    }
}

impl Div<f32> for Vec3 {
    type Output = Self;
    #[inline]
    fn div(self, t: f32) -> Self {
        self * (1.0 / t)
    }
}

impl Neg for Vec3 {
    type Output = Self;
    #[inline]
    fn neg(self) -> Self {
        Self::new(-self.x, -self.y, -self.z)
    }
}
```

### Ray

```rust
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

    #[inline]
    pub fn at(&self, t: f32) -> Point3 {
        self.origin + self.direction * t
    }
}
```

### Interval

```rust
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

    pub fn from_intervals(a: Self, b: Self) -> Self {
        Self {
            min: a.min.min(b.min),
            max: a.max.max(b.max),
        }
    }

    #[inline]
    pub fn size(&self) -> f32 {
        self.max - self.min
    }

    #[inline]
    pub fn contains(&self, x: f32) -> bool {
        self.min <= x && x <= self.max
    }

    #[inline]
    pub fn surrounds(&self, x: f32) -> bool {
        self.min < x && x < self.max
    }

    #[inline]
    pub fn clamp(&self, x: f32) -> f32 {
        x.clamp(self.min, self.max)
    }

    pub fn expand(&self, delta: f32) -> Self {
        Self::new(self.min - delta, self.max + delta)
    }
}
```

### AABB

```rust
#[derive(Debug, Clone, Copy)]
pub struct AABB {
    pub x: Interval,
    pub y: Interval,
    pub z: Interval,
}

impl AABB {
    pub const EMPTY: Self = Self {
        x: Interval::EMPTY,
        y: Interval::EMPTY,
        z: Interval::EMPTY,
    };

    pub fn from_points(a: Point3, b: Point3) -> Self {
        let mut aabb = Self {
            x: Interval::new(a.x.min(b.x), a.x.max(b.x)),
            y: Interval::new(a.y.min(b.y), a.y.max(b.y)),
            z: Interval::new(a.z.min(b.z), a.z.max(b.z)),
        };
        aabb.pad_to_minimums();
        aabb
    }

    pub fn from_boxes(a: Self, b: Self) -> Self {
        Self {
            x: Interval::from_intervals(a.x, b.x),
            y: Interval::from_intervals(a.y, b.y),
            z: Interval::from_intervals(a.z, b.z),
        }
    }

    fn pad_to_minimums(&mut self) {
        const DELTA: f32 = 0.0001;
        if self.x.size() < DELTA { self.x = self.x.expand(DELTA); }
        if self.y.size() < DELTA { self.y = self.y.expand(DELTA); }
        if self.z.size() < DELTA { self.z = self.z.expand(DELTA); }
    }

    #[inline]
    pub fn axis(&self, n: usize) -> Interval {
        match n {
            1 => self.y,
            2 => self.z,
            _ => self.x,
        }
    }

    pub fn hit(&self, ray: &Ray, mut ray_t: Interval) -> bool {
        let origin = [ray.origin.x, ray.origin.y, ray.origin.z];
        let dir = [ray.direction.x, ray.direction.y, ray.direction.z];
        let axes = [self.x, self.y, self.z];

        for i in 0..3 {
            let inv_d = 1.0 / dir[i];
            let mut t0 = (axes[i].min - origin[i]) * inv_d;
            let mut t1 = (axes[i].max - origin[i]) * inv_d;

            if inv_d < 0.0 {
                std::mem::swap(&mut t0, &mut t1);
            }

            ray_t.min = ray_t.min.max(t0);
            ray_t.max = ray_t.max.min(t1);

            if ray_t.max <= ray_t.min {
                return false;
            }
        }
        true
    }

    pub fn longest_axis(&self) -> usize {
        let sizes = [self.x.size(), self.y.size(), self.z.size()];
        if sizes[0] > sizes[1] && sizes[0] > sizes[2] { 0 }
        else if sizes[1] > sizes[2] { 1 }
        else { 2 }
    }

    pub fn centroid(&self) -> Vec3 {
        Vec3::new(
            (self.x.min + self.x.max) * 0.5,
            (self.y.min + self.y.max) * 0.5,
            (self.z.min + self.z.max) * 0.5,
        )
    }
}
```

---

## Traits

### Hittable

```rust
use std::sync::Arc;

pub struct HitRecord {
    pub p: Point3,
    pub normal: Vec3,
    pub material: Arc<dyn Material>,
    pub t: f32,
    pub u: f32,
    pub v: f32,
    pub front_face: bool,
}

impl HitRecord {
    pub fn set_face_normal(&mut self, ray: &Ray, outward_normal: Vec3) {
        self.front_face = ray.direction.dot(outward_normal) < 0.0;
        self.normal = if self.front_face { outward_normal } else { -outward_normal };
    }
}

pub trait Hittable: Send + Sync {
    fn hit(&self, ray: &Ray, ray_t: Interval, rec: &mut HitRecord) -> bool;
    fn bounding_box(&self) -> AABB;
}
```

### Material

```rust
pub struct ScatterResult {
    pub attenuation: Color,
    pub scattered: Ray,
}

pub trait Material: Send + Sync {
    fn scatter(&self, ray_in: &Ray, rec: &HitRecord) -> Option<ScatterResult>;
    
    fn emitted(&self, _u: f32, _v: f32, _p: Point3) -> Color {
        Color::ZERO
    }
}
```

### Texture

```rust
pub trait Texture: Send + Sync {
    fn value(&self, u: f32, v: f32, p: Point3) -> Color;
}

pub struct SolidColor {
    pub albedo: Color,
}

impl Texture for SolidColor {
    fn value(&self, _u: f32, _v: f32, _p: Point3) -> Color {
        self.albedo
    }
}
```

---

## Primitives

### Sphere

```rust
pub struct Sphere {
    center: Ray,  // origin = center, direction = velocity
    radius: f32,
    material: Arc<dyn Material>,
    bbox: AABB,
}

impl Sphere {
    pub fn new(center: Point3, radius: f32, material: Arc<dyn Material>) -> Self {
        let rvec = Vec3::splat(radius);
        Self {
            center: Ray::new(center, Vec3::ZERO, 0.0),
            radius: radius.max(0.0),
            material,
            bbox: AABB::from_points(center - rvec, center + rvec),
        }
    }

    fn sphere_center(&self, time: f32) -> Point3 {
        self.center.at(time)
    }
}

impl Hittable for Sphere {
    fn hit(&self, ray: &Ray, ray_t: Interval, rec: &mut HitRecord) -> bool {
        let center = self.sphere_center(ray.time);
        let oc = center - ray.origin;
        let a = ray.direction.length_squared();
        let h = ray.direction.dot(oc);
        let c = oc.length_squared() - self.radius * self.radius;
        let discriminant = h * h - a * c;

        if discriminant < 0.0 {
            return false;
        }

        let sqrtd = discriminant.sqrt();
        let mut root = (h - sqrtd) / a;
        if !ray_t.surrounds(root) {
            root = (h + sqrtd) / a;
            if !ray_t.surrounds(root) {
                return false;
            }
        }

        rec.t = root;
        rec.p = ray.at(root);
        let outward_normal = (rec.p - center) / self.radius;
        rec.set_face_normal(ray, outward_normal);
        (rec.u, rec.v) = Self::get_sphere_uv(outward_normal);
        rec.material = Arc::clone(&self.material);
        true
    }

    fn bounding_box(&self) -> AABB {
        self.bbox
    }
}

impl Sphere {
    fn get_sphere_uv(p: Vec3) -> (f32, f32) {
        let theta = (-p.y).acos();
        let phi = (-p.z).atan2(p.x) + std::f32::consts::PI;
        let u = phi / (2.0 * std::f32::consts::PI);
        let v = theta / std::f32::consts::PI;
        (u, v)
    }
}
```

### Triangle (Möller-Trumbore)

```rust
pub struct Triangle {
    v0: Point3,
    v1: Point3,
    v2: Point3,
    normal: Vec3,
    material: Arc<dyn Material>,
    bbox: AABB,
}

impl Hittable for Triangle {
    fn hit(&self, ray: &Ray, ray_t: Interval, rec: &mut HitRecord) -> bool {
        let edge1 = self.v1 - self.v0;
        let edge2 = self.v2 - self.v0;

        let h = ray.direction.cross(edge2);
        let a = edge1.dot(h);

        if a.abs() < 1e-8 {
            return false; // Parallel
        }

        let f = 1.0 / a;
        let s = ray.origin - self.v0;
        let u = f * s.dot(h);

        if !(0.0..=1.0).contains(&u) {
            return false;
        }

        let q = s.cross(edge1);
        let v = f * ray.direction.dot(q);

        if v < 0.0 || u + v > 1.0 {
            return false;
        }

        let t = f * edge2.dot(q);

        if !ray_t.contains(t) {
            return false;
        }

        rec.t = t;
        rec.p = ray.at(t);
        rec.set_face_normal(ray, self.normal);
        rec.u = u;
        rec.v = v;
        rec.material = Arc::clone(&self.material);
        true
    }

    fn bounding_box(&self) -> AABB {
        self.bbox
    }
}
```

---

## Materials

### Lambertian

```rust
pub struct Lambertian {
    albedo: Arc<dyn Texture>,
}

impl Material for Lambertian {
    fn scatter(&self, ray_in: &Ray, rec: &HitRecord) -> Option<ScatterResult> {
        let mut scatter_direction = rec.normal + random_unit_vector();
        
        if scatter_direction.near_zero() {
            scatter_direction = rec.normal;
        }

        Some(ScatterResult {
            attenuation: self.albedo.value(rec.u, rec.v, rec.p),
            scattered: Ray::new(rec.p, scatter_direction, ray_in.time),
        })
    }
}
```

### Metal

```rust
pub struct Metal {
    albedo: Color,
    fuzz: f32,
}

impl Material for Metal {
    fn scatter(&self, ray_in: &Ray, rec: &HitRecord) -> Option<ScatterResult> {
        let reflected = reflect(ray_in.direction.normalize(), rec.normal);
        let scattered_dir = reflected + random_unit_vector() * self.fuzz;
        
        if scattered_dir.dot(rec.normal) > 0.0 {
            Some(ScatterResult {
                attenuation: self.albedo,
                scattered: Ray::new(rec.p, scattered_dir, ray_in.time),
            })
        } else {
            None
        }
    }
}

fn reflect(v: Vec3, n: Vec3) -> Vec3 {
    v - n * 2.0 * v.dot(n)
}
```

### Dielectric

```rust
pub struct Dielectric {
    refraction_index: f32,
}

impl Material for Dielectric {
    fn scatter(&self, ray_in: &Ray, rec: &HitRecord) -> Option<ScatterResult> {
        let ri = if rec.front_face {
            1.0 / self.refraction_index
        } else {
            self.refraction_index
        };

        let unit_direction = ray_in.direction.normalize();
        let cos_theta = (-unit_direction).dot(rec.normal).min(1.0);
        let sin_theta = (1.0 - cos_theta * cos_theta).sqrt();

        let cannot_refract = ri * sin_theta > 1.0;
        let direction = if cannot_refract || reflectance(cos_theta, ri) > rand::random() {
            reflect(unit_direction, rec.normal)
        } else {
            refract(unit_direction, rec.normal, ri)
        };

        Some(ScatterResult {
            attenuation: Color::ONE,
            scattered: Ray::new(rec.p, direction, ray_in.time),
        })
    }
}

fn refract(uv: Vec3, n: Vec3, etai_over_etat: f32) -> Vec3 {
    let cos_theta = (-uv).dot(n).min(1.0);
    let r_out_perp = (uv + n * cos_theta) * etai_over_etat;
    let r_out_parallel = n * -(1.0 - r_out_perp.length_squared()).abs().sqrt();
    r_out_perp + r_out_parallel
}

fn reflectance(cosine: f32, ref_idx: f32) -> f32 {
    // Schlick's approximation
    let r0 = ((1.0 - ref_idx) / (1.0 + ref_idx)).powi(2);
    r0 + (1.0 - r0) * (1.0 - cosine).powi(5)
}
```

---

## Random Sampling

```rust
use rand::Rng;

pub fn random_f32() -> f32 {
    rand::thread_rng().gen()
}

pub fn random_range(min: f32, max: f32) -> f32 {
    rand::thread_rng().gen_range(min..max)
}

pub fn random_vec3() -> Vec3 {
    Vec3::new(random_f32(), random_f32(), random_f32())
}

pub fn random_vec3_range(min: f32, max: f32) -> Vec3 {
    Vec3::new(
        random_range(min, max),
        random_range(min, max),
        random_range(min, max),
    )
}

pub fn random_unit_vector() -> Vec3 {
    loop {
        let p = random_vec3_range(-1.0, 1.0);
        let len_sq = p.length_squared();
        if 1e-160 < len_sq && len_sq <= 1.0 {
            return p / len_sq.sqrt();
        }
    }
}

pub fn random_in_unit_disk() -> Vec3 {
    loop {
        let p = Vec3::new(random_range(-1.0, 1.0), random_range(-1.0, 1.0), 0.0);
        if p.length_squared() < 1.0 {
            return p;
        }
    }
}
```

---

## BVH

```rust
use rayon::prelude::*;

pub enum BVHNode {
    Leaf {
        objects: Vec<Arc<dyn Hittable>>,
        bbox: AABB,
    },
    Branch {
        left: Box<BVHNode>,
        right: Box<BVHNode>,
        bbox: AABB,
    },
}

impl BVHNode {
    pub fn new(mut objects: Vec<Arc<dyn Hittable>>) -> Self {
        if objects.is_empty() {
            return Self::Leaf {
                objects: vec![],
                bbox: AABB::EMPTY,
            };
        }

        // Compute combined bounding box
        let bbox = objects.iter()
            .map(|o| o.bounding_box())
            .reduce(AABB::from_boxes)
            .unwrap_or(AABB::EMPTY);

        // Base case: small number of objects
        if objects.len() <= 4 {
            return Self::Leaf { objects, bbox };
        }

        // Sort by longest axis
        let axis = bbox.longest_axis();
        objects.sort_by(|a, b| {
            let a_center = a.bounding_box().centroid();
            let b_center = b.bounding_box().centroid();
            let a_val = match axis { 0 => a_center.x, 1 => a_center.y, _ => a_center.z };
            let b_val = match axis { 0 => b_center.x, 1 => b_center.y, _ => b_center.z };
            a_val.partial_cmp(&b_val).unwrap()
        });

        // Split at midpoint
        let mid = objects.len() / 2;
        let right_objects = objects.split_off(mid);

        // Parallel construction for large sets
        let (left, right) = if objects.len() > 8192 {
            rayon::join(
                || Box::new(Self::new(objects)),
                || Box::new(Self::new(right_objects)),
            )
        } else {
            (
                Box::new(Self::new(objects)),
                Box::new(Self::new(right_objects)),
            )
        };

        Self::Branch { left, right, bbox }
    }
}

impl Hittable for BVHNode {
    fn hit(&self, ray: &Ray, ray_t: Interval, rec: &mut HitRecord) -> bool {
        match self {
            Self::Leaf { objects, bbox } => {
                if !bbox.hit(ray, ray_t) {
                    return false;
                }
                let mut hit_anything = false;
                let mut closest = ray_t.max;
                for obj in objects {
                    if obj.hit(ray, Interval::new(ray_t.min, closest), rec) {
                        hit_anything = true;
                        closest = rec.t;
                    }
                }
                hit_anything
            }
            Self::Branch { left, right, bbox } => {
                if !bbox.hit(ray, ray_t) {
                    return false;
                }
                let hit_left = left.hit(ray, ray_t, rec);
                let max = if hit_left { rec.t } else { ray_t.max };
                let hit_right = right.hit(ray, Interval::new(ray_t.min, max), rec);
                hit_left || hit_right
            }
        }
    }

    fn bounding_box(&self) -> AABB {
        match self {
            Self::Leaf { bbox, .. } => *bbox,
            Self::Branch { bbox, .. } => *bbox,
        }
    }
}
```

---

## Parallel Bucket Rendering

```rust
use rayon::prelude::*;
use std::sync::{Arc, Mutex, atomic::{AtomicU32, Ordering}};
use image::RgbaImage;

#[derive(Clone, Copy)]
pub struct Bucket {
    pub x: u32,
    pub y: u32,
    pub width: u32,
    pub height: u32,
}

pub struct BucketRenderer {
    camera: Camera,
    world: Arc<dyn Hittable>,
    framebuffer: Arc<Mutex<RgbaImage>>,
    completed: AtomicU32,
}

impl BucketRenderer {
    pub fn render(&self, buckets: &[Bucket], spp: u32, max_depth: u32) {
        buckets.par_iter().for_each(|bucket| {
            // Render to local buffer first (avoid lock contention)
            let mut local_buffer = vec![
                [0u8; 4]; 
                (bucket.width * bucket.height) as usize
            ];

            for dy in 0..bucket.height {
                for dx in 0..bucket.width {
                    let x = bucket.x + dx;
                    let y = bucket.y + dy;
                    
                    let mut color = Color::ZERO;
                    for _ in 0..spp {
                        let ray = self.camera.get_ray(x, y);
                        color += self.ray_color(&ray, max_depth);
                    }
                    color /= spp as f32;

                    // Gamma correction
                    let r = (color.x.sqrt().clamp(0.0, 0.999) * 256.0) as u8;
                    let g = (color.y.sqrt().clamp(0.0, 0.999) * 256.0) as u8;
                    let b = (color.z.sqrt().clamp(0.0, 0.999) * 256.0) as u8;

                    let idx = (dy * bucket.width + dx) as usize;
                    local_buffer[idx] = [r, g, b, 255];
                }
            }

            // Copy to framebuffer (single lock per bucket)
            {
                let mut fb = self.framebuffer.lock().unwrap();
                for dy in 0..bucket.height {
                    for dx in 0..bucket.width {
                        let idx = (dy * bucket.width + dx) as usize;
                        fb.put_pixel(
                            bucket.x + dx,
                            bucket.y + dy,
                            image::Rgba(local_buffer[idx]),
                        );
                    }
                }
            }

            self.completed.fetch_add(1, Ordering::Relaxed);
        });
    }

    fn ray_color(&self, ray: &Ray, depth: u32) -> Color {
        if depth == 0 {
            return Color::ZERO;
        }

        let mut rec = HitRecord::default();
        if self.world.hit(ray, Interval::new(0.001, f32::INFINITY), &mut rec) {
            let emitted = rec.material.emitted(rec.u, rec.v, rec.p);
            
            if let Some(scatter) = rec.material.scatter(ray, &rec) {
                emitted + scatter.attenuation * self.ray_color(&scatter.scattered, depth - 1)
            } else {
                emitted
            }
        } else {
            // Background
            self.camera.background
        }
    }
}
```

---

## Utility Macros

```rust
/// Degrees to radians
#[inline]
pub fn degrees_to_radians(degrees: f32) -> f32 {
    degrees * std::f32::consts::PI / 180.0
}

/// Linear interpolation
#[inline]
pub fn lerp(a: f32, b: f32, t: f32) -> f32 {
    a + (b - a) * t
}

/// Color to bytes (with gamma correction)
pub fn color_to_rgba(color: Color, samples: u32) -> [u8; 4] {
    let scale = 1.0 / samples as f32;
    let r = (color.x * scale).sqrt().clamp(0.0, 0.999);
    let g = (color.y * scale).sqrt().clamp(0.0, 0.999);
    let b = (color.z * scale).sqrt().clamp(0.0, 0.999);
    [
        (r * 256.0) as u8,
        (g * 256.0) as u8,
        (b * 256.0) as u8,
        255,
    ]
}
```

---

## Testing Patterns

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_vec3_dot() {
        let a = Vec3::new(1.0, 2.0, 3.0);
        let b = Vec3::new(4.0, 5.0, 6.0);
        assert!((a.dot(b) - 32.0).abs() < 1e-6);
    }

    #[test]
    fn test_ray_at() {
        let ray = Ray::new(
            Vec3::new(0.0, 0.0, 0.0),
            Vec3::new(1.0, 0.0, 0.0),
            0.0,
        );
        let p = ray.at(5.0);
        assert!((p.x - 5.0).abs() < 1e-6);
    }

    #[test]
    fn test_sphere_hit() {
        let mat = Arc::new(Lambertian::new(Color::new(0.5, 0.5, 0.5)));
        let sphere = Sphere::new(Vec3::new(0.0, 0.0, -1.0), 0.5, mat);
        let ray = Ray::new(Vec3::ZERO, Vec3::new(0.0, 0.0, -1.0), 0.0);
        let mut rec = HitRecord::default();
        
        assert!(sphere.hit(&ray, Interval::new(0.001, f32::INFINITY), &mut rec));
        assert!((rec.t - 0.5).abs() < 1e-6);
    }
}
```

---

## File Structure Template

```
bif/
├── Cargo.toml
├── src/
│   ├── main.rs
│   ├── lib.rs
│   ├── types/
│   │   ├── mod.rs
│   │   ├── vec3.rs
│   │   ├── ray.rs
│   │   ├── interval.rs
│   │   └── aabb.rs
│   ├── traits/
│   │   ├── mod.rs
│   │   ├── hittable.rs
│   │   ├── material.rs
│   │   └── texture.rs
│   ├── primitives/
│   │   ├── mod.rs
│   │   ├── sphere.rs
│   │   ├── triangle.rs
│   │   ├── quad.rs
│   │   └── plane.rs
│   ├── materials/
│   │   ├── mod.rs
│   │   ├── lambertian.rs
│   │   ├── metal.rs
│   │   ├── dielectric.rs
│   │   └── diffuse_light.rs
│   ├── acceleration/
│   │   ├── mod.rs
│   │   └── bvh.rs
│   ├── camera.rs
│   ├── renderer/
│   │   ├── mod.rs
│   │   └── bucket.rs
│   ├── io/
│   │   ├── mod.rs
│   │   └── obj.rs
│   └── utils.rs
└── tests/
    └── integration.rs
```
