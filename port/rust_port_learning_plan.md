# Rust Port Learning Plan

This document reorganizes the existing porting material in `port/` into a step-by-step loop that lets you learn Rust while keeping the Go raytracer as a reference. Every stage ends with a concrete validation gate so the migration never drifts into "messy" territory.

## Quick Reference for AI Assistants

**What this project is:** Porting a Go raytracer (`rt/` folder) to Rust, while adding scene assembly features (instancing, USD export, viewport) and a user-friendly UI for interactive scene building.

**Key files to read first:**
1. `rt/vec3.go`, `rt/ray.go`, `rt/material.go` — Go reference implementations
2. `port/bif_poc_guide.md` — Step-by-step Rust code to generate
3. `port/RUST_PORT_CHECKLIST.md` — Detailed task list

**Crates to create (PoC):** `app`, `scene`, `renderer`, `viewport`, `scatter`, `io_gltf`, `usd_bridge` (USD is optional)

**UI Development Path:**
1. **PoC Phase**: `egui` for rapid prototyping (pure Rust, immediate mode)
2. **Production Phase**: Migrate to Qt 6 for professional UI (docking, node editor, better viewport integration)

**Viewport Strategy:** The 3D viewport IS the framebuffer. Renders write directly to the viewport texture—no separate "render view" vs "viewport" distinction.

**Do NOT create:** A separate `crates/math` crate — use `glam` directly for the PoC.

**Two paths:**
1. **PoC Path** (fast): Follow `bif_poc_guide.md` Steps 1-7 to get a working demo
2. **Learning Path** (thorough): Follow Stages 0-7 below, implementing custom types before using libraries

**Key PoC Features:**
- 3D viewport that doubles as the render framebuffer
- Simple node-based scatter system (surface points → instance placement)
- egui panels for scene hierarchy, properties, and scatter controls

---

## Document Roles and How to Use Them

| File | Primary Purpose | When to consult |
| --- | --- | --- |
| `port/README.md` | Vision, product goals, and high-level roadmap (Phases 1-5). | Whenever you need context for *why* a step matters. |
| `port/RUST_PORT_CHECKLIST.md` | Granular build list (32 sections) covering math, traits, renderer, testing, and optimizations. | Use as the master checklist; each stage below links to specific sections. |
| `port/bif_migration_guide.md` | Code-oriented migration story with Go → Rust examples plus a week-by-week timeline. | When you want concrete snippets or need to recall an equivalent Go implementation. |
| `port/bif_poc_guide.md` | Proof-of-concept execution path, workspace layout, and crate-by-crate scaffolding. | During initial workspace bring-up and when wiring crates together. |
| `port/bif_gpu_rendering_architecture.md` | Long-term GPU/viewport architecture for hybrid rendering. | Only after CPU renderer parity; feeds Advanced Stage (Stage 6+). |
| `devlog/DEVLOG_2025-11-21.md` | Architecture notes, crate suggestions, and prior experiments (eg. bucket renderer sketch). | When making design trade-offs or crate selections. |

Keeping each document in its lane prevents duplication: the plan below references them rather than restating full content.

## Process Pillars

1. **Mirror → Validate → Integrate**: direct ports first, tests against Go reference second, BIF-specific features third.
2. **Tight Learning Loops**: each stage should compile, run, and have at least one comparison image/log before moving forward.
3. **Single Source of Truth**: update `RUST_PORT_CHECKLIST.md` as items finish; link commits/renders from devlog entries for traceability.
4. **Automation Early**: add unit/integration tests as soon as a component exists to avoid regressions while learning Rust idioms.

## Stage-by-Stage Plan

Each stage references the checklist sections (`RUST_PORT_CHECKLIST.md`) and migration guide sections. Use `rt/` Go code as the oracle for behavior. Suggested durations assume evenings/weekends; adjust as needed.

### Stage 0 – Workspace Reset (½ day)

- **Goal**: Clean slate Rust workspace + baseline tests ready to run.
- **Inputs**: `bif_poc_guide.md` Step 1, `port/README.md` prerequisites.
- **Tasks**:
  - Create the `bif/` directory as a sibling to `go-raytracing/` (or inside it as `go-raytracing/bif/`).
  - Recreate the minimal Cargo workspace per PoC guide: `app`, `scene`, `renderer`, `viewport`, `scatter`, `io_gltf`, and optionally `usd_bridge`.
  - Configure toolchain (`rustup`, `cargo fmt/clippy`, VS Code rust-analyzer).
  - Copy small Go fixtures (e.g., unit test scenes) into `assets/` for later comparisons.
  - Set up basic egui window with wgpu backend (this will become the viewport).
- **Done when**: `cargo check` succeeds, a `tests/smoke.rs` runs, and an empty egui window opens.

> **For AI Assistants**: The PoC does NOT create a `crates/math` crate—it uses `glam` directly. See `bif_poc_guide.md` for the exact crate list.

### Stage 1 – Core Math & Utility Types (1-2 days)

- **Checklist refs**: Sections 1-4 (Vec3, Ray, Interval, AABB).
- **Docs**: `bif_migration_guide.md` Phase 1.1, Go files `rt/vec3.go`, `rt/ray.go`, `rt/interval.go`, `rt/aabb.go` (in parent workspace).
- **Learning focus**: Ownership basics, trait derivations, unit tests with `approx` or custom asserts.
- **Deliverables**:
  - For **PoC path**: Use `glam` crate directly (no custom math crate needed).
  - For **learning path**: Create `crates/math` with custom Vec3/Ray/Interval/AABB types that mirror Go, then optionally swap to `glam` later.
  - Benchmark or quick comparison that matches Go outputs for dot/cross operations.
- **Validation**: `cargo test` comparing expected values against Go reference; document results in devlog.

> **For AI Assistants**: The Go reference files are at `rt/vec3.go`, `rt/ray.go`, etc. in the `go-raytracing` workspace. Read those files to understand the target behavior.

### Stage 2 – Traits and Primitives (2-3 days)

- **Checklist refs**: Sections 5-11 (Hittable, Material shell, Texture shell, Sphere, Plane, Quad, Triangle).
- **Docs**: `bif_migration_guide.md` Phase 2.1/2.2, devlog Port Recommendations.
- **Learning focus**: Trait objects vs enums, `Arc` vs `Box`, pattern matching.
- **Tasks**:
  - Implement `HitRecord`, `Hittable` trait, and primitive structs mirroring Go logic.
  - Port just two materials (Lambertian, Metal) as scaffolding for tests.
- **Validation**: Unit tests invoking single-ray intersection cases mirrored from Go; optional snapshot tests saving normals/UVs to `assets/test_outputs`.

### Stage 3 – Materials & Textures (2-3 days)

- **Checklist refs**: Sections 13-16 (Lambertian, Metal, Dielectric, DiffuseLight) + Section 7 (Texture).
- **Docs**: `bif_migration_guide.md` Phase 1.3, texture section; refer to Go `material.go`, `texture.go`.
- **Learning focus**: Random sampling utilities, trait object ergonomics, error handling for IO textures.
- **Tasks**:
  - Complete texture trait implementations (Solid, Checker, Image, Noise) using the `image` crate.
  - Flesh out all four baseline materials and ensure `scatter` semantics match Go.
- **Validation**: Scene-level tests that render a Cornell box variant at low resolution and compare histograms against Go output (within tolerance).

### Stage 4 – BVH, Camera, and CPU Renderer Skeleton (1-2 weeks)

- **Checklist refs**: Sections 12, 17-19, 18 specifically for bucket renderer.
- **Docs**: `bif_migration_guide.md` Phase 2.2 & 3.1, devlog bucket renderer snippet, PoC Step 5.
- **Learning focus**: Recursion + ownership, concurrency via `rayon`, builder patterns.
- **Tasks**:
  - Port BVH builder using `Vec<Arc<dyn Hittable>>`; add property tests for bounding boxes.
  - Implement camera builder with DOF/motion blur parity.
  - Build a minimal renderer crate that can render a static scene using scanline first, then upgrade to bucket rendering.
  - **Viewport-as-Framebuffer**: The renderer writes directly to a wgpu texture that egui displays. No separate "render window"—the viewport IS the framebuffer. This unifies interactive preview and final render.
  - Implement progressive refinement: low-SPP preview while moving camera, accumulate samples when idle.
- **Validation**: Render `rt/scenes/weekend_final.go` equivalent at 400×225 in the viewport, compare sample counts/time vs Go; record numbers in devlog and check off Checklist §§12,17,18.

> **Viewport = Framebuffer**: Unlike traditional DCCs with separate render views, BIF renders directly into the 3D viewport. When you orbit the camera, you see low-quality preview; when you stop, samples accumulate in-place. This is the "continuous rendering" philosophy from `bif_gpu_rendering_architecture.md`.

### Stage 5 – Scene Graph, Instancing & Scatter System (1-2 weeks)

- **Checklist refs**: Portions of Sections 18-21 + Code Organization §32 + new Scatter section.
- **Docs**: `port/README.md` "Phase 2: Scene Assembly", `bif_poc_guide.md` Steps 2-4, devlog ownership notes.
- **Learning focus**: `Arc<dyn Trait>` management, data-oriented layout, serialization, procedural generation.
- **Tasks**:
  - Implement a `scene` crate that hosts prototypes, layers, and instancer data structures.
  - Add importers (OBJ/tobj) and tie them into BVH construction.
  - Author a USD stub or JSON scene file to validate instancing logic before touching real USD.
  - **Scatter System** (`crates/scatter/`):
    - Surface point sampling: generate points on mesh surfaces (uniform, density-mapped)
    - Instance placement: scatter prototype meshes at generated points
    - Simple node graph or parameter UI in egui for controlling scatter (density, scale variance, rotation randomness)
    - Real-time viewport preview of scatter points before committing
- **Validation**: CLI that instantiates 10k spheres and reports memory/time (targets from README). Scatter 1000 trees on a ground plane via the UI. Document metrics.

### Stage 6 – USD, UI Polish & Node Editor (2+ weeks)

- **Checklist refs**: Migration Strategy Phases 6-8, Sections 20-24, Testing §§25-26, UI sections.
- **Docs**: `bif_poc_guide.md` Step 4, PoC Step 8 (build & run), GPU architecture doc for eventual viewport coupling.
- **Learning focus**: FFI (`cxx`, `usd_bridge`), error handling with `anyhow/thiserror`, integration tests, egui custom widgets.
- **Tasks**:
  - Port/export minimal USD via the shim, hooking into the scene graph.
  - Stand up integration tests that compare exported USD against fixtures.
  - **egui UI Panels**:
    - Scene hierarchy panel (tree view of prototypes and instances)
    - Properties panel (transform, material assignment)
    - Scatter controls panel (density slider, randomization seeds, preview toggle)
  - **Simple Node Editor** (optional, can use `egui_node_graph` crate):
    - Visual graph for scatter rules: Surface → Sample Points → Filter → Place Instances
    - Nodes for: mesh input, point sampler, density mask, instance placer
    - Real-time preview updates as nodes connect
- **Validation**: USD file opens in `usdview` with correct PointInstancer; viewport displays 60 FPS with egui overlay; scatter graph produces visible instances.

### Stage 7 – Advanced Rendering, GPU Path & Qt Migration (ongoing)

- **Checklist refs**: Advanced Features §§27-29, Optimizations §§22-24, Performance Targets, GPU architecture doc.
- **Docs**: `bif_gpu_rendering_architecture.md`, PoC Step 6, devlog GPU notes.
- **Learning focus**: Hybrid rendering coordinator, progressive accumulation, profiling, Qt/Rust FFI.
- **Tasks**:
  - Introduce `RenderCoordinator` state machine from GPU doc.
  - Bridge CPU renderer outputs to viewport for progressive updates.
  - Profile with `cargo flamegraph`, capture findings in devlog, and iterate.
  - **Qt Migration** (post-PoC):
    - Replace egui with Qt 6 / QML frontend using `cxx-qt`
    - Native Qt viewport widget with embedded wgpu surface
    - Qt-based node editor (Qt Node Editor or custom QGraphicsScene)
    - Dockable panels, keyboard shortcuts, professional UX
- **Validation**: Achieve <1s preview pass for 600×600 image; document FPS/sample rates over time. Qt prototype shows same functionality as egui version.

## Unified Viewport + Framebuffer Surface

**Core Principle:** The 3D viewport IS the render framebuffer. There is no separate "Render View" window.

- The CPU/GPU renderer writes directly to a wgpu texture that the viewport displays.
- egui (PoC) or Qt (production) renders UI overlays (gizmos, selection outlines, scatter previews) on top of this texture.
- When you orbit the camera, the renderer switches to low-SPP preview mode.
- When interaction stops, samples accumulate in-place—same viewport, same texture.
- "Render" button simply lets the accumulation run to target SPP, then optionally saves to disk.

### Viewport Responsibilities:
- Camera manipulation (orbit, pan, zoom, fly)
- Selection highlighting and transform gizmos
- Scatter point preview (before committing)
- Progress overlay (samples, time, pass info)

### Renderer Responsibilities:
- Ray tracing into the shared texture
- Sample accumulation and tone mapping
- Mode switching (interactive ↔ progressive ↔ final)
- AOV output (when saving to disk)

### Snapshot Mode:
- If you need to freeze a reference render while continuing to work, snapshot the current framebuffer to a separate texture.
- This "pinned render" can be displayed in a split view or saved immediately.

## Maintenance Loop

1. **After each stage** update `RUST_PORT_CHECKLIST.md` (checked boxes + notes) and append an entry to `devlog/DEVLOG_*.md` with screenshots or metrics.
2. **Retrospective**: capture what you learned about Rust (ownership trick, lifetime insight) so the port doubles as study notes.
3. **Housekeeping**: once a stage completes, archive experimental Claude-generated code under `devlog/sandboxes/` (if you want to keep it) to keep `src/` clean.

## Suggested Weekly Rhythm

- **Day 1**: Review Go reference + plan tests.
- **Day 2-3**: Implement Rust version with small commits tied to checklist items.
- **Day 4**: Write/extend tests, run comparisons, update docs.
- **Day 5**: Refactor or spike on next stage’s unknowns; log findings.

Following this cadence should keep the port deliberate, help you internalize Rust, and leave an auditable trail for every subsystem you migrate.
