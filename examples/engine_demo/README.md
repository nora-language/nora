# Nora Engine Demo

A 3D engine demonstration built with the **Nora Programming Language**. 

This project showcases the integration of several core systems to create a functional 3D rendering and physics environment.

## Features
- **WebGPU Rendering**: Utilizes a custom C-backend wrapper (`native.c`) to interface with WebGPU, demonstrating shader pipelines (`shader.wgsl`), depth testing, backface culling, and vertex-colored 3D meshes.
- **Jolt Physics**: Integrates the Jolt physics engine for rigid body dynamics. Features a simulation loop where a 3D cube falls under the influence of gravity and interacts with a floor.
- **Entity Component System (ECS)**: Uses `port_gecs` to manage entities, cleanly separating `Transform` components, `RigidBody` components, and `MeshRenderer` components.
- **Math Library**: Includes a custom matrix and vector math library written in Nora for handling perspective projection, look-at cameras, and 3D model transformations.

## Project Structure
- `src/main.nr` - The engine entry point, ECS initialization, system registration, and main game loop.
- `src/core/math.nr` - Vector and Matrix math utilities.
- `src/features/physics/` - Jolt physics C-bindings and physics ECS systems.
- `src/features/rendering/` - WebGPU initialization, pipeline setup, WebGPU C-bindings, shaders, and mesh definitions.
- `nora.yaml` - Project manifest and dependencies.

## Building and Running
To build and run this demo, navigate to the `engine_demo` directory and run the following commands with the Nora compiler:

```bash
# Build the project in release mode
nora build -r

# Run the compiled executable
nora run
```
