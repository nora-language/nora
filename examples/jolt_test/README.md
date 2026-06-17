# JoltPhysics 5.5.0 Integration Example

This directory contains a complete, working example of integrating a modern C++17 library ([JoltPhysics](https://github.com/jrouwe/JoltPhysics)) into the **Nora Programming Language**.

Since Nora compiles directly to C11 and doesn't natively support C++ ABIs, this example demonstrates how to create a bridging C-wrapper compiled with `clang++` that links against the pre-compiled C++ `.lib` and exposes an opaque C Application Binary Interface (ABI) to Nora.

## Prerequisites

To build and run this example on Windows, you must have the following installed:
1. **Nora Compiler Toolchain** (`nora_real`)
2. **CMake** (in your PATH) for compiling JoltPhysics.
3. **Visual Studio C++ Build Tools** (specifically the MSVC compiler `cl.exe`, which CMake defaults to on Windows).
4. **Clang/LLVM for Windows** (`clang++` must be in your PATH and configured to use the MSVC ABI).

## 1. Setup and Build

Before running the Nora application, you must fetch the JoltPhysics repository, compile it into a static library (`.lib`), and compile our C++ wrapper (`src/wrapper.cpp`) into an object file (`.obj`).

We have provided a convenient batch script to automate this entire process:

```cmd
cd examples/jolt_test
.\setup.bat
```

### What `setup.bat` does:
1. **Clones** the JoltPhysics repository.
2. **Checks out** the stable `v5.5.0` tag.
3. **Builds** `Jolt.lib` using CMake in `Release` mode (which defaults to using MSVC on Windows).
4. **Compiles** `src/wrapper.cpp` into `lib/wrapper.obj` using `clang++`. 
   - *Note:* The script explicitly uses `-O3 -DNDEBUG` alongside AVX2 flags to ensure the standard library (`std::vector`, etc.) memory layout inside our wrapper perfectly matches the MSVC `Release` build layout of `Jolt.lib`, avoiding catastrophic ABI mismatches and stack corruption.

## 2. Running the Example

Once the `wrapper.obj` and `Jolt.lib` files are successfully placed in the `lib/` folder, you can run the simulation natively via the Nora CLI.

Ensure your `NORA_STD_PATH` environment variable is set to the absolute path of the `std` library directory, then run:

```cmd
nora_real run src\main.nr
```

*(Alternatively, if your CLI is just `nora`):*
```cmd
nora run src\main.nr
```

### Expected Output
The application initializes the physics world, creates a static floor, and drops a dynamic sphere from `Y = 10.0`. It steps the physics simulation 60 times (simulating 1 second of time at 60 FPS) and prints the position of the sphere as it accelerates downward:

```text
Initializing JoltPhysics via Custom C++ Wrapper...
System initialized.
Floor added successfully.
Sphere added successfully.
Created physics bodies and optimized broadphase.
Step 0: Sphere Y Position = 9.997491
Step 10: Sphere Y Position = 9.834825
Step 20: Sphere Y Position = 9.423423
Step 30: Sphere Y Position = 8.765265
Step 40: Sphere Y Position = 7.862319
Step 50: Sphere Y Position = 6.716537
Simulation complete!
```

## How It Works

### The C++ Wrapper (`src/wrapper.cpp`)
JoltPhysics utilizes heavy C++ paradigms (templates, inheritance, overloaded operators). We shield Nora from this complexity by defining `extern "C"` functions that take opaque pointers (`JoltC_System*`, `JoltC_Body*`).
- **`jolt_init`**: Handles allocating Jolt's TempAllocator, JobSystem, PhysicsSystem, and standard BroadPhaseLayer interfaces.
- **`jolt_create_floor` / `jolt_create_sphere`**: Translates simple primitive parameters into Jolt's `BodyCreationSettings`.
- **`jolt_step`**: Triggers `sys->physics_system->Update(deltaTime, ...)`.

### The Nora Manifest (`nora.yaml`)
To link everything together, `nora.yaml` defines the necessary C headers, the object files, and the static library for the generated C11 code:
```yaml
native:
  headers: ["src/wrapper.h"]
  cflags: ["-Isrc", "lib/wrapper.obj", "lib/Jolt.lib"]
```

### The Nora Entrypoint (`src/main.nr`)
In Nora, we simply declare the external C functions using `extern fn` bindings and call them just like native Nora code. Memory for the system is managed opaquely through the `ptr` type, passed around by value.

```nora
[include("src/wrapper.h")]
extern fn jolt_init() ptr

[include("src/wrapper.h")]
extern fn jolt_step(system: ptr, deltaTime: f32)
```
