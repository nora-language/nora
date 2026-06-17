package main

import "ffi"
import "io"
import "runtime"

[include("<GLFW/glfw3.h>")]
type GLFWwindow = struct {
}

[include("<GLFW/glfw3.h>")]
extern fn glfwInit() i32

extern fn glfwCreateWindow(width: i32, height: i32, title: ptr, monitor: ptr, share: ptr) #GLFWwindow

[include("<GLFW/glfw3.h>")]
extern fn glfwMakeContextCurrent(window: #GLFWwindow)

[include("<GLFW/glfw3.h>")]
extern fn glfwWindowShouldClose(window: #GLFWwindow) i32

[include("<GLFW/glfw3.h>")]
extern fn glfwSwapBuffers(window: #GLFWwindow)

[include("<GLFW/glfw3.h>")]
extern fn glfwPollEvents()

[include("<GLFW/glfw3.h>")]
extern fn glfwTerminate()

[include("<GLFW/glfw3.h>")]
extern fn glfwDestroyWindow(window: #GLFWwindow)

[include("<GLFW/glfw3.h>")]
extern fn glfwWindowHint(hint: i32, value: i32)

// Nora Safe Wrapper
type Window = struct {
    handle: #GLFWwindow
}

fn make_window(w: i32, h: i32, title: ptr) Window {
    glfwInit()
    glfwWindowHint(0x00022001, 0) // GLFW_CLIENT_API = GLFW_NO_API
    var handle = glfwCreateWindow(w, h, title, none, none)
    return Window {
        handle: handle
    }
}

fn (self: #Window) should_close() bool {
    return glfwWindowShouldClose(self.handle) != 0
}

fn (self: #Window) swap_buffers() {
    glfwSwapBuffers(self.handle)
}

fn (self: #Window) poll_events() {
    glfwPollEvents()
}

fn (self: #Window) drop() {
    glfwDestroyWindow(self.handle)
    glfwTerminate()
}

// WebGPU Callback Trampolines

type CmdBufferContainer = struct {
    cmd: ptr
}

type AdapterCallback = struct {
    cb: fn(i32, ptr, WGPUStringView)
}

export fn adapter_trampoline(status: i32, adapter: ptr, message: WGPUStringView, ud1: ptr, ud2: ptr) {
    if (status != 1) {
        io.PrintLn("Adapter Request Failed!")
    }
    var wrapper = ffi.FromRaw[AdapterCallback](ud1)
    wrapper.cb(status, adapter, message)
}

type DeviceCallback = struct {
    cb: fn(i32, ptr, WGPUStringView)
}

export fn device_trampoline(status: i32, device: ptr, message: WGPUStringView, ud1: ptr, ud2: ptr) {
    if (status != 1) {
        io.PrintLn("Device Request Failed!")
    }
    var wrapper = ffi.FromRaw[DeviceCallback](ud1)
    wrapper.cb(status, device, message)
}

export fn error_stateless_cb(device: ptr, err_type: i32, message: WGPUStringView, ud1: ptr, ud2: ptr) {
    io.PrintLn("WebGPU Error!")
}

// Globals to bypass compiler double-free on local owned variable captures
var global_adapter = make(ptr[], 1)
var global_device = make(ptr[], 1)

fn main() {
    runtime.PinThread()
    // 0. Use raw string for title
    var c_title = ffi.CString("Nora + WebGPU")
    var win = make_window(800, 600, c_title)

    // 1. Instance
    var inst_desc = alloc WGPUInstanceDescriptor {
        nextInChain: none,
        requiredFeatureCount: 0,
        requiredFeatures: none,
        requiredLimits: none
    }
    var instance = wgpuCreateInstance(#inst_desc)

    // 2. Surface
    var surface = nr_create_surface(instance, win.handle)

    // 3. Adapter
    var adapter_opts = alloc WGPURequestAdapterOptions {
        nextInChain: none,
        featureLevel: 0,
        powerPreference: 0,
        forceFallbackAdapter: 0,
        backendType: 0,
        compatibleSurface: surface
    }

    // Use Nora Closure
    var on_adapter = fn(status: i32, adapter: ptr, message: WGPUStringView) {
        global_adapter[0] = adapter
    }
    var adapter_wrapper = alloc AdapterCallback {
        cb: on_adapter
    }

    var adapter_cb = ptr(adapter_trampoline)
    var adapter_info = WGPURequestAdapterCallbackInfo {
        nextInChain: none,
        mode: 0,
        callback: adapter_cb,
        userdata1: ffi.IntoRaw[AdapterCallback](adapter_wrapper),
        userdata2: none
    }
    wgpuInstanceRequestAdapter(instance, #adapter_opts, adapter_info)
    var adapter = global_adapter[0]

    // 4. Device
    var err_cb = ptr(error_stateless_cb)
    var dev_desc = alloc WGPUDeviceDescriptor {
        nextInChain: none,
        label: WGPUStringView {
            data: none,
            length: 0
        },
        requiredFeatureCount: 0,
        requiredFeatures: none,
        requiredLimits: none,
        defaultQueue: alloc WGPUQueueDescriptor {
            nextInChain: none,
            label: WGPUStringView {
                data: none,
                length: 0
            }
        },
        deviceLostCallbackInfo: alloc WGPUDeviceLostCallbackInfo {
            nextInChain: none,
            mode: 0,
            callback: none,
            userdata1: none,
            userdata2: none
        },
        uncapturedErrorCallbackInfo: WGPUUncapturedErrorCallbackInfo {
            nextInChain: none,
            callback: err_cb,
            userdata1: none,
            userdata2: none
        }
    }

    var on_device = fn(status: i32, device: ptr, message: WGPUStringView) {
        global_device[0] = device
    }
    var device_wrapper = alloc DeviceCallback {
        cb: on_device
    }

    var device_cb = ptr(device_trampoline)
    var device_info = WGPURequestDeviceCallbackInfo {
        nextInChain: none,
        mode: 0,
        callback: device_cb,
        userdata1: ffi.IntoRaw[DeviceCallback](device_wrapper),
        userdata2: none
    }
    wgpuAdapterRequestDevice(adapter, #dev_desc, device_info)
    var device = global_device[0]

    // 5. Queue
    var queue = wgpuDeviceGetQueue(device)

    // 6. Surface Configure
    var config = alloc WGPUSurfaceConfiguration {
        nextInChain: none,
        device: device,
        format: 0x1B, // WGPUTextureFormat_BGRA8Unorm
        usage: 0x00000010, // WGPUTextureUsage_RenderAttachment
        viewFormatCount: 0,
        viewFormats: none,
        alphaMode: 1, // WGPUCompositeAlphaMode_Opaque
        width: 800,
        height: 600,
        presentMode: 1 // WGPUPresentMode_Fifo
    }
    wgpuSurfaceConfigure(surface, #config)

    // 7. Shader Module
    var wgsl_code = "@vertex\nfn vs_main(@builtin(vertex_index) in_vertex_index: u32) -> @builtin(position) vec4<f32> {\n    let x = f32(i32(in_vertex_index) - 1);\n    let y = f32(i32(in_vertex_index & 1u) * 2 - 1);\n    return vec4<f32>(x, y, 0.0, 1.0);\n}\n\n@fragment\nfn fs_main() -> @location(0) vec4<f32> {\n    return vec4<f32>(1.0, 0.0, 0.0, 1.0);\n}"

    var wgsl_desc = alloc WGPUShaderSourceWGSL {
        chain: WGPUChainedStruct { // WGPUSType_ShaderSourceWGSL
            next: none,
            sType: 2
        },
        code: WGPUStringView {
            data: ptr(wgsl_code),
            length: u64(len(wgsl_code))
        }
    }
    wgsl_desc.code.length = u64(0xFFFFFFFFFFFFFFFF) // webgpu.h WGPU_WHOLE_SIZE

    var shader_desc = alloc WGPUShaderModuleDescriptor {
        nextInChain: #wgsl_desc,
        label: WGPUStringView {
            data: none,
            length: 0
        }
    }

    var shader_module = wgpuDeviceCreateShaderModule(device, #shader_desc)

    // 8. Pipeline
    var color_target = WGPUColorTargetState {
        nextInChain: none,
        format: 0x1B,
        blend: none,
        writeMask: 15
    }

    var fragment_state = WGPUFragmentState {
        nextInChain: none,
        module: shader_module,
        entryPoint: WGPUStringView {
            data: ptr("fs_main"),
            length: u64(len("fs_main"))
        },
        constantCount: 0,
        constants: none,
        targetCount: 1,
        targets: #color_target
    }

    var pipeline_desc = WGPURenderPipelineDescriptor {
        nextInChain: none,
        label: WGPUStringView {
            data: none,
            length: 0
        },
        layout: none,
        vertex: WGPUVertexState {
            nextInChain: none,
            module: shader_module,
            entryPoint: WGPUStringView {
                data: ptr("vs_main"),
                length: u64(len("vs_main"))
            },
            constantCount: 0,
            constants: none,
            bufferCount: 0,
            buffers: none
        },
        primitive: alloc WGPUPrimitiveState {
            nextInChain: none,
            topology: 4, // TriangleList
            stripIndexFormat: 0,
            frontFace: 0,
            cullMode: 0,
            unclippedDepth: 0
        },
        depthStencil: none,
        multisample: alloc WGPUMultisampleState {
            nextInChain: none,
            count: 1,
            mask: 0xFFFFFFFF,
            alphaToCoverageEnabled: 0
        },
        fragment: #fragment_state
    }

    var pipeline = wgpuDeviceCreateRenderPipeline(device, #pipeline_desc)

    // Main Loop
    while !win.should_close() {
        var surface_tex = alloc WGPUSurfaceTexture {
            nextInChain: none,
            texture: none,
            suboptimal: 0,
            status: 0
        }
        wgpuSurfaceGetCurrentTexture(surface, #surface_tex)

        var view = wgpuTextureCreateView(surface_tex.texture, none)

        var encoder_desc = alloc WGPUCommandEncoderDescriptor {
            nextInChain: none,
            label: WGPUStringView {
                data: none,
                length: 0
            }
        }
        var encoder = wgpuDeviceCreateCommandEncoder(device, #encoder_desc)

        var color_attachment = alloc WGPURenderPassColorAttachment {
            nextInChain: none,
            view: view,
            depthSlice: 0xFFFFFFFF, // WGPU_DEPTH_SLICE_UNDEFINED
            resolveTarget: none,
            loadOp: 2, // Clear
            storeOp: 1, // Store
            clearValue: WGPUColor {
                r: 0.1,
                g: 0.2,
                b: 0.3,
                a: 1.0
            }
        }

        var pass_desc = alloc WGPURenderPassDescriptor {
            nextInChain: none,
            label: WGPUStringView {
                data: none,
                length: 0
            },
            colorAttachmentCount: 1,
            colorAttachments: #color_attachment,
            depthStencilAttachment: none,
            occlusionQuerySet: none,
            timestampWrites: none
        }

        var pass = wgpuCommandEncoderBeginRenderPass(encoder, #pass_desc)
        wgpuRenderPassEncoderSetPipeline(pass, pipeline)
        wgpuRenderPassEncoderDraw(pass, 3, 1, 0, 0)
        wgpuRenderPassEncoderEnd(pass)

        var cmd_buffer = wgpuCommandEncoderFinish(encoder, none)

        var cmd_cont = alloc CmdBufferContainer {
            cmd: cmd_buffer
        }
        wgpuQueueSubmit(queue, 1, #cmd_cont)

        wgpuSurfacePresent(surface)

        wgpuTextureViewRelease(view)
        wgpuCommandBufferRelease(cmd_buffer)
        wgpuRenderPassEncoderRelease(pass)
        wgpuCommandEncoderRelease(encoder)
        wgpuTextureRelease(surface_tex.texture)

        win.poll_events()
    }
}
