# Integrating C++ Libraries with Nora

Nora compiles down to **C11**, which means it does not natively understand C++ features like classes, templates, namespaces, or function overloading. Because of this, you cannot use Nora's `[include("<library.hpp>")]` to directly import C++ headers. 

However, Nora is fully capable of interoperating with C++ libraries using a **C-API Wrapper boundary (`extern "C"`)**. 

This walkthrough outlines the architectural requirements and the step-by-step process to achieve seamless C++ integration.

---

## The Architecture of C++ Interop

When the Nora compiler builds your project, it transpiles `.nr` source files into `.c` files and feeds them directly into the underlying C compiler toolchain (typically Clang).

If you inject a C++ header into this pipeline:
1. Clang will attempt to compile it using its C frontend.
2. It will encounter C++ keywords (`class`, `public`, `template`) and instantly fail with syntax errors.

To solve this, you must **bridge the languages** by wrapping your C++ functionality into an "opaque" C interface.

---

## Step-by-Step Guide

### 1. Create a C-compatible Header (`wrapper.h`)
Define a header that only uses standard C constructs. You will use opaque pointers (often represented as incomplete structs or `void*`) to pass C++ object references around.

```c
// wrapper.h
#pragma once

#ifdef __cplusplus
extern "C" {
#endif

// Opaque struct representing our C++ class
typedef struct OpaqueCppClass OpaqueCppClass;

// C-compatible function declarations
OpaqueCppClass* my_cpp_create();
void my_cpp_do_work(OpaqueCppClass* instance, int value);
void my_cpp_destroy(OpaqueCppClass* instance);

#ifdef __cplusplus
}
#endif
```

### 2. Write the C++ Implementation (`wrapper.cpp`)
Inside your `.cpp` file, you have full access to C++ features and the C++ Standard Library. 

```cpp
// wrapper.cpp
#include "wrapper.h"
#include <vector>
#include <iostream>

// Your actual complex C++ class
class MyComplexCppClass {
public:
    std::vector<int> data;
    void DoWork(int val) {
        data.push_back(val);
        std::cout << "C++ vector size: " << data.size() << " (Added " << val << ")\n";
    }
};

extern "C" {
    OpaqueCppClass* my_cpp_create() {
        return reinterpret_cast<OpaqueCppClass*>(new MyComplexCppClass());
    }

    void my_cpp_do_work(OpaqueCppClass* instance, int value) {
        auto obj = reinterpret_cast<MyComplexCppClass*>(instance);
        obj->DoWork(value);
    }

    void my_cpp_destroy(OpaqueCppClass* instance) {
        delete reinterpret_cast<MyComplexCppClass*>(instance);
    }
}
```

### 3. Bind the Wrapper in Nora
In your `.nr` code, you use `[include()]` and `extern fn` to bind directly to the C-functions you defined.

```nora
// src/main.nr
package main

[include("wrapper.h")]
type OpaqueCppClass = struct {}

[include("wrapper.h")]
extern fn my_cpp_create() #OpaqueCppClass

[include("wrapper.h")]
extern fn my_cpp_do_work(instance: #OpaqueCppClass, val: i32)

[include("wrapper.h")]
extern fn my_cpp_destroy(instance: #OpaqueCppClass)

fn main() {
    var obj = my_cpp_create()
    my_cpp_do_work(obj, 42)
    my_cpp_do_work(obj, 100)
    my_cpp_destroy(obj)
}
```

---

## Building and Linking

How you build the project depends entirely on the C++ standard required by the external library.

### Scenario A: Standard C++ Usage (No strict C++ flags needed)
If your C++ code doesn't require modern C++ flags (e.g., `-std=c++17`), you can let Nora invoke Clang to compile both the C and C++ files simultaneously.

**nora.yaml:**
```yaml
name: cpp_integration
version: 1.0.0
entry: src/main.nr
output: demo_cpp.exe

native:
  source_files:
    - wrapper.cpp  # Clang will automatically detect this as C++
  cflags:
    - "-lstdc++"   # Tell the linker to link the C++ standard library (-lc++ on macOS)
```

### Scenario B: Modern C++ (C++17, C++20, etc.)
If your wrapper or library requires modern C++ flags like `-std=c++17`, **you cannot place `-std=c++17` in Nora's `cflags`**. Because Nora feeds both its generated C code and your `source_files` into the same compiler invocation, Clang will attempt to apply the C++ standard to the C files, resulting in an error (`error: invalid argument '-std=c++17' not allowed with 'C'`).

To bypass this, you must **pre-compile your C++ code** using your system's build tools (CMake, Make, or manually) into an object file (`.o`) or a static library (`.a` / `.lib`).

**1. Pre-compile:**
```bash
clang++ -std=c++17 -c wrapper.cpp -o wrapper.o
```

**2. Link the binary in `nora.yaml`:**
```yaml
name: cpp_integration
version: 1.0.0
entry: src/main.nr
output: demo_cpp.exe

native:
  # Link the pre-compiled C++ binary
  static_libs:
    - wrapper.o
  
  cflags:
    - "-lstdc++" # Link the C++ standard library
```

This is the recommended, production-ready approach. For larger integrations, you would use CMake to generate a static library containing all C++ dependencies and simply pass that `.lib` or `.a` file directly into Nora's `static_libs`.
