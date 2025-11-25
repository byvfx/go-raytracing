# BIF - Modern Scene Assembler & Renderer

BIF is a Clarisse-inspired scene assembler and renderer built in Rust, focusing on massive scene scalability through instancing and non-destructive USD workflows.

## Project Status

Currently migrating from a feature-complete Go raytracer to a production Rust architecture.

### Completed (Go Implementation)
- Path tracer with multiple importance sampling
- BVH acceleration structure  
- Complete primitive library (sphere, plane, quad, triangle, disk, box, pyramid)
- Material system (diffuse, metal, dielectric, emissive)
- Texture support (procedural and image-based)
- Camera with depth of field and motion blur
- Next Event Estimation for direct lighting
- Progressive rendering

### In Progress (Rust Migration)
- Core math library port
- Scene graph with instancing
- USD authoring via C++ shim
- Qt/egui UI integration

## Repository Structure

```
bif/
├── README.md           # This file
├── Cargo.toml          # Rust workspace
├── CMakeLists.txt      # Qt build (optional)
├── cpp/                # C++ components
│   └── shims/          # USD/MaterialX bridges
├── crates/             # Rust crates
│   ├── app/            # Main application
│   ├── scene/          # Scene graph & instances
│   ├── renderer/       # Path tracer (ported from Go)
│   ├── viewport/       # wgpu preview renderer
│   ├── io_gltf/        # glTF loader
│   └── usd_bridge/     # USD FFI wrapper
├── docs/               # Documentation
│   ├── bif_poc_guide.md
│   ├── bif_migration_guide.md
│   └── bif_gpu_rendering_architecture.md
└── legacy/             # Go implementation reference
    └── rt/             # Original Go raytracer
```

## Quick Start

### Prerequisites

```bash
# Rust
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# System dependencies (Ubuntu/Debian)
sudo apt-get install cmake pkg-config libssl-dev python3-dev

# USD (choose one)
# Option 1: Pre-built from NVIDIA
wget https://developer.nvidia.com/usd-22-11-linux-python-39

# Option 2: vcpkg
vcpkg install usd
```

### Build & Run

```bash
# Clone repository
git clone https://github.com/yourusername/bif.git
cd bif

# Build Rust components
cargo build --release

# Run proof of concept
cargo run --release --bin bif
```

## Development Roadmap

### Phase 1: Core Migration (Current)
- [x] Project structure setup
- [ ] Math library port from Go
- [ ] Basic scene representation
- [ ] Simple path tracer

### Phase 2: Scene Assembly
- [ ] USD authoring support
- [ ] Instance management
- [ ] Layer system
- [ ] Procedural scattering

### Phase 3: Production Renderer
- [ ] Port full material system
- [ ] Texture pipeline (TX/OIIO)
- [ ] Progressive refinement
- [ ] Denoising (OIDN)

### Phase 4: User Interface
- [ ] egui prototype UI
- [ ] Qt production interface
- [ ] Viewport integration
- [ ] Node graph editor

### Phase 5: Advanced Features
- [ ] GPU acceleration (wgpu/OptiX)
- [ ] Hydra render delegate
- [ ] Python scripting
- [ ] Network rendering

## Design Principles

1. **Everything is an instance** - Never expand geometry, only reference prototypes
2. **Non-destructive workflow** - All edits are layers that can be toggled/reordered
3. **Lazy evaluation** - Only compute what's needed for the current view
4. **USD-native** - All operations author USD, ensuring pipeline compatibility

## Performance Targets

- 10K instances: < 500ms scene creation
- 100K instances: < 5s USD export
- 1M instances: < 60MB memory overhead
- Interactive viewport: 60 FPS with bounding boxes
- Progressive render: First pixels in < 1s

## Contributing

This project is in early development. Contributions welcome, especially in:
- Rust performance optimization
- USD pipeline integration
- Qt UI development
- GPU compute shaders

## License

MIT (pending final decision)

## Acknowledgments

- Inspired by Isotropix Clarisse
- Based on "Ray Tracing in One Weekend" series
- Built with USD, MaterialX, and OpenImageIO
