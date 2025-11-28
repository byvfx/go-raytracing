# Rust Port Learning Plan

This document reorganizes the existing porting material in `port/` into a step-by-step loop that lets you learn Rust while keeping the Go raytracer as a reference. Every stage ends with a concrete validation gate so the migration never drifts into "messy" territory.

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
  - Recreate the minimal Cargo workspace (root + `crates/math`, `crates/renderer`, etc.) per PoC guide.
  - Configure toolchain (`rustup`, `cargo fmt/clippy`, VS Code rust-analyzer).
  - Copy small Go fixtures (e.g., unit test scenes) into `assets/` for later comparisons.
- **Done when**: `cargo check` succeeds and a `tests/smoke.rs` (even empty) runs.

### Stage 1 – Core Math & Utility Types (1-2 days)

- **Checklist refs**: Sections 1-4 (Vec3, Ray, Interval, AABB).
- **Docs**: `bif_migration_guide.md` Phase 1.1, Go files `rt/vec3.go`, `rt/ray.go`, `rt/interval.go`, `rt/aabb.go`.
- **Learning focus**: Ownership basics, trait derivations, unit tests with `approx` or custom asserts.
- **Deliverables**:
  - `crates/math` with fully tested vec3/point/color types and helper functions.
  - Benchmark or quick comparison that matches Go outputs for dot/cross operations.
- **Validation**: `cargo test -p math` comparing expected values; document results in devlog.

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
  - Design a CPU framebuffer module (image-sized accumulation buffer + tone mapping) that can back both offline exports and an on-screen viewport surface. This becomes a separate `crates/framebuffer` crate post-PoC when integrating progressive refinement.
- **Validation**: Render `rt/scenes/weekend_final.go` equivalent at 400×225, compare sample counts/time vs Go; record numbers in devlog and check off Checklist §§12,17,18.

> **Note:** The PoC (`bif_poc_guide.md`) defers the `framebuffer` crate to keep the initial workspace lean. Add it here when you need shared accumulation between viewport and offline rendering.

### Stage 5 – Scene Graph & Instancing (1-2 weeks)

- **Checklist refs**: Portions of Sections 18-21 + Code Organization §32.
- **Docs**: `port/README.md` “Phase 2: Scene Assembly”, `bif_poc_guide.md` Steps 2-4, devlog ownership notes.
- **Learning focus**: `Arc<dyn Trait>` management, data-oriented layout, serialization.
- **Tasks**:
  - Implement a `scene` crate that hosts prototypes, layers, and instancer data structures.
  - Add importers (OBJ/tobj) and tie them into BVH construction.
  - Author a USD stub or JSON scene file to validate instancing logic before touching real USD.
- **Validation**: CLI that instantiates 10k spheres and reports memory/time (targets from README). Document metrics.

### Stage 6 – USD + Tooling Integration (2+ weeks)

- **Checklist refs**: Migration Strategy Phases 6-8, Sections 20-24, Testing §§25-26.
- **Docs**: `bif_poc_guide.md` Step 4, PoC Step 8 (build & run), GPU architecture doc for eventual viewport coupling.
- **Learning focus**: FFI (`cxx`, `usd_bridge`), error handling with `anyhow/thiserror`, integration tests.
- **Tasks**:
  - Port/export minimal USD via the shim, hooking into the scene graph.
  - Stand up integration tests that compare exported USD against fixtures.
  - Begin wiring viewport (wgpu) for bounding-box visual feedback.
- **Validation**: USD file opens in `usdview` with correct PointInstancer; viewport displays 60 FPS bounding boxes.

### Stage 7 – Advanced Rendering & GPU Path (ongoing)

- **Checklist refs**: Advanced Features §§27-29, Optimizations §§22-24, Performance Targets, GPU architecture doc.
- **Docs**: `bif_gpu_rendering_architecture.md`, PoC Step 6, devlog GPU notes.
- **Learning focus**: Hybrid rendering coordinator, progressive accumulation, profiling.
- **Tasks**:
  - Introduce `RenderCoordinator` state machine from GPU doc.
  - Bridge CPU renderer outputs to viewport for progressive updates.
  - Profile with `cargo flamegraph`, capture findings in devlog, and iterate.
- **Validation**: Achieve <1s preview pass for 600×600 image; document FPS/sample rates over time.

## Unified Viewport + Framebuffer Surface

- Keep the CPU framebuffer as a first-class crate (Stage 4 deliverable). It still owns accumulation, sample counting, AOV exports, and file output, but it now also exposes a texture handle (or shared buffer) suitable for viewport presentation.
- The viewport layer (wgpu/egui or Qt) renders manipulators, gizmos, and overlays directly atop the framebuffer image. When you orbit the camera or drag lights, the viewport requests quick preview passes that write into the same framebuffer memory, so what you see in the editor is identical to the offline output.
- Provide a thin messaging/API boundary—e.g., the viewport drives camera transforms and enqueue render jobs, while the renderer publishes swapchain-ready images plus histograms/progress. This keeps responsibilities clear even though both sides share the final surface.
- If you need to freeze a reference render, snapshot the framebuffer and detach it from the live viewport texture; otherwise, the default mode is a unified surface that blends interactive placement with progressive refinement.

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
