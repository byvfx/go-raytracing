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

### Current Go Repository (`go-raytracing/`)

```text
go-raytracing/
├── main.go                # Entry point for the Go raytracer
├── go.mod / go.sum        # Go module definition
├── assets/                # Shared textures/models
│   ├── images/
│   └── models/
├── rt/                    # Production Go path tracer (reference implementation)
├── devlog/                # Daily logs and experiment notes
├── RUST_PORT_CHECKLIST.md # Root-level alias for quick access
└── port/                  # Rust migration plans (this folder)
    ├── README.md
    ├── RUST_PORT_CHECKLIST.md
    ├── rust_port_learning_plan.md
    ├── bif_migration_guide.md
    ├── bif_poc_guide.md
    └── bif_gpu_rendering_architecture.md
```

This structure keeps the feature-complete Go renderer available as the ground truth while you build Rust scaffolding inside `port/`.

### Target Rust Workspace (`bif/`)

When you spin up the standalone Rust workspace (see `bif_poc_guide.md` Step 1), the layout will look like:

```text
bif/
├── README.md           # Top-level vision & roadmap
├── Cargo.toml          # Rust workspace manifest
├── CMakeLists.txt      # Qt build (optional)
├── cpp/                # C++ components (USD/MaterialX bridges)
│   └── shims/
│       └── usd_shim/
├── crates/
│   ├── app/            # Main application (Step 7)
│   ├── scene/          # Scene graph & instances (Step 2)
│   ├── renderer/       # CPU path tracer (Step 5)
│   ├── viewport/       # wgpu preview renderer (Step 6)
│   ├── io_gltf/        # glTF loader (Step 3)
│   └── usd_bridge/     # USD FFI wrapper (Step 4)
└── docs/
    ├── bif_poc_guide.md
    ├── bif_migration_guide.md
    ├── bif_gpu_rendering_architecture.md
    └── rust_port_learning_plan.md
```

> **Note:** A `framebuffer` crate for shared accumulation may be added post-PoC when integrating progressive refinement (see `rust_port_learning_plan.md` Stage 4).

Use `bif_poc_guide.md` for bring-up details and `rust_port_learning_plan.md` for the staged learning workflow that ties back to the checklist.

## Documentation Scope Map

| Doc | Scope & When to Use |
| --- | --- |
| `rust_port_learning_plan.md` | Stage-by-stage timeline (Stage 0-7) describing what to build first, validation gates, and how progress is measured. Start here whenever you plan the next sprint. |
| `RUST_PORT_CHECKLIST.md` | Exhaustive task list organized by subsystems. Use it alongside the learning plan; each Stage references specific checklist sections to tick off. |
| `bif_poc_guide.md` | Step-by-step instructions for standing up the Rust workspace (Stages 0-2) with crate scaffolding, build commands, and PoC goals. |
| `bif_migration_guide.md` | Narrative Go→Rust migration story with code snippets and a week-by-week breakdown aligned with Stages 1-5. Reference it when translating concrete features. |
| `bif_gpu_rendering_architecture.md` | Future-looking GPU/hybrid roadmap covering Stage 7+ work (progressive viewport, coordinator). Review after CPU parity is achieved. |
| `devlog/` entries | Chronological decisions, metrics, and experiment logs. Update after each stage to keep history synced with the checklist. |

This alignment ensures every document answers a specific question (what, how, or when) without contradicting the others.

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
git clone https://github.com/byvfx/bif.git
cd bif

# Build Rust components
cargo build --release

# Run proof of concept
cargo run --release --bin bif
```

## Development Roadmap

### Phase 1: Core Migration (Current)

- [ ] Project structure setup
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
