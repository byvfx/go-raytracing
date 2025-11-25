# BIF Scene Assembler - Proof of Concept Guide

## Vision
BIF is a modern scene assembler and renderer inspired by Isotropix Clarisse, building on your existing Go raytracer implementation. The project focuses on:
- Massive scene scalability via instancing and strict lazy evaluation
- Non-destructive USD-style layering and overrides  
- Hybrid Rust/C++ architecture for safety and performance
- Choice of UI: lightweight egui for prototyping, Qt for production
- Deterministic procedural workflows with USD authoring

## Existing Foundation
Your Go raytracer already implements:
- Complete path tracing with BVH acceleration
- Multiple primitives (sphere, plane, quad, triangle, disk, box, pyramid)
- Material system (Lambertian, Metal, Dielectric, DiffuseLight)
- Texture support (solid, checker, image, Perlin noise)
- Camera with DOF and motion blur
- Next Event Estimation for direct lighting
- Progressive rendering with Ebiten

This proof of concept will port these features to Rust while adding BIF's scene assembly capabilities.

## Architecture Overview

```
+-------------------- UI Layer (Choose One) -------------------+
| Option A: egui (Pure Rust, quick prototyping)               |
| Option B: Qt 6 (C++/QML, production-ready with docking)      |
+------------------------------|--------------------------------+
                               v
+------------------- UI ↔ Engine Bridge ------------------------+
| egui: Direct Rust calls                                      |
| Qt: cxx-qt / CXX FFI (Qt <-> Rust)                          |
+------------------------------|--------------------------------+
                               v
+------------------------ BIF Core (Rust) ----------------------+
| scene/        USD-style layering, selections, overrides      |
| mats/         StdSurface-like materials, texture slots       |
| scatter/      Rules -> point instancers                      |
| renderer/     CPU path tracer; optional GPU backend          |
| io-*          glTF, images (TX pipeline), USD shim           |
| viewport/     wgpu-based preview renderer                    |
| denoise/      OIDN/OptiX bindings                           |
+------------------------------|--------------------------------+
                               v
+-------------------- Native Libraries -------------------------+
| USD, MaterialX, OIIO (maketx), OpenVDB/NanoVDB, OIDN        |
| Optional: OptiX, Embree, Hydra                               |
+---------------------------------------------------------------+
```

## Week 1 Proof of Concept Goal
Build a minimal scene assembler that can:
- Load a glTF file as a prototype
- Create 10,000 instances with random transforms  
- Export to USD with PointInstancer
- Render with a basic CPU path tracer
- Display results in a viewport (egui/wgpu for quick start, Qt native for production)

## Prerequisites

```bash
# Install Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# Install system dependencies (Ubuntu/Debian)
sudo apt-get install cmake pkg-config libssl-dev python3-dev

# For Qt UI option (production):
sudo apt-get install qt6-base-dev qt6-declarative-dev

# Install USD (for the C++ shim)
# Option 1: Download pre-built USD from NVIDIA
# Option 2: Build minimal USD (see appendix)
```

## Repository Structure

```
bif/
├─ CMakeLists.txt      # Top-level build (if using Qt)
├─ Cargo.toml          # Rust workspace
├─ cpp/                # C++ components (optional for PoC)
│  ├─ ui/              # Qt application 
│  ├─ bridge/          # cxx-qt glue
│  └─ shims/
│     └─ usd_shim/     # Minimal USD C ABI
├─ crates/
│  ├─ app/             # Main application
│  ├─ scene/           # Scene graph, instances
│  ├─ renderer/        # CPU path tracer
│  ├─ viewport/        # wgpu preview
│  ├─ io_gltf/         # glTF loader
│  └─ usd_bridge/      # USD FFI wrapper
└─ assets/             # Test assets
```

## Build Paths

### Path A: Pure Rust with egui (Recommended for PoC)
Fast iteration, single language, minimal dependencies. Use this for Week 1.

### Path B: Qt Frontend + Rust Engine (Production)
Professional UI with docking, better viewport integration. Add after PoC validates core.

---

## Step 1: Project Setup

### 1.1 Create Workspace

```bash
mkdir bif && cd bif
cargo init --name bif
```

### 1.2 Workspace Cargo.toml

```toml
[workspace]
members = [
    "crates/app",
    "crates/scene", 
    "crates/renderer",
    "crates/viewport",
    "crates/io_gltf",
    "crates/usd_bridge",
]
resolver = "2"

[workspace.package]
version = "0.1.0"
edition = "2021"
authors = ["Your Name"]

[workspace.dependencies]
# Math & Geometry
glam = "0.29"
nalgebra = "0.33"

# Async & Parallelism  
rayon = "1.10"
tokio = { version = "1.42", features = ["full"] }

# Graphics
wgpu = "23.0"
winit = "0.30"

# UI Options
egui = "0.29"
egui-wgpu = "0.29"
egui-winit = "0.29"

# File I/O
gltf = "1.4"
image = "0.25"
exr = "1.7"

# Utilities
anyhow = "1.0"
thiserror = "2.0"
env_logger = "0.11"
log = "0.4"
tracing = "0.1"
tracy-client = { version = "0.17", optional = true }

# Serialization
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"

# Random
rand = "0.8"

# FFI (for Qt bridge later)
cxx = "1.0"
```

### 1.3 Optional: CMakeLists.txt (for Qt UI)

```cmake
cmake_minimum_required(VERSION 3.20)
project(bif VERSION 0.1.0)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_AUTOMOC ON)

find_package(Qt6 REQUIRED COMPONENTS Core Widgets)
find_package(Corrosion REQUIRED)

# Import Rust crates
corrosion_import_crate(MANIFEST_PATH Cargo.toml)

# Qt UI executable (for production path)
if(BUILD_QT_UI)
    add_executable(bif-ui
        cpp/ui/main.cpp
        cpp/ui/MainWindow.cpp
        cpp/bridge/rust_engine.cpp
    )
    target_link_libraries(bif-ui
        Qt6::Core
        Qt6::Widgets
        bif_core  # Rust library
    )
endif()
```

## Step 2: Core Scene Representation

### 2.1 Create `crates/scene/Cargo.toml`

```toml
[package]
name = "scene"
version.workspace = true
edition.workspace = true

[dependencies]
glam.workspace = true
serde.workspace = true
anyhow.workspace = true
rand.workspace = true
rayon.workspace = true
```

### 2.2 Create `crates/scene/src/lib.rs`

```rust
use glam::{Mat4, Vec3};
use serde::{Deserialize, Serialize};
use std::sync::Arc;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Transform {
    pub matrix: Mat4,
}

#[derive(Debug, Clone)]
pub struct Mesh {
    pub vertices: Vec<Vec3>,
    pub normals: Vec<Vec3>,
    pub uvs: Vec<[f32; 2]>,
    pub indices: Vec<u32>,
    pub bounds: AABB,
}

#[derive(Debug, Clone)]
pub struct AABB {
    pub min: Vec3,
    pub max: Vec3,
}

#[derive(Debug, Clone)]
pub struct Instance {
    pub prototype_id: usize,
    pub transform: Transform,
    pub id: u32,
    pub visibility: bool,
}

// Layer system for non-destructive edits
#[derive(Debug)]
pub struct Layer {
    pub name: String,
    pub enabled: bool,
    pub overrides: Vec<Override>,
}

#[derive(Debug)]
pub enum Override {
    Transform { instance_id: u32, transform: Transform },
    Visibility { instance_id: u32, visible: bool },
    Material { instance_id: u32, material_path: String },
}

pub struct Scene {
    pub prototypes: Vec<Arc<Mesh>>,
    pub instances: Vec<Instance>,
    pub layers: Vec<Layer>,
}

impl Scene {
    pub fn new() -> Self {
        Scene {
            prototypes: Vec::new(),
            instances: Vec::new(),
            layers: vec![Layer {
                name: "base".to_string(),
                enabled: true,
                overrides: Vec::new(),
            }],
        }
    }

    pub fn add_prototype(&mut self, mesh: Mesh) -> usize {
        self.prototypes.push(Arc::new(mesh));
        self.prototypes.len() - 1
    }

    pub fn add_instance(&mut self, prototype_id: usize, transform: Transform) -> u32 {
        let id = self.instances.len() as u32;
        self.instances.push(Instance {
            prototype_id,
            transform,
            id,
            visibility: true,
        });
        id
    }

    pub fn scatter_instances(
        &mut self,
        prototype_id: usize,
        count: usize,
        bounds: (Vec3, Vec3),
        seed: u64,
    ) {
        use rand::{Rng, SeedableRng};
        use rand::rngs::StdRng;
        
        let mut rng = StdRng::seed_from_u64(seed);
        
        for _ in 0..count {
            let x = rng.gen_range(bounds.0.x..bounds.1.x);
            let y = rng.gen_range(bounds.0.y..bounds.1.y);
            let z = rng.gen_range(bounds.0.z..bounds.1.z);
            
            let position = Vec3::new(x, y, z);
            let rotation = rng.gen_range(0.0..std::f32::consts::TAU);
            let scale = rng.gen_range(0.8..1.2);
            
            let transform = Transform {
                matrix: Mat4::from_scale_rotation_translation(
                    Vec3::splat(scale),
                    glam::Quat::from_rotation_y(rotation),
                    position,
                ),
            };
            
            self.add_instance(prototype_id, transform);
        }
    }

    // Apply layer overrides to get final instance state
    pub fn evaluate_instance(&self, instance_id: u32) -> Instance {
        let mut instance = self.instances[instance_id as usize].clone();
        
        for layer in &self.layers {
            if !layer.enabled { continue; }
            
            for override_op in &layer.overrides {
                match override_op {
                    Override::Transform { id, transform } if *id == instance_id => {
                        instance.transform = transform.clone();
                    }
                    Override::Visibility { id, visible } if *id == instance_id => {
                        instance.visibility = *visible;
                    }
                    _ => {}
                }
            }
        }
        
        instance
    }
}
```

## Step 3: glTF Loader

### 3.1 Create `crates/io_gltf/Cargo.toml`

```toml
[package]
name = "io_gltf"
version.workspace = true
edition.workspace = true

[dependencies]
scene = { path = "../scene" }
gltf.workspace = true
glam.workspace = true
anyhow.workspace = true
log.workspace = true
```

### 3.2 Create `crates/io_gltf/src/lib.rs`

```rust
use anyhow::Result;
use glam::Vec3;
use scene::{Mesh, Scene, Transform, AABB};
use std::path::Path;

pub fn load_gltf<P: AsRef<Path>>(path: P) -> Result<Scene> {
    let (document, buffers, _) = gltf::import(path)?;
    let mut scene = Scene::new();

    for mesh in document.meshes() {
        for primitive in mesh.primitives() {
            let reader = primitive.reader(|buffer| Some(&buffers[buffer.index()]));
            
            let vertices: Vec<Vec3> = reader
                .read_positions()
                .map(|iter| iter.map(Vec3::from).collect())
                .unwrap_or_default();

            let normals: Vec<Vec3> = reader
                .read_normals()
                .map(|iter| iter.map(Vec3::from).collect())
                .unwrap_or_default();

            let uvs: Vec<[f32; 2]> = reader
                .read_tex_coords(0)
                .map(|iter| iter.into_f32().collect())
                .unwrap_or_default();

            let indices: Vec<u32> = reader
                .read_indices()
                .map(|iter| iter.into_u32().collect())
                .unwrap_or_default();

            // Calculate bounds
            let bounds = calculate_bounds(&vertices);

            let mesh = Mesh {
                vertices,
                normals,
                uvs,
                indices,
                bounds,
            };

            scene.add_prototype(mesh);
            break; // For PoC, just take first primitive
        }
        break; // For PoC, just take first mesh
    }

    Ok(scene)
}

fn calculate_bounds(vertices: &[Vec3]) -> AABB {
    let mut min = Vec3::splat(f32::MAX);
    let mut max = Vec3::splat(f32::MIN);
    
    for v in vertices {
        min = min.min(*v);
        max = max.max(*v);
    }
    
    AABB { min, max }
}
```

## Step 4: USD Bridge (FFI)

### 4.1 Create `crates/usd_bridge/Cargo.toml`

```toml
[package]
name = "usd_bridge"
version.workspace = true
edition.workspace = true

[dependencies]
scene = { path = "../scene" }
glam.workspace = true
anyhow.workspace = true
libc = "0.2"

[build-dependencies]
cc = "1.0"
```

### 4.2 Create `cpp/shims/usd_shim/usd_shim.cpp`

```cpp
// Minimal USD C++ shim for Rust FFI
#include <pxr/pxr.h>
#include <pxr/usd/usd/stage.h>
#include <pxr/usd/usdGeom/xform.h>
#include <pxr/usd/usdGeom/mesh.h>
#include <pxr/usd/usdGeom/pointInstancer.h>
#include <pxr/usd/sdf/changeBlock.h>
#include <pxr/base/gf/matrix4f.h>
#include <vector>

PXR_NAMESPACE_USING_DIRECTIVE

extern "C" {

void* usd_create_stage(const char* path) {
    auto stage = UsdStage::CreateNew(std::string(path));
    return new UsdStageRefPtr(stage);
}

void usd_close_stage(void* stage_ptr) {
    delete static_cast<UsdStageRefPtr*>(stage_ptr);
}

int usd_create_point_instancer(
    void* stage_ptr,
    const char* path,
    const float* positions,
    const float* orientations, 
    const float* scales,
    uint32_t count,
    const char* prototype_path
) {
    auto& stage = *static_cast<UsdStageRefPtr*>(stage_ptr);
    
    // Batch edits for performance
    SdfChangeBlock changeBlock;
    
    // Create instancer
    auto instancer = UsdGeomPointInstancer::Define(stage, SdfPath(path));
    
    // Set positions
    VtVec3fArray posArray(count);
    for (size_t i = 0; i < count; ++i) {
        posArray[i] = GfVec3f(
            positions[i * 3],
            positions[i * 3 + 1],
            positions[i * 3 + 2]
        );
    }
    instancer.GetPositionsAttr().Set(posArray);
    
    // Set orientations (quaternions)
    VtQuathArray orientArray(count);
    for (size_t i = 0; i < count; ++i) {
        orientArray[i] = GfQuath(
            orientations[i * 4],     // w
            orientations[i * 4 + 1], // x
            orientations[i * 4 + 2], // y
            orientations[i * 4 + 3]  // z
        );
    }
    instancer.GetOrientationsAttr().Set(orientArray);
    
    // Set scales
    VtVec3fArray scaleArray(count);
    for (size_t i = 0; i < count; ++i) {
        scaleArray[i] = GfVec3f(
            scales[i * 3],
            scales[i * 3 + 1],
            scales[i * 3 + 2]
        );
    }
    instancer.GetScalesAttr().Set(scaleArray);
    
    // Set prototype
    instancer.GetPrototypesRel().AddTarget(SdfPath(prototype_path));
    
    // Set indices (all instances use prototype 0 for now)
    VtIntArray protoIndices(count, 0);
    instancer.GetProtoIndicesAttr().Set(protoIndices);
    
    return 1;
}

int usd_save_stage(void* stage_ptr) {
    auto& stage = *static_cast<UsdStageRefPtr*>(stage_ptr);
    stage->Save();
    return 1;
}

// Additional authoring functions for layer system
void* usd_create_layer(const char* path) {
    auto layer = SdfLayer::CreateNew(std::string(path));
    return new SdfLayerRefPtr(layer);
}

int usd_set_transform(void* stage_ptr, const char* prim_path, const float* matrix) {
    auto& stage = *static_cast<UsdStageRefPtr*>(stage_ptr);
    auto prim = stage->GetPrimAtPath(SdfPath(prim_path));
    
    if (!prim) {
        prim = stage->DefinePrim(SdfPath(prim_path));
    }
    
    UsdGeomXformable xform(prim);
    if (xform) {
        GfMatrix4d mat;
        for (int i = 0; i < 4; ++i) {
            for (int j = 0; j < 4; ++j) {
                mat[i][j] = matrix[i * 4 + j];
            }
        }
        
        UsdGeomXformOp transformOp = xform.AddTransformOp();
        transformOp.Set(mat);
        return 1;
    }
    return 0;
}

} // extern "C"
```

### 4.3 Create `crates/usd_bridge/build.rs`

```rust
fn main() {
    // This is simplified - you'll need to point to your USD installation
    cc::Build::new()
        .cpp(true)
        .file("../../cpp/shims/usd_shim/usd_shim.cpp")
        .include("/usr/local/include")  // USD headers location
        .flag("-std=c++17")
        .compile("usd_shim");
        
    println!("cargo:rustc-link-search=/usr/local/lib");  // USD libs location
    println!("cargo:rustc-link-lib=usd_ms");
}
```

### 4.4 Create `crates/usd_bridge/src/lib.rs`

```rust
use std::ffi::CString;
use std::ptr;
use scene::Scene;
use anyhow::Result;

#[link(name = "usd_shim")]
extern "C" {
    fn usd_create_stage(path: *const i8) -> *mut std::ffi::c_void;
    fn usd_close_stage(stage: *mut std::ffi::c_void);
    fn usd_create_point_instancer(
        stage: *mut std::ffi::c_void,
        path: *const i8,
        positions: *const f32,
        orientations: *const f32,
        scales: *const f32,
        count: u32,
        prototype_path: *const i8,
    ) -> i32;
    fn usd_save_stage(stage: *mut std::ffi::c_void) -> i32;
    fn usd_create_layer(path: *const i8) -> *mut std::ffi::c_void;
    fn usd_set_transform(
        stage: *mut std::ffi::c_void,
        prim_path: *const i8,
        matrix: *const f32,
    ) -> i32;
}

pub struct UsdStage {
    ptr: *mut std::ffi::c_void,
}

impl UsdStage {
    pub fn create(path: &str) -> Result<Self> {
        let c_path = CString::new(path)?;
        let ptr = unsafe { usd_create_stage(c_path.as_ptr()) };
        if ptr.is_null() {
            anyhow::bail!("Failed to create USD stage");
        }
        Ok(UsdStage { ptr })
    }

    pub fn export_instances(&mut self, scene: &Scene) -> Result<()> {
        // Flatten transform data for bulk export
        let mut positions = Vec::new();
        let mut orientations = Vec::new();
        let mut scales = Vec::new();

        for instance in &scene.instances {
            let evaluated = scene.evaluate_instance(instance.id);
            if !evaluated.visibility { continue; }
            
            let (scale, rotation, translation) = 
                evaluated.transform.matrix.to_scale_rotation_translation();
            
            positions.extend_from_slice(&[translation.x, translation.y, translation.z]);
            orientations.extend_from_slice(&[rotation.w, rotation.x, rotation.y, rotation.z]);
            scales.extend_from_slice(&[scale.x, scale.y, scale.z]);
        }

        let instancer_path = CString::new("/instancer")?;
        let prototype_path = CString::new("/instancer/prototypes/mesh0")?;

        unsafe {
            usd_create_point_instancer(
                self.ptr,
                instancer_path.as_ptr(),
                positions.as_ptr(),
                orientations.as_ptr(),
                scales.as_ptr(),
                (positions.len() / 3) as u32,
                prototype_path.as_ptr(),
            );
        }

        Ok(())
    }

    pub fn export_layers(&mut self, scene: &Scene) -> Result<()> {
        // Export each layer as a separate USD layer
        for layer in &scene.layers {
            if !layer.enabled { continue; }
            
            for override_op in &layer.overrides {
                if let scene::Override::Transform { instance_id, transform } = override_op {
                    let prim_path = format!("/instances/inst_{}", instance_id);
                    let c_path = CString::new(prim_path)?;
                    let matrix = transform.matrix.to_cols_array();
                    
                    unsafe {
                        usd_set_transform(self.ptr, c_path.as_ptr(), matrix.as_ptr());
                    }
                }
            }
        }
        
        Ok(())
    }

    pub fn save(&mut self) -> Result<()> {
        unsafe {
            if usd_save_stage(self.ptr) != 1 {
                anyhow::bail!("Failed to save USD stage");
            }
        }
        Ok(())
    }
}

impl Drop for UsdStage {
    fn drop(&mut self) {
        unsafe {
            if !self.ptr.is_null() {
                usd_close_stage(self.ptr);
            }
        }
    }
}
```

## Step 5: Basic Path Tracer

### 5.1 Create `crates/renderer/Cargo.toml`

```toml
[package]
name = "renderer"
version.workspace = true  
edition.workspace = true

[dependencies]
scene = { path = "../scene" }
glam.workspace = true
rayon.workspace = true
image.workspace = true
rand.workspace = true
```

### 5.2 Create `crates/renderer/src/lib.rs`

```rust
use glam::{Vec3, Vec3A};
use rayon::prelude::*;
use scene::{Scene, Mesh, AABB};
use image::{RgbImage, Rgb};
use rand::Rng;

pub struct Camera {
    pub origin: Vec3,
    pub forward: Vec3,
    pub up: Vec3,
    pub fov: f32,
}

pub struct Ray {
    pub origin: Vec3A,
    pub direction: Vec3A,
    pub inv_direction: Vec3A,
}

pub struct Hit {
    pub t: f32,
    pub normal: Vec3A,
    pub position: Vec3A,
    pub instance_id: u32,
}

impl Ray {
    pub fn new(origin: Vec3, direction: Vec3) -> Self {
        let direction = Vec3A::from(direction);
        Ray {
            origin: Vec3A::from(origin),
            direction,
            inv_direction: Vec3A::ONE / direction,
        }
    }

    fn at(&self, t: f32) -> Vec3A {
        self.origin + self.direction * t
    }
}

// Fast AABB intersection for culling
fn ray_aabb_intersect(ray: &Ray, aabb: &AABB) -> bool {
    let t1 = (Vec3A::from(aabb.min) - ray.origin) * ray.inv_direction;
    let t2 = (Vec3A::from(aabb.max) - ray.origin) * ray.inv_direction;
    
    let tmin = t1.min(t2);
    let tmax = t1.max(t2);
    
    let tmin = tmin.max_element();
    let tmax = tmax.min_element();
    
    tmax >= 0.0 && tmin <= tmax
}

fn ray_triangle_intersect(
    ray: &Ray,
    v0: Vec3A,
    v1: Vec3A,
    v2: Vec3A,
) -> Option<f32> {
    let edge1 = v1 - v0;
    let edge2 = v2 - v0;
    let h = ray.direction.cross(edge2);
    let a = edge1.dot(h);

    if a > -0.00001 && a < 0.00001 {
        return None;
    }

    let f = 1.0 / a;
    let s = ray.origin - v0;
    let u = f * s.dot(h);

    if u < 0.0 || u > 1.0 {
        return None;
    }

    let q = s.cross(edge1);
    let v = f * ray.direction.dot(q);

    if v < 0.0 || u + v > 1.0 {
        return None;
    }

    let t = f * edge2.dot(q);
    if t > 0.00001 {
        Some(t)
    } else {
        None
    }
}

fn intersect_mesh(ray: &Ray, mesh: &Mesh, transform: &glam::Mat4) -> Option<Hit> {
    // Early out with AABB test
    let transformed_aabb = AABB {
        min: transform.transform_point3(mesh.bounds.min),
        max: transform.transform_point3(mesh.bounds.max),
    };
    
    if !ray_aabb_intersect(ray, &transformed_aabb) {
        return None;
    }

    let mut closest_hit: Option<Hit> = None;
    let mut closest_t = f32::MAX;

    for i in (0..mesh.indices.len()).step_by(3) {
        let i0 = mesh.indices[i] as usize;
        let i1 = mesh.indices[i + 1] as usize;
        let i2 = mesh.indices[i + 2] as usize;

        let v0 = transform.transform_point3(mesh.vertices[i0]);
        let v1 = transform.transform_point3(mesh.vertices[i1]);
        let v2 = transform.transform_point3(mesh.vertices[i2]);

        if let Some(t) = ray_triangle_intersect(
            ray,
            Vec3A::from(v0),
            Vec3A::from(v1),
            Vec3A::from(v2),
        ) {
            if t < closest_t {
                closest_t = t;
                let position = ray.at(t);
                let normal = (v1 - v0).cross(v2 - v0).normalize();
                closest_hit = Some(Hit {
                    t,
                    position,
                    normal: Vec3A::from(normal),
                    instance_id: 0, // Will be set by caller
                });
            }
        }
    }

    closest_hit
}

pub fn render(scene: &Scene, camera: &Camera, width: u32, height: u32) -> RgbImage {
    let mut image = RgbImage::new(width, height);
    let aspect = width as f32 / height as f32;
    
    let right = camera.forward.cross(camera.up).normalize();
    let up = right.cross(camera.forward).normalize();
    
    let half_height = (camera.fov.to_radians() / 2.0).tan();
    let half_width = aspect * half_height;

    // Parallel rendering with rayon
    let pixels: Vec<_> = (0..height)
        .into_par_iter()
        .flat_map(|y| {
            (0..width).into_par_iter().map(move |x| {
                let u = (x as f32 / width as f32) * 2.0 - 1.0;
                let v = 1.0 - (y as f32 / height as f32) * 2.0;

                let direction = (camera.forward 
                    + right * (u * half_width)
                    + up * (v * half_height))
                    .normalize();

                let ray = Ray::new(camera.origin, direction);

                let mut color = Vec3::ZERO;
                let mut hit_anything = false;

                // Check all instances
                for instance in &scene.instances {
                    let evaluated = scene.evaluate_instance(instance.id);
                    if !evaluated.visibility { continue; }
                    
                    if instance.prototype_id < scene.prototypes.len() {
                        let mesh = &scene.prototypes[instance.prototype_id];
                        
                        if let Some(mut hit) = intersect_mesh(&ray, mesh, &evaluated.transform.matrix) {
                            hit_anything = true;
                            hit.instance_id = instance.id;
                            
                            // Simple shading: normal visualization
                            color = (hit.normal.into(): Vec3) * 0.5 + Vec3::splat(0.5);
                            break; // Take first hit for now
                        }
                    }
                }

                if !hit_anything {
                    // Sky gradient
                    let t = 0.5 * (direction.y + 1.0);
                    color = Vec3::ONE * (1.0 - t) + Vec3::new(0.5, 0.7, 1.0) * t;
                }

                (x, y, Rgb([
                    (color.x * 255.0) as u8,
                    (color.y * 255.0) as u8,
                    (color.z * 255.0) as u8,
                ]))
            })
        })
        .collect();

    // Write pixels to image
    for (x, y, pixel) in pixels {
        image.put_pixel(x, y, pixel);
    }

    image
}
```

## Step 6: Viewport (wgpu for PoC, Qt native for production)

### 6.1 Create `crates/viewport/Cargo.toml`

```toml
[package]
name = "viewport"
version.workspace = true
edition.workspace = true

[dependencies]
wgpu.workspace = true
winit.workspace = true
scene = { path = "../scene" }
glam.workspace = true
bytemuck = "1.19"
log.workspace = true
```

### 6.2 Create `crates/viewport/src/lib.rs`

```rust
use wgpu::util::DeviceExt;
use scene::Scene;
use glam::{Mat4, Vec3};

#[repr(C)]
#[derive(Debug, Copy, Clone, bytemuck::Pod, bytemuck::Zeroable)]
struct InstanceData {
    transform: [[f32; 4]; 4],
    color: [f32; 4],
}

pub struct ViewportCamera {
    pub position: Vec3,
    pub target: Vec3,
    pub up: Vec3,
    pub fov: f32,
    pub aspect: f32,
}

impl ViewportCamera {
    pub fn view_matrix(&self) -> Mat4 {
        Mat4::look_at_rh(self.position, self.target, self.up)
    }

    pub fn projection_matrix(&self) -> Mat4 {
        Mat4::perspective_rh(self.fov.to_radians(), self.aspect, 0.1, 1000.0)
    }
}

pub struct Viewport {
    device: wgpu::Device,
    queue: wgpu::Queue,
    surface: wgpu::Surface<'static>,
    config: wgpu::SurfaceConfiguration,
    instance_buffer: Option<wgpu::Buffer>,
    pipeline: wgpu::RenderPipeline,
    camera: ViewportCamera,
}

impl Viewport {
    pub async fn new(window: &winit::window::Window) -> Self {
        let size = window.inner_size();
        let instance = wgpu::Instance::default();
        
        let surface = instance.create_surface(window).unwrap();
        let adapter = instance
            .request_adapter(&wgpu::RequestAdapterOptions {
                power_preference: wgpu::PowerPreference::HighPerformance,
                compatible_surface: Some(&surface),
                force_fallback_adapter: false,
            })
            .await
            .unwrap();

        let (device, queue) = adapter
            .request_device(&wgpu::DeviceDescriptor::default(), None)
            .await
            .unwrap();

        let config = surface
            .get_default_config(&adapter, size.width, size.height)
            .unwrap();
        surface.configure(&device, &config);

        // Create simple pipeline for drawing bounding boxes
        let shader = device.create_shader_module(wgpu::ShaderModuleDescriptor {
            label: Some("Viewport Shader"),
            source: wgpu::ShaderSource::Wgsl(include_str!("shader.wgsl")),
        });

        let pipeline_layout = device.create_pipeline_layout(&wgpu::PipelineLayoutDescriptor {
            label: Some("Pipeline Layout"),
            bind_group_layouts: &[],
            push_constant_ranges: &[
                wgpu::PushConstantRange {
                    stages: wgpu::ShaderStages::VERTEX,
                    range: 0..128, // mat4 view + mat4 proj
                },
            ],
        });

        let pipeline = device.create_render_pipeline(&wgpu::RenderPipelineDescriptor {
            label: Some("Render Pipeline"),
            layout: Some(&pipeline_layout),
            vertex: wgpu::VertexState {
                module: &shader,
                entry_point: Some("vs_main"),
                buffers: &[wgpu::VertexBufferLayout {
                    array_stride: std::mem::size_of::<InstanceData>() as wgpu::BufferAddress,
                    step_mode: wgpu::VertexStepMode::Instance,
                    attributes: &wgpu::vertex_attr_array![
                        0 => Float32x4, 1 => Float32x4, 2 => Float32x4, 3 => Float32x4,
                        4 => Float32x4,
                    ],
                }],
                compilation_options: Default::default(),
            },
            fragment: Some(wgpu::FragmentState {
                module: &shader,
                entry_point: Some("fs_main"),
                targets: &[Some(config.format.into())],
                compilation_options: Default::default(),
            }),
            primitive: wgpu::PrimitiveState {
                topology: wgpu::PrimitiveTopology::LineList,
                ..Default::default()
            },
            depth_stencil: None,
            multisample: wgpu::MultisampleState::default(),
            multiview: None,
            cache: None,
        });

        let camera = ViewportCamera {
            position: Vec3::new(20.0, 20.0, 20.0),
            target: Vec3::ZERO,
            up: Vec3::Y,
            fov: 45.0,
            aspect: size.width as f32 / size.height as f32,
        };

        Self {
            device,
            queue,
            surface,
            config,
            instance_buffer: None,
            pipeline,
            camera,
        }
    }

    pub fn update_instances(&mut self, scene: &Scene) {
        let instance_data: Vec<InstanceData> = scene.instances
            .iter()
            .filter_map(|inst| {
                let evaluated = scene.evaluate_instance(inst.id);
                if evaluated.visibility {
                    Some(InstanceData {
                        transform: evaluated.transform.matrix.to_cols_array_2d(),
                        color: [0.8, 0.8, 0.8, 1.0],
                    })
                } else {
                    None
                }
            })
            .collect();

        if !instance_data.is_empty() {
            self.instance_buffer = Some(
                self.device.create_buffer_init(&wgpu::util::BufferInitDescriptor {
                    label: Some("Instance Buffer"),
                    contents: bytemuck::cast_slice(&instance_data),
                    usage: wgpu::BufferUsages::VERTEX | wgpu::BufferUsages::COPY_DST,
                })
            );
        }
        
        log::info!("Updated viewport with {} visible instances", instance_data.len());
    }

    pub fn render(&self) {
        let output = self.surface.get_current_texture().unwrap();
        let view = output.texture.create_view(&wgpu::TextureViewDescriptor::default());

        let mut encoder = self.device.create_command_encoder(&wgpu::CommandEncoderDescriptor {
            label: Some("Render Encoder"),
        });

        {
            let mut render_pass = encoder.begin_render_pass(&wgpu::RenderPassDescriptor {
                label: Some("Render Pass"),
                color_attachments: &[Some(wgpu::RenderPassColorAttachment {
                    view: &view,
                    resolve_target: None,
                    ops: wgpu::Operations {
                        load: wgpu::LoadOp::Clear(wgpu::Color {
                            r: 0.1,
                            g: 0.2,
                            b: 0.3,
                            a: 1.0,
                        }),
                        store: wgpu::StoreOp::Store,
                    },
                })],
                depth_stencil_attachment: None,
                ..Default::default()
            });

            render_pass.set_pipeline(&self.pipeline);
            
            // Set camera matrices via push constants
            let view_proj = self.camera.projection_matrix() * self.camera.view_matrix();
            render_pass.set_push_constants(
                wgpu::ShaderStages::VERTEX,
                0,
                bytemuck::cast_slice(&view_proj.to_cols_array()),
            );
            
            // Draw bounding boxes for all instances
            if let Some(ref buffer) = self.instance_buffer {
                render_pass.set_vertex_buffer(0, buffer.slice(..));
                // Draw cube edges (24 vertices for wireframe cube)
                render_pass.draw(0..24, 0..1);
            }
        }

        self.queue.submit(std::iter::once(encoder.finish()));
        output.present();
    }

    pub fn resize(&mut self, new_size: winit::dpi::PhysicalSize<u32>) {
        if new_size.width > 0 && new_size.height > 0 {
            self.config.width = new_size.width;
            self.config.height = new_size.height;
            self.surface.configure(&self.device, &self.config);
            self.camera.aspect = new_size.width as f32 / new_size.height as f32;
        }
    }
}
```

## Step 7: Main Application

### 7.1 Create `crates/app/Cargo.toml`

```toml
[package]
name = "app"
version.workspace = true
edition.workspace = true

[[bin]]
name = "bif"
path = "src/main.rs"

[dependencies]
scene = { path = "../scene" }
io_gltf = { path = "../io_gltf" }
renderer = { path = "../renderer" }
viewport = { path = "../viewport" }
usd_bridge = { path = "../usd_bridge" }

winit.workspace = true
egui.workspace = true
egui-winit.workspace = true
egui-wgpu.workspace = true
glam.workspace = true
anyhow.workspace = true
env_logger.workspace = true
log.workspace = true
tokio.workspace = true
```

### 7.2 Create `crates/app/src/main.rs`

```rust
use anyhow::Result;
use glam::Vec3;
use winit::{
    event::{Event, WindowEvent},
    event_loop::{ControlFlow, EventLoop},
    window::WindowBuilder,
};
use scene::{Scene, Override};
use renderer::{Camera, render};
use std::time::Instant;

#[tokio::main]
async fn main() -> Result<()> {
    env_logger::init();
    
    log::info!("Starting BIF Scene Assembler PoC");

    // Step 1: Load glTF or create default cube
    log::info!("Loading scene...");
    let mut scene = io_gltf::load_gltf("assets/cube.gltf")
        .unwrap_or_else(|_| {
            log::warn!("Could not load glTF, creating default cube");
            let mut scene = Scene::new();
            
            // Create a simple cube mesh
            let vertices = vec![
                Vec3::new(-0.5, -0.5, -0.5),
                Vec3::new(0.5, -0.5, -0.5),
                Vec3::new(0.5, 0.5, -0.5),
                Vec3::new(-0.5, 0.5, -0.5),
                Vec3::new(-0.5, -0.5, 0.5),
                Vec3::new(0.5, -0.5, 0.5),
                Vec3::new(0.5, 0.5, 0.5),
                Vec3::new(-0.5, 0.5, 0.5),
            ];
            
            let normals = vec![Vec3::Y; 8];
            let uvs = vec![[0.0, 0.0]; 8];
            let indices = vec![
                0, 1, 2, 2, 3, 0, // Front
                1, 5, 6, 6, 2, 1, // Right
                5, 4, 7, 7, 6, 5, // Back
                4, 0, 3, 3, 7, 4, // Left
                3, 2, 6, 6, 7, 3, // Top
                4, 5, 1, 1, 0, 4, // Bottom
            ];
            
            let bounds = scene::AABB {
                min: Vec3::splat(-0.5),
                max: Vec3::splat(0.5),
            };
            
            let mesh = scene::Mesh {
                vertices,
                normals,
                uvs,
                indices,
                bounds,
            };
            
            scene.add_prototype(mesh);
            scene
        });

    // Step 2: Create 10,000 instances
    log::info!("Creating 10,000 instances...");
    let start = Instant::now();
    
    scene.scatter_instances(
        0, // prototype_id
        10_000,
        (Vec3::new(-50.0, 0.0, -50.0), Vec3::new(50.0, 5.0, 50.0)),
        42, // seed
    );
    
    log::info!("Created {} instances in {:?}", 
        scene.instances.len(), 
        start.elapsed()
    );

    // Step 3: Add some layer overrides for testing
    log::info!("Adding layer overrides...");
    scene.layers.push(scene::Layer {
        name: "lookdev".to_string(),
        enabled: true,
        overrides: vec![
            Override::Visibility { instance_id: 0, visible: false },
            Override::Visibility { instance_id: 1, visible: false },
            // Hide first 100 instances for testing
        ],
    });

    // Step 4: Export to USD
    log::info!("Exporting to USD...");
    std::fs::create_dir_all("output")?;
    
    let mut usd_stage = usd_bridge::UsdStage::create("output/instances.usda")?;
    usd_stage.export_instances(&scene)?;
    usd_stage.export_layers(&scene)?;
    usd_stage.save()?;
    log::info!("USD export complete");

    // Step 5: Render with path tracer
    log::info!("Rendering with path tracer...");
    let camera = Camera {
        origin: Vec3::new(0.0, 20.0, 50.0),
        forward: Vec3::new(0.0, -0.3, -1.0).normalize(),
        up: Vec3::Y,
        fov: 45.0,
    };
    
    let render_start = Instant::now();
    let image = render(&scene, &camera, 800, 600);
    image.save("output/render.png")?;
    log::info!("Render complete in {:?}", render_start.elapsed());

    // Step 6: Display in viewport
    log::info!("Starting viewport...");
    let event_loop = EventLoop::new()?;
    let window = WindowBuilder::new()
        .with_title("BIF Scene Assembler PoC")
        .with_inner_size(winit::dpi::LogicalSize::new(1280, 720))
        .build(&event_loop)?;

    let mut viewport = viewport::Viewport::new(&window).await;
    viewport.update_instances(&scene);

    // Simple UI state
    let mut show_stats = true;
    let instance_count = scene.instances.len();
    let visible_count = scene.instances.iter()
        .filter(|i| scene.evaluate_instance(i.id).visibility)
        .count();

    event_loop.run(move |event, elwt| {
        match event {
            Event::WindowEvent { event, .. } => match event {
                WindowEvent::CloseRequested => {
                    log::info!("Window closed");
                    elwt.exit();
                }
                WindowEvent::Resized(physical_size) => {
                    viewport.resize(physical_size);
                }
                WindowEvent::RedrawRequested => {
                    viewport.render();
                    
                    if show_stats {
                        log::info!("Instances: {} total, {} visible", 
                            instance_count, visible_count);
                    }
                }
                WindowEvent::KeyboardInput { event, .. } => {
                    if event.state == winit::event::ElementState::Pressed {
                        match event.logical_key {
                            winit::keyboard::Key::Character(c) if c == "s" => {
                                show_stats = !show_stats;
                            }
                            _ => {}
                        }
                    }
                }
                _ => {}
            }
            Event::AboutToWait => {
                window.request_redraw();
            }
            _ => {}
        }
    })?;

    Ok(())
}
```

## Step 8: Build & Run

### 8.1 Create Test Assets

```bash
mkdir -p assets output

# Create a simple cube.gltf or download one
# You can use Blender to export a cube as glTF 2.0
```

### 8.2 Build Options

#### Option A: Pure Rust (Quick Start)

```bash
# Build all crates
cargo build --release

# Run the PoC
cargo run --release --bin bif
```

#### Option B: With Qt UI (Production)

```bash
# Build USD shim first
cd cpp/shims/usd_shim
c++ -shared -fPIC usd_shim.cpp -o libusd_shim.so \
    -I/usr/local/include -L/usr/local/lib -lusd

# Configure with CMake
cd ../../..
mkdir build && cd build
cmake .. -DBUILD_QT_UI=ON
make

# Run
./bif-ui
```

## Step 9: Performance Profiling

Add profiling to track performance:

```toml
# In Cargo.toml
[profile.release]
debug = true  # Keep debug symbols for profiling

# Run with tracy
cargo run --release --features tracy
```

## Step 10: Validation Checklist

### Core Functionality
- [ ] **glTF Loading**: Mesh loads with proper bounds
- [ ] **Instance Creation**: 10,000 instances created < 500ms
- [ ] **Layer System**: Overrides apply correctly
- [ ] **USD Export**: Valid USD with PointInstancer
- [ ] **Path Tracer**: Renders with proper culling
- [ ] **Viewport**: 60 FPS with wireframe boxes

### Performance Targets
- Load time: < 100ms for glTF
- Instance creation: < 500ms for 10,000 instances  
- USD export: < 2 seconds
- First render: < 5 seconds
- Viewport: 60 FPS with 10,000 instances

### Memory Usage
- Monitor RSS with `top` or `htop`
- Should stay under 500MB for 10K instances
- Profile with `valgrind --tool=massif` if needed

## Troubleshooting

### USD Build Issues
```bash
# Download pre-built USD from NVIDIA
wget https://developer.nvidia.com/usd-22-11-linux-python-39
# Or use vcpkg
vcpkg install usd
```

### Performance Issues
```bash
# CPU profiling
cargo install flamegraph
cargo flamegraph --release --bin bif

# GPU profiling (for viewport)
# Use RenderDoc or NSight
```

### Qt Integration Issues
```bash
# Ensure Qt6 and cxx-qt are properly installed
cargo install cxx-qt-build
# Check CMake can find Qt
cmake --find-package -DNAME=Qt6 -DCOMPILER_ID=GNU -DLANGUAGE=CXX -DMODE=EXIST
```

## Next Steps After PoC Success

### Immediate (Leveraging Go Implementation)
1. **Port Core Renderer**: Migrate your Go raytracer to Rust (materials, textures, camera)
2. **Enhance BVH**: Adapt your existing BVH for instance-aware traversal
3. **Integrate NEE**: Port your Next Event Estimation for production lighting

### Short Term (Building on Foundation)
4. **TX Pipeline**: Add OIIO for texture streaming
5. **MaterialX**: Extend your material system with industry-standard shading
6. **Layer System**: Implement non-destructive overrides
7. **Progressive Refinement**: Port and enhance your progressive renderer

### Medium Term (Production Features)
8. **Qt Interface**: Build professional UI around your renderer
9. **GPU Acceleration**: Add wgpu compute or OptiX path
10. **Python Scripting**: Embed Python for pipeline integration
11. **Hydra Delegate**: Create render delegate for USD ecosystem

## Architecture Benefits

This PoC validates:
- **Instancing efficiency**: Never expand geometry
- **Layer composition**: Non-destructive overrides
- **USD authoring**: Direct stage manipulation
- **Hybrid architecture**: Rust core with C++ bridges
- **Flexible UI**: Both quick (egui) and production (Qt) paths

The architecture supports scaling to millions of instances while maintaining interactivity!
