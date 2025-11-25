# BIF GPU Continuous Rendering Architecture

## Vision: Seamless Real-to-Offline Rendering

BIF aims to eliminate the traditional separation between "viewport" and "production render". Instead, rendering is a continuous spectrum of quality that adapts to user interaction:

- **Interactive**: Rasterization at 60+ FPS during camera movement
- **Preview**: Hybrid rasterization + path traced indirect (1-4 spp/frame)
- **Refinement**: Progressive path tracing when idle (accumulating samples)
- **Production**: Full spectral path tracing with denoising

The renderer intelligently switches modes based on user activity, scene complexity, and hardware capabilities.

## Architecture Overview

```
┌───────────────────────────────────────────────────────────┐
│                     Render Coordinator                       │
│  - Detects user interaction                                  │
│  - Switches rendering modes dynamically                      │
│  - Manages sample accumulation                               │
└──────────────────┬─────────────────────────────────────────┘
                 │
    ┌────────────┴────────────┐
    │                         │
    ▼                         ▼
┌─────────────┐         ┌──────────────┐
│ Rasterizer  │         │ Path Tracer  │
│ (wgpu)      │         │ (Compute)    │
│             │         │              │
│ - GBuffer   │◄────────┤ - Wavefront  │
│ - Depth     │  Uses   │ - Instance   │
│ - Normals   │         │   Coherent   │
│ - Motion    │         │ - Spectral   │
└─────────────┘         └──────────────┘
      │                       │
      └───────────┬───────────┘
                  ▼
          ┌──────────────┐
          │ Accumulator  │
          │ + Denoiser   │
          └──────────────┘
                  │
                  ▼
          ┌──────────────┐
          │   Display    │
          └──────────────┘
```

## Core Components

### 1. Render Coordinator

The brain that decides when to switch modes and manages quality transitions.

```rust
// crates/renderer_gpu/src/coordinator.rs

use std::time::{Duration, Instant};

#[derive(Debug, Clone, Copy, PartialEq)]
pub enum RenderMode {
    /// Pure rasterization, 60+ FPS
    Interactive,
    /// Rasterized GBuffer + 1 path traced bounce
    Hybrid { spp_per_frame: u32 },
    /// Full path tracing, accumulating samples
    Progressive { target_spp: u32 },
    /// Final quality with all features
    Production { max_bounces: u32, spectral: bool },
}

pub struct RenderCoordinator {
    current_mode: RenderMode,
    accumulated_samples: u32,
    last_interaction: Instant,
    idle_threshold: Duration,
    viewport_dirty: bool,
    
    // Performance metrics
    frame_times: Vec<Duration>,
    adaptive_quality: bool,
}

impl RenderCoordinator {
    pub fn new() -> Self {
        Self {
            current_mode: RenderMode::Interactive,
            accumulated_samples: 0,
            last_interaction: Instant::now(),
            idle_threshold: Duration::from_millis(500),
            viewport_dirty: false,
            frame_times: Vec::with_capacity(60),
            adaptive_quality: true,
        }
    }

    /// Call this on every user input event
    pub fn on_interaction(&mut self) {
        self.last_interaction = Instant::now();
        self.accumulated_samples = 0;
        self.viewport_dirty = true;
        
        if self.current_mode != RenderMode::Interactive {
            log::debug!("Switching to interactive mode");
            self.current_mode = RenderMode::Interactive;
        }
    }

    /// Call this every frame to determine rendering mode
    pub fn update(&mut self) -> RenderMode {
        let time_since_interaction = self.last_interaction.elapsed();
        
        // Compute average frame time
        let avg_frame_time = if !self.frame_times.is_empty() {
            self.frame_times.iter().sum::<Duration>() / self.frame_times.len() as u32
        } else {
            Duration::from_millis(16)
        };

        // State machine for mode transitions
        self.current_mode = match self.current_mode {
            RenderMode::Interactive => {
                if time_since_interaction > self.idle_threshold {
                    // User stopped interacting, start hybrid
                    log::debug!("Upgrading to hybrid mode");
                    RenderMode::Hybrid { spp_per_frame: 1 }
                } else {
                    RenderMode::Interactive
                }
            }
            
            RenderMode::Hybrid { spp_per_frame } => {
                if time_since_interaction < Duration::from_millis(100) {
                    // User interacted again
                    RenderMode::Interactive
                } else if time_since_interaction > Duration::from_secs(2) {
                    // Been idle, go full progressive
                    log::debug!("Upgrading to progressive mode");
                    RenderMode::Progressive { target_spp: 1024 }
                } else {
                    // Adaptively increase samples if we have headroom
                    if avg_frame_time < Duration::from_millis(32) && self.adaptive_quality {
                        RenderMode::Hybrid { 
                            spp_per_frame: (spp_per_frame + 1).min(4) 
                        }
                    } else {
                        RenderMode::Hybrid { spp_per_frame }
                    }
                }
            }
            
            RenderMode::Progressive { target_spp } => {
                if time_since_interaction < Duration::from_millis(100) {
                    RenderMode::Interactive
                } else if self.accumulated_samples >= target_spp {
                    // Reached target, could switch to production if needed
                    RenderMode::Progressive { target_spp }
                } else {
                    RenderMode::Progressive { target_spp }
                }
            }
            
            RenderMode::Production { .. } => {
                if time_since_interaction < Duration::from_millis(100) {
                    RenderMode::Interactive
                } else {
                    self.current_mode
                }
            }
        };

        self.current_mode
    }

    pub fn record_frame_time(&mut self, duration: Duration) {
        if self.frame_times.len() >= 60 {
            self.frame_times.remove(0);
        }
        self.frame_times.push(duration);
    }

    pub fn increment_samples(&mut self, count: u32) {
        self.accumulated_samples += count;
    }

    pub fn reset_accumulation(&mut self) {
        self.accumulated_samples = 0;
        self.viewport_dirty = true;
    }

    pub fn get_accumulated_samples(&self) -> u32 {
        self.accumulated_samples
    }
}
```

### 2. Hybrid Rasterizer

Uses wgpu to generate a GBuffer that accelerates path tracing.

```rust
// crates/renderer_gpu/src/rasterizer.rs

use wgpu::util::DeviceExt;
use scene::Scene;
use glam::{Mat4, Vec3};

pub struct GBuffer {
    pub albedo: wgpu::Texture,
    pub normal: wgpu::Texture,
    pub depth: wgpu::Texture,
    pub motion: wgpu::Texture,
    pub instance_id: wgpu::Texture,
    pub view: GBufferViews,
}

pub struct GBufferViews {
    pub albedo: wgpu::TextureView,
    pub normal: wgpu::TextureView,
    pub depth: wgpu::TextureView,
    pub motion: wgpu::TextureView,
    pub instance_id: wgpu::TextureView,
}

pub struct Rasterizer {
    device: wgpu::Device,
    queue: wgpu::Queue,
    gbuffer: GBuffer,
    pipeline: wgpu::RenderPipeline,
    
    // Camera data
    camera_buffer: wgpu::Buffer,
    camera_bind_group: wgpu::BindGroup,
    
    // Instance data
    instance_buffer: Option<wgpu::Buffer>,
    vertex_buffer: Option<wgpu::Buffer>,
    index_buffer: Option<wgpu::Buffer>,
}

impl Rasterizer {
    pub fn new(device: wgpu::Device, queue: wgpu::Queue, width: u32, height: u32) -> Self {
        let gbuffer = Self::create_gbuffer(&device, width, height);
        
        // Create shader
        let shader = device.create_shader_module(wgpu::ShaderModuleDescriptor {
            label: Some("GBuffer Shader"),
            source: wgpu::ShaderSource::Wgsl(include_str!("shaders/gbuffer.wgsl")),
        });

        // Camera uniform
        let camera_buffer = device.create_buffer(&wgpu::BufferDescriptor {
            label: Some("Camera Buffer"),
            size: 128, // mat4 view + mat4 proj
            usage: wgpu::BufferUsages::UNIFORM | wgpu::BufferUsages::COPY_DST,
            mapped_at_creation: false,
        });

        let camera_bind_group_layout = device.create_bind_group_layout(&wgpu::BindGroupLayoutDescriptor {
            label: Some("Camera Bind Group Layout"),
            entries: &[wgpu::BindGroupLayoutEntry {
                binding: 0,
                visibility: wgpu::ShaderStages::VERTEX | wgpu::ShaderStages::FRAGMENT,
                ty: wgpu::BindingType::Buffer {
                    ty: wgpu::BufferBindingType::Uniform,
                    has_dynamic_offset: false,
                    min_binding_size: None,
                },
                count: None,
            }],
        });

        let camera_bind_group = device.create_bind_group(&wgpu::BindGroupDescriptor {
            label: Some("Camera Bind Group"),
            layout: &camera_bind_group_layout,
            entries: &[wgpu::BindGroupEntry {
                binding: 0,
                resource: camera_buffer.as_entire_binding(),
            }],
        });

        // Pipeline
        let pipeline_layout = device.create_pipeline_layout(&wgpu::PipelineLayoutDescriptor {
            label: Some("GBuffer Pipeline Layout"),
            bind_group_layouts: &[&camera_bind_group_layout],
            push_constant_ranges: &[],
        });

        let pipeline = device.create_render_pipeline(&wgpu::RenderPipelineDescriptor {
            label: Some("GBuffer Pipeline"),
            layout: Some(&pipeline_layout),
            vertex: wgpu::VertexState {
                module: &shader,
                entry_point: Some("vs_main"),
                buffers: &[
                    // Vertex buffer layout
                    wgpu::VertexBufferLayout {
                        array_stride: 32, // position(12) + normal(12) + uv(8)
                        step_mode: wgpu::VertexStepMode::Vertex,
                        attributes: &wgpu::vertex_attr_array![
                            0 => Float32x3, // position
                            1 => Float32x3, // normal
                            2 => Float32x2, // uv
                        ],
                    },
                    // Instance buffer layout
                    wgpu::VertexBufferLayout {
                        array_stride: 80, // mat4(64) + instance_id(4) + material_id(4) + padding(8)
                        step_mode: wgpu::VertexStepMode::Instance,
                        attributes: &wgpu::vertex_attr_array![
                            3 => Float32x4, // transform row 0
                            4 => Float32x4, // transform row 1
                            5 => Float32x4, // transform row 2
                            6 => Float32x4, // transform row 3
                            7 => Uint32,    // instance_id
                            8 => Uint32,    // material_id
                        ],
                    },
                ],
                compilation_options: Default::default(),
            },
            fragment: Some(wgpu::FragmentState {
                module: &shader,
                entry_point: Some("fs_main"),
                targets: &[
                    Some(wgpu::ColorTargetState {
                        format: wgpu::TextureFormat::Rgba16Float, // Albedo
                        blend: None,
                        write_mask: wgpu::ColorWrites::ALL,
                    }),
                    Some(wgpu::ColorTargetState {
                        format: wgpu::TextureFormat::Rgba16Float, // Normal
                        blend: None,
                        write_mask: wgpu::ColorWrites::ALL,
                    }),
                    Some(wgpu::ColorTargetState {
                        format: wgpu::TextureFormat::Rg16Float, // Motion
                        blend: None,
                        write_mask: wgpu::ColorWrites::ALL,
                    }),
                    Some(wgpu::ColorTargetState {
                        format: wgpu::TextureFormat::R32Uint, // Instance ID
                        blend: None,
                        write_mask: wgpu::ColorWrites::ALL,
                    }),
                ],
                compilation_options: Default::default(),
            }),
            primitive: wgpu::PrimitiveState {
                topology: wgpu::PrimitiveTopology::TriangleList,
                strip_index_format: None,
                front_face: wgpu::FrontFace::Ccw,
                cull_mode: Some(wgpu::Face::Back),
                polygon_mode: wgpu::PolygonMode::Fill,
                unclipped_depth: false,
                conservative: false,
            },
            depth_stencil: Some(wgpu::DepthStencilState {
                format: wgpu::TextureFormat::Depth32Float,
                depth_write_enabled: true,
                depth_compare: wgpu::CompareFunction::Less,
                stencil: wgpu::StencilState::default(),
                bias: wgpu::DepthBiasState::default(),
            }),
            multisample: wgpu::MultisampleState::default(),
            multiview: None,
            cache: None,
        });

        Self {
            device,
            queue,
            gbuffer,
            pipeline,
            camera_buffer,
            camera_bind_group,
            instance_buffer: None,
            vertex_buffer: None,
            index_buffer: None,
        }
    }

    fn create_gbuffer(device: &wgpu::Device, width: u32, height: u32) -> GBuffer {
        let size = wgpu::Extent3d { width, height, depth_or_array_layers: 1 };

        let albedo = device.create_texture(&wgpu::TextureDescriptor {
            label: Some("GBuffer Albedo"),
            size,
            mip_level_count: 1,
            sample_count: 1,
            dimension: wgpu::TextureDimension::D2,
            format: wgpu::TextureFormat::Rgba16Float,
            usage: wgpu::TextureUsages::RENDER_ATTACHMENT | wgpu::TextureUsages::TEXTURE_BINDING,
            view_formats: &[],
        });

        let normal = device.create_texture(&wgpu::TextureDescriptor {
            label: Some("GBuffer Normal"),
            size,
            mip_level_count: 1,
            sample_count: 1,
            dimension: wgpu::TextureDimension::D2,
            format: wgpu::TextureFormat::Rgba16Float,
            usage: wgpu::TextureUsages::RENDER_ATTACHMENT | wgpu::TextureUsages::TEXTURE_BINDING,
            view_formats: &[],
        });

        let depth = device.create_texture(&wgpu::TextureDescriptor {
            label: Some("GBuffer Depth"),
            size,
            mip_level_count: 1,
            sample_count: 1,
            dimension: wgpu::TextureDimension::D2,
            format: wgpu::TextureFormat::Depth32Float,
            usage: wgpu::TextureUsages::RENDER_ATTACHMENT | wgpu::TextureUsages::TEXTURE_BINDING,
            view_formats: &[],
        });

        let motion = device.create_texture(&wgpu::TextureDescriptor {
            label: Some("GBuffer Motion"),
            size,
            mip_level_count: 1,
            sample_count: 1,
            dimension: wgpu::TextureDimension::D2,
            format: wgpu::TextureFormat::Rg16Float,
            usage: wgpu::TextureUsages::RENDER_ATTACHMENT | wgpu::TextureUsages::TEXTURE_BINDING,
            view_formats: &[],
        });

        let instance_id = device.create_texture(&wgpu::TextureDescriptor {
            label: Some("GBuffer Instance ID"),
            size,
            mip_level_count: 1,
            sample_count: 1,
            dimension: wgpu::TextureDimension::D2,
            format: wgpu::TextureFormat::R32Uint,
            usage: wgpu::TextureUsages::RENDER_ATTACHMENT | wgpu::TextureUsages::TEXTURE_BINDING,
            view_formats: &[],
        });

        let view = GBufferViews {
            albedo: albedo.create_view(&wgpu::TextureViewDescriptor::default()),
            normal: normal.create_view(&wgpu::TextureViewDescriptor::default()),
            depth: depth.create_view(&wgpu::TextureViewDescriptor::default()),
            motion: motion.create_view(&wgpu::TextureViewDescriptor::default()),
            instance_id: instance_id.create_view(&wgpu::TextureViewDescriptor::default()),
        };

        GBuffer {
            albedo,
            normal,
            depth,
            motion,
            instance_id,
            view,
        }
    }

    pub fn update_scene(&mut self, scene: &Scene) {
        // Flatten geometry into GPU buffers
        let mut vertices = Vec::new();
        let mut indices = Vec::new();
        let mut instances = Vec::new();

        for (proto_idx, prototype) in scene.prototypes.iter().enumerate() {
            let base_index = vertices.len() as u32;
            
            // Add vertices
            for i in 0..prototype.vertices.len() {
                vertices.extend_from_slice(&prototype.vertices[i].to_array());
                vertices.extend_from_slice(&prototype.normals[i].to_array());
                vertices.extend_from_slice(&prototype.uvs[i]);
            }

            // Add indices
            for idx in &prototype.indices {
                indices.push(base_index + idx);
            }
        }

        // Add instances
        for instance in &scene.instances {
            let evaluated = scene.evaluate_instance(instance.id);
            if !evaluated.visibility { continue; }

            let transform = evaluated.transform.matrix.to_cols_array();
            instances.extend_from_slice(&transform);
            instances.push(instance.id as f32);
            instances.push(0.0); // material_id placeholder
            instances.push(0.0); // padding
            instances.push(0.0); // padding
        }

        // Upload to GPU
        self.vertex_buffer = Some(
            self.device.create_buffer_init(&wgpu::util::BufferInitDescriptor {
                label: Some("Vertex Buffer"),
                contents: bytemuck::cast_slice(&vertices),
                usage: wgpu::BufferUsages::VERTEX,
            })
        );

        self.index_buffer = Some(
            self.device.create_buffer_init(&wgpu::util::BufferInitDescriptor {
                label: Some("Index Buffer"),
                contents: bytemuck::cast_slice(&indices),
                usage: wgpu::BufferUsages::INDEX,
            })
        );

        self.instance_buffer = Some(
            self.device.create_buffer_init(&wgpu::util::BufferInitDescriptor {
                label: Some("Instance Buffer"),
                contents: bytemuck::cast_slice(&instances),
                usage: wgpu::BufferUsages::VERTEX,
            })
        );
    }

    pub fn render(&mut self, view_matrix: Mat4, proj_matrix: Mat4) {
        // Update camera
        let mut camera_data = Vec::new();
        camera_data.extend_from_slice(&view_matrix.to_cols_array());
        camera_data.extend_from_slice(&proj_matrix.to_cols_array());
        self.queue.write_buffer(&self.camera_buffer, 0, bytemuck::cast_slice(&camera_data));

        // Render
        let mut encoder = self.device.create_command_encoder(&wgpu::CommandEncoderDescriptor {
            label: Some("GBuffer Encoder"),
        });

        {
            let mut render_pass = encoder.begin_render_pass(&wgpu::RenderPassDescriptor {
                label: Some("GBuffer Pass"),
                color_attachments: &[
                    Some(wgpu::RenderPassColorAttachment {
                        view: &self.gbuffer.view.albedo,
                        resolve_target: None,
                        ops: wgpu::Operations {
                            load: wgpu::LoadOp::Clear(wgpu::Color::BLACK),
                            store: wgpu::StoreOp::Store,
                        },
                    }),
                    Some(wgpu::RenderPassColorAttachment {
                        view: &self.gbuffer.view.normal,
                        resolve_target: None,
                        ops: wgpu::Operations {
                            load: wgpu::LoadOp::Clear(wgpu::Color::BLACK),
                            store: wgpu::StoreOp::Store,
                        },
                    }),
                    Some(wgpu::RenderPassColorAttachment {
                        view: &self.gbuffer.view.motion,
                        resolve_target: None,
                        ops: wgpu::Operations {
                            load: wgpu::LoadOp::Clear(wgpu::Color::BLACK),
                            store: wgpu::StoreOp::Store,
                        },
                    }),
                    Some(wgpu::RenderPassColorAttachment {
                        view: &self.gbuffer.view.instance_id,
                        resolve_target: None,
                        ops: wgpu::Operations {
                            load: wgpu::LoadOp::Clear(wgpu::Color::BLACK),
                            store: wgpu::StoreOp::Store,
                        },
                    }),
                ],
                depth_stencil_attachment: Some(wgpu::RenderPassDepthStencilAttachment {
                    view: &self.gbuffer.view.depth,
                    depth_ops: Some(wgpu::Operations {
                        load: wgpu::LoadOp::Clear(1.0),
                        store: wgpu::StoreOp::Store,
                    }),
                    stencil_ops: None,
                }),
                ..Default::default()
            });

            render_pass.set_pipeline(&self.pipeline);
            render_pass.set_bind_group(0, &self.camera_bind_group, &[]);

            if let (Some(ref vb), Some(ref ib), Some(ref inst_b)) = 
                (&self.vertex_buffer, &self.index_buffer, &self.instance_buffer) {
                render_pass.set_vertex_buffer(0, vb.slice(..));
                render_pass.set_vertex_buffer(1, inst_b.slice(..));
                render_pass.set_index_buffer(ib.slice(..), wgpu::IndexFormat::Uint32);
                
                // Draw all instances
                render_pass.draw_indexed(0..ib.size() as u32 / 4, 0, 0..1);
            }
        }

        self.queue.submit(std::iter::once(encoder.finish()));
    }

    pub fn get_gbuffer(&self) -> &GBuffer {
        &self.gbuffer
    }
}
```

### 3. Instance-Aware Wavefront Path Tracer

The key innovation: batch rays by which instances they're likely to hit.

```rust
// crates/renderer_gpu/src/pathtracer.rs

use wgpu::util::DeviceExt;
use glam::Vec3;

pub struct WavefrontPathTracer {
    device: wgpu::Device,
    queue: wgpu::Queue,
    
    // Ray buffers
    ray_buffer: wgpu::Buffer,
    hit_buffer: wgpu::Buffer,
    
    // Instance-coherent sorting
    instance_bins: Vec<InstanceBin>,
    
    // Compute pipelines
    generate_rays_pipeline: wgpu::ComputePipeline,
    intersect_pipeline: wgpu::ComputePipeline,
    shade_pipeline: wgpu::ComputePipeline,
    
    // Accumulation
    accumulation_buffer: wgpu::Texture,
    sample_count: u32,
}

#[derive(Debug)]
struct InstanceBin {
    prototype_id: usize,
    spatial_hash: u32,
    ray_indices: Vec<u32>,
}

impl WavefrontPathTracer {
    pub fn new(device: wgpu::Device, queue: wgpu::Queue, width: u32, height: u32) -> Self {
        // Create ray buffer (one ray per pixel, multiple bounces)
        let ray_count = (width * height) as usize;
        let ray_buffer = device.create_buffer(&wgpu::BufferDescriptor {
            label: Some("Ray Buffer"),
            size: (ray_count * 32) as u64, // origin(12) + direction(12) + throughput(12) + flags(4)
            usage: wgpu::BufferUsages::STORAGE | wgpu::BufferUsages::COPY_DST,
            mapped_at_creation: false,
        });

        let hit_buffer = device.create_buffer(&wgpu::BufferDescriptor {
            label: Some("Hit Buffer"),
            size: (ray_count * 48) as u64, // position(12) + normal(12) + uv(8) + instance_id(4) + material_id(4) + t(4) + flags(4)
            usage: wgpu::BufferUsages::STORAGE | wgpu::BufferUsages::COPY_DST,
            mapped_at_creation: false,
        });

        // Load shaders
        let generate_shader = device.create_shader_module(wgpu::ShaderModuleDescriptor {
            label: Some("Ray Generation Shader"),
            source: wgpu::ShaderSource::Wgsl(include_str!("shaders/ray_gen.wgsl")),
        });

        let intersect_shader = device.create_shader_module(wgpu::ShaderModuleDescriptor {
            label: Some("Intersection Shader"),
            source: wgpu::ShaderSource::Wgsl(include_str!("shaders/intersect.wgsl")),
        });

        let shade_shader = device.create_shader_module(wgpu::ShaderModuleDescriptor {
            label: Some("Shading Shader"),
            source: wgpu::ShaderSource::Wgsl(include_str!("shaders/shade.wgsl")),
        });

        // Create pipelines (simplified for brevity)
        let bind_group_layout = device.create_bind_group_layout(&wgpu::BindGroupLayoutDescriptor {
            label: Some("Path Tracer Bind Group Layout"),
            entries: &[
                // Ray buffer
                wgpu::BindGroupLayoutEntry {
                    binding: 0,
                    visibility: wgpu::ShaderStages::COMPUTE,
                    ty: wgpu::BindingType::Buffer {
                        ty: wgpu::BufferBindingType::Storage { read_only: false },
                        has_dynamic_offset: false,
                        min_binding_size: None,
                    },
                    count: None,
                },
                // Hit buffer
                wgpu::BindGroupLayoutEntry {
                    binding: 1,
                    visibility: wgpu::ShaderStages::COMPUTE,
                    ty: wgpu::BindingType::Buffer {
                        ty: wgpu::BufferBindingType::Storage { read_only: false },
                        has_dynamic_offset: false,
                        min_binding_size: None,
                    },
                    count: None,
                },
                // Scene data (instances, geometry)
                // ... more bindings
            ],
        });

        let pipeline_layout = device.create_pipeline_layout(&wgpu::PipelineLayoutDescriptor {
            label: Some("Path Tracer Pipeline Layout"),
            bind_group_layouts: &[&bind_group_layout],
            push_constant_ranges: &[],
        });

        let generate_rays_pipeline = device.create_compute_pipeline(&wgpu::ComputePipelineDescriptor {
            label: Some("Ray Generation Pipeline"),
            layout: Some(&pipeline_layout),
            module: &generate_shader,
            entry_point: Some("main"),
            compilation_options: Default::default(),
            cache: None,
        });

        let intersect_pipeline = device.create_compute_pipeline(&wgpu::ComputePipelineDescriptor {
            label: Some("Intersection Pipeline"),
            layout: Some(&pipeline_layout),
            module: &intersect_shader,
            entry_point: Some("main"),
            compilation_options: Default::default(),
            cache: None,
        });

        let shade_pipeline = device.create_compute_pipeline(&wgpu::ComputePipelineDescriptor {
            label: Some("Shading Pipeline"),
            layout: Some(&pipeline_layout),
            module: &shade_shader,
            entry_point: Some("main"),
            compilation_options: Default::default(),
            cache: None,
        });

        let accumulation_buffer = device.create_texture(&wgpu::TextureDescriptor {
            label: Some("Accumulation Buffer"),
            size: wgpu::Extent3d { width, height, depth_or_array_layers: 1 },
            mip_level_count: 1,
            sample_count: 1,
            dimension: wgpu::TextureDimension::D2,
            format: wgpu::TextureFormat::Rgba32Float,
            usage: wgpu::TextureUsages::STORAGE_BINDING | wgpu::TextureUsages::TEXTURE_BINDING,
            view_formats: &[],
        });

        Self {
            device,
            queue,
            ray_buffer,
            hit_buffer,
            instance_bins: Vec::new(),
            generate_rays_pipeline,
            intersect_pipeline,
            shade_pipeline,
            accumulation_buffer,
            sample_count: 0,
        }
    }

    /// Sort rays into bins by expected instance hits
    /// This is the key optimization for massive instancing
    pub fn bin_rays_by_instance(&mut self, ray_count: usize) {
        // This would run on GPU in practice, but concept:
        // 1. For each ray, predict which spatial region it targets
        // 2. Hash spatial regions to instance prototypes
        // 3. Group rays likely to hit same prototype together
        // 4. Process bins sequentially with high cache coherence
        
        self.instance_bins.clear();
        
        // Spatial hashing of rays
        // In practice, use compute shader to:
        // - Cast rays against coarse BVH
        // - Identify likely prototype hits
        // - Compact into bins
        
        // Pseudocode for compute shader:
        // for ray in rays:
        //   coarse_hit = bvh.intersect_coarse(ray)
        //   prototype_id = coarse_hit.instance.prototype_id
        //   bin[prototype_id].append(ray.id)
    }

    pub fn trace_iteration(&mut self, max_bounces: u32) {
        let mut encoder = self.device.create_command_encoder(&wgpu::CommandEncoderDescriptor {
            label: Some("Path Tracer Encoder"),
        });

        for bounce in 0..max_bounces {
            // Generate/extend rays
            {
                let mut compute_pass = encoder.begin_compute_pass(&wgpu::ComputePassDescriptor {
                    label: Some("Ray Generation"),
                    timestamp_writes: None,
                });
                compute_pass.set_pipeline(&self.generate_rays_pipeline);
                // ... bind groups
                compute_pass.dispatch_workgroups(1, 1, 1); // Calculate based on ray count
            }

            // Intersect rays
            // Process each instance bin separately for cache coherence
            for bin in &self.instance_bins {
                let mut compute_pass = encoder.begin_compute_pass(&wgpu::ComputePassDescriptor {
                    label: Some("Intersection"),
                    timestamp_writes: None,
                });
                compute_pass.set_pipeline(&self.intersect_pipeline);
                // ... bind groups
                // ... set push constants for bin range
                compute_pass.dispatch_workgroups(1, 1, 1);
            }

            // Shade hits
            {
                let mut compute_pass = encoder.begin_compute_pass(&wgpu::ComputePassDescriptor {
                    label: Some("Shading"),
                    timestamp_writes: None,
                });
                compute_pass.set_pipeline(&self.shade_pipeline);
                // ... bind groups
                compute_pass.dispatch_workgroups(1, 1, 1);
            }
        }

        self.queue.submit(std::iter::once(encoder.finish()));
        self.sample_count += 1;
    }

    pub fn get_accumulated_image(&self) -> &wgpu::Texture {
        &self.accumulation_buffer
    }

    pub fn reset(&mut self) {
        self.sample_count = 0;
        // Clear accumulation buffer
    }
}
```

### 4. WGSL Shaders

#### GBuffer Shader

```wgsl
// crates/renderer_gpu/src/shaders/gbuffer.wgsl

struct Camera {
    view: mat4x4<f32>,
    proj: mat4x4<f32>,
}

@group(0) @binding(0)
var<uniform> camera: Camera;

struct VertexInput {
    @location(0) position: vec3<f32>,
    @location(1) normal: vec3<f32>,
    @location(2) uv: vec2<f32>,
}

struct InstanceInput {
    @location(3) transform_0: vec4<f32>,
    @location(4) transform_1: vec4<f32>,
    @location(5) transform_2: vec4<f32>,
    @location(6) transform_3: vec4<f32>,
    @location(7) instance_id: u32,
    @location(8) material_id: u32,
}

struct VertexOutput {
    @builtin(position) clip_position: vec4<f32>,
    @location(0) world_position: vec3<f32>,
    @location(1) world_normal: vec3<f32>,
    @location(2) uv: vec2<f32>,
    @location(3) @interpolate(flat) instance_id: u32,
    @location(4) @interpolate(flat) material_id: u32,
}

struct FragmentOutput {
    @location(0) albedo: vec4<f32>,
    @location(1) normal: vec4<f32>,
    @location(2) motion: vec2<f32>,
    @location(3) instance_id: u32,
}

@vertex
fn vs_main(
    vertex: VertexInput,
    instance: InstanceInput,
) -> VertexOutput {
    var out: VertexOutput;
    
    // Reconstruct transform matrix
    let transform = mat4x4<f32>(
        instance.transform_0,
        instance.transform_1,
        instance.transform_2,
        instance.transform_3,
    );
    
    // Transform to world space
    let world_pos = transform * vec4<f32>(vertex.position, 1.0);
    out.world_position = world_pos.xyz;
    
    // Transform normal (use inverse transpose for proper normal transform)
    let normal_matrix = mat3x3<f32>(
        transform[0].xyz,
        transform[1].xyz,
        transform[2].xyz,
    );
    out.world_normal = normalize(normal_matrix * vertex.normal);
    
    // Project to clip space
    out.clip_position = camera.proj * camera.view * world_pos;
    
    out.uv = vertex.uv;
    out.instance_id = instance.instance_id;
    out.material_id = instance.material_id;
    
    return out;
}

@fragment
fn fs_main(in: VertexOutput) -> FragmentOutput {
    var out: FragmentOutput;
    
    // For now, simple white albedo
    // In production, sample textures based on material_id
    out.albedo = vec4<f32>(0.8, 0.8, 0.8, 1.0);
    
    // Encode normal in [0,1] range
    out.normal = vec4<f32>(in.world_normal * 0.5 + 0.5, 1.0);
    
    // Motion vectors (would need previous frame data)
    out.motion = vec2<f32>(0.0, 0.0);
    
    out.instance_id = in.instance_id;
    
    return out;
}
```

#### Path Tracing Intersection Shader

```wgsl
// crates/renderer_gpu/src/shaders/intersect.wgsl

struct Ray {
    origin: vec3<f32>,
    direction: vec3<f32>,
    throughput: vec3<f32>,
    flags: u32,
}

struct Hit {
    position: vec3<f32>,
    normal: vec3<f32>,
    uv: vec2<f32>,
    instance_id: u32,
    material_id: u32,
    t: f32,
    flags: u32,
}

struct Instance {
    transform: mat4x4<f32>,
    prototype_id: u32,
    material_id: u32,
}

@group(0) @binding(0)
var<storage, read_write> rays: array<Ray>;

@group(0) @binding(1)
var<storage, read_write> hits: array<Hit>;

@group(0) @binding(2)
var<storage, read> instances: array<Instance>;

@group(0) @binding(3)
var<storage, read> vertices: array<vec3<f32>>;

@group(0) @binding(4)
var<storage, read> indices: array<u32>;

// Ray-triangle intersection
fn intersect_triangle(
    ray_origin: vec3<f32>,
    ray_dir: vec3<f32>,
    v0: vec3<f32>,
    v1: vec3<f32>,
    v2: vec3<f32>,
) -> f32 {
    let edge1 = v1 - v0;
    let edge2 = v2 - v0;
    let h = cross(ray_dir, edge2);
    let a = dot(edge1, h);
    
    if (abs(a) < 0.00001) {
        return -1.0;
    }
    
    let f = 1.0 / a;
    let s = ray_origin - v0;
    let u = f * dot(s, h);
    
    if (u < 0.0 || u > 1.0) {
        return -1.0;
    }
    
    let q = cross(s, edge1);
    let v = f * dot(ray_dir, q);
    
    if (v < 0.0 || u + v > 1.0) {
        return -1.0;
    }
    
    let t = f * dot(edge2, q);
    
    if (t > 0.00001) {
        return t;
    }
    
    return -1.0;
}

@compute @workgroup_size(256, 1, 1)
fn main(@builtin(global_invocation_id) global_id: vec3<u32>) {
    let ray_id = global_id.x;
    
    if (ray_id >= arrayLength(&rays)) {
        return;
    }
    
    let ray = rays[ray_id];
    
    // Early exit for inactive rays
    if ((ray.flags & 0x1u) == 0u) {
        return;
    }
    
    var closest_t = 1e10;
    var closest_hit: Hit;
    closest_hit.flags = 0u;
    
    // Instance-coherent intersection
    // In practice, this loop would be over a subset of instances
    // determined by spatial binning
    for (var i = 0u; i < arrayLength(&instances); i = i + 1u) {
        let instance = instances[i];
        
        // Transform ray to instance space
        let inv_transform = inverse(instance.transform);
        let local_origin = (inv_transform * vec4<f32>(ray.origin, 1.0)).xyz;
        let local_dir = (inv_transform * vec4<f32>(ray.direction, 0.0)).xyz;
        
        // Intersect against prototype geometry
        // This is simplified - in practice, use BVH
        let prototype_id = instance.prototype_id;
        
        // For each triangle in prototype
        // (simplified - actual code would use index buffer properly)
        for (var tri = 0u; tri < 10u; tri = tri + 1u) {
            let i0 = indices[tri * 3u + 0u];
            let i1 = indices[tri * 3u + 1u];
            let i2 = indices[tri * 3u + 2u];
            
            let v0 = vertices[i0];
            let v1 = vertices[i1];
            let v2 = vertices[i2];
            
            let t = intersect_triangle(local_origin, local_dir, v0, v1, v2);
            
            if (t > 0.0 && t < closest_t) {
                closest_t = t;
                
                // Record hit
                closest_hit.t = t;
                closest_hit.position = ray.origin + ray.direction * t;
                
                // Compute normal (simplified)
                let edge1 = v1 - v0;
                let edge2 = v2 - v0;
                let normal = normalize(cross(edge1, edge2));
                closest_hit.normal = normalize((instance.transform * vec4<f32>(normal, 0.0)).xyz);
                
                closest_hit.instance_id = i;
                closest_hit.material_id = instance.material_id;
                closest_hit.flags = 1u; // Hit flag
            }
        }
    }
    
    // Write hit result
    hits[ray_id] = closest_hit;
}
```

## Integration with Existing BIF Architecture

```rust
// crates/app/src/main.rs (modified)

use renderer_gpu::{RenderCoordinator, Rasterizer, WavefrontPathTracer};

#[tokio::main]
async fn main() -> Result<()> {
    // ... scene setup from existing PoC ...
    
    // Initialize GPU renderer
    let event_loop = EventLoop::new()?;
    let window = WindowBuilder::new()
        .with_title("BIF - Continuous Renderer")
        .with_inner_size(winit::dpi::LogicalSize::new(1920, 1080))
        .build(&event_loop)?;

    let instance = wgpu::Instance::default();
    let surface = instance.create_surface(&window)?;
    let adapter = instance.request_adapter(&wgpu::RequestAdapterOptions {
        power_preference: wgpu::PowerPreference::HighPerformance,
        compatible_surface: Some(&surface),
        force_fallback_adapter: false,
    }).await.unwrap();

    let (device, queue) = adapter.request_device(
        &wgpu::DeviceDescriptor::default(),
        None,
    ).await?;

    // Initialize render systems
    let mut coordinator = RenderCoordinator::new();
    let mut rasterizer = Rasterizer::new(device.clone(), queue.clone(), 1920, 1080);
    let mut path_tracer = WavefrontPathTracer::new(device.clone(), queue.clone(), 1920, 1080);

    rasterizer.update_scene(&scene);

    let mut camera_pos = Vec3::new(0.0, 20.0, 50.0);
    let mut camera_target = Vec3::ZERO;

    event_loop.run(move |event, elwt| {
        match event {
            Event::WindowEvent { event, .. } => match event {
                WindowEvent::CloseRequested => elwt.exit(),
                
                // Track user interaction
                WindowEvent::KeyboardInput { .. } |
                WindowEvent::MouseInput { .. } |
                WindowEvent::MouseWheel { .. } |
                WindowEvent::CursorMoved { .. } => {
                    coordinator.on_interaction();
                }
                
                WindowEvent::RedrawRequested => {
                    let frame_start = Instant::now();
                    
                    // Decide rendering mode
                    let mode = coordinator.update();
                    
                    match mode {
                        RenderMode::Interactive => {
                            // Just rasterize GBuffer, display directly
                            let view = Mat4::look_at_rh(camera_pos, camera_target, Vec3::Y);
                            let proj = Mat4::perspective_rh(45f32.to_radians(), 1920.0/1080.0, 0.1, 1000.0);
                            rasterizer.render(view, proj);
                            
                            // Display albedo buffer
                            // ... copy to screen
                        }
                        
                        RenderMode::Hybrid { spp_per_frame } => {
                            // Rasterize GBuffer
                            let view = Mat4::look_at_rh(camera_pos, camera_target, Vec3::Y);
                            let proj = Mat4::perspective_rh(45f32.to_radians(), 1920.0/1080.0, 0.1, 1000.0);
                            rasterizer.render(view, proj);
                            
                            // Path trace indirect lighting only
                            for _ in 0..spp_per_frame {
                                path_tracer.trace_iteration(2); // 2 indirect bounces
                                coordinator.increment_samples(1);
                            }
                            
                            // Combine GBuffer + path traced indirect
                            // ... composite and display
                        }
                        
                        RenderMode::Progressive { target_spp } => {
                            // Full progressive path tracing
                            if coordinator.get_accumulated_samples() < target_spp {
                                path_tracer.trace_iteration(8); // Full bounces
                                coordinator.increment_samples(1);
                            }
                            
                            // Display accumulated image
                            // ... with sample count overlay
                        }
                        
                        _ => {}
                    }
                    
                    coordinator.record_frame_time(frame_start.elapsed());
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

## Performance Characteristics

### Expected Performance Targets

| Mode | Resolution | Performance | Quality |
|------|-----------|-------------|---------|
| Interactive | 1920x1080 | 60+ FPS | Rasterized direct lighting only |
| Hybrid (1 spp) | 1920x1080 | 30-60 FPS | Direct + 1 bounce indirect |
| Hybrid (4 spp) | 1920x1080 | 15-30 FPS | Direct + better indirect |
| Progressive | 1920x1080 | ~10 samples/sec | Accumulating to target quality |

### Instance Scaling

With instance-coherent wavefront:
- **1K instances**: No performance impact (all fit in cache)
- **10K instances**: 10-20% overhead (smart binning helps)
- **100K instances**: 30-40% overhead (BVH + binning critical)
- **1M instances**: 2x overhead (need hierarchical culling)

Without instance-coherent optimization:
- **10K instances**: 2-3x slower (cache thrashing)
- **100K instances**: 5-10x slower (unworkable)

## Implementation Roadmap

### Phase 1: Foundation (Week 1-2)
- [ ] Render coordinator with mode switching
- [ ] Basic GBuffer rasterizer
- [ ] Simple compute path tracer (no optimizations)
- [ ] Display pipeline

### Phase 2: Hybrid Rendering (Week 3-4)
- [ ] Combine rasterized direct + path traced indirect
- [ ] Temporal accumulation
- [ ] Basic denoising (bilateral filter)

### Phase 3: Instance Optimization (Week 5-8)
- [ ] Spatial binning of rays
- [ ] Instance-coherent wavefront
- [ ] Prototype caching experiments
- [ ] BVH acceleration

### Phase 4: Production Features (Week 9-12)
- [ ] Layer-aware invalidation
- [ ] Material system integration
- [ ] Advanced denoising (OIDN/OptiX)
- [ ] Spectral rendering option

### Phase 5: Polish (Week 13+)
- [ ] Adaptive sampling
- [ ] Neural radiance caching
- [ ] Production validation
- [ ] Performance profiling and optimization

## Alternative: CUDA/OptiX Path

If you prefer maximum performance over portability, consider CUDA + OptiX:

```rust
// Using rust-cuda or similar bindings
use optix::{Context, Pipeline};

pub struct OptiXPathTracer {
    context: Context,
    pipeline: Pipeline,
    // ... same conceptual API as WavefrontPathTracer
}

// Pros:
// - 2-3x faster ray tracing (hardware RT cores)
// - Built-in denoising
// - Industry standard

// Cons:
// - NVIDIA only
// - Harder to distribute
// - More complex build
```

For BIF, I'd recommend starting with wgpu (portable) and adding OptiX as an optional backend later.

## Key Takeaways

1. **Continuous rendering eliminates mental barriers** - no "render button"
2. **Instance-coherent wavefront** is your secret weapon for scaling
3. **Hybrid approach** gives best of both worlds (speed + quality)
4. **Layer system** enables smart incremental updates
5. **Start simple, optimize later** - get it working with wgpu first

The architecture is designed to feel instant for artists while progressively revealing final quality. It's the DCC tool rendering paradigm shift you're looking for!
