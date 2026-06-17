package rendering

import "ffi"
import "string"
import "gecs"
import "io"
import "core"
import "collections"
import "strconv"

pub type MeshRenderer = struct {
    vertex_buffer: ptr,
    vertex_count: u32,
    model_buffer: ptr,
    model_bind_group: ptr
}

pub type RenderSystem = struct {
    window_handle: #GLFWwindow,
    instance: ptr,
    surface: ptr,
    adapter: ptr,
    device: ptr,
    queue: ptr,
    clear_color: f32,
    pipeline: ptr,
    depth_view: ptr,
    camera_buffer: ptr,
    camera_bind_group: ptr,
    model_layout: ptr
}

extern fn nr_duplicate_ptr(p: ptr) ptr

pub fn NewRenderSystem(handle: #GLFWwindow) RenderSystem {
    return RenderSystem {
        window_handle: handle,
        instance: ptr(0),
        surface: ptr(0),
        adapter: ptr(0),
        device: ptr(0),
        queue: ptr(0),
        clear_color: f32(0.1),
        pipeline: ptr(0),
        depth_view: ptr(0),
        camera_buffer: ptr(0),
        camera_bind_group: ptr(0),
        model_layout: ptr(0)
    }
}

pub type AdapterCallback = struct { cb: fn(i32, ptr, WGPUStringView) }
pub type DeviceCallback = struct { cb: fn(i32, ptr, WGPUStringView) }

export fn adapter_trampoline(status: i32, adapter: ptr, message: WGPUStringView, userdata: ptr) {
    var wrapper = ffi.FromRaw[AdapterCallback](userdata)
    wrapper.cb(status, adapter, message)
}

export fn device_trampoline(status: i32, device: ptr, message: WGPUStringView, userdata: ptr) {
    var wrapper = ffi.FromRaw[DeviceCallback](userdata)
    wrapper.cb(status, device, message)
}

extern fn nr_print_string_view(data: ptr, length: u64)

export fn error_stateless_cb(_device: ptr, _err_type: i32, _message: WGPUStringView, _userdata1: ptr, _userdata2: ptr) {
    nr_print_string_view(_message.data, _message.length)
}

pub type CmdBufferContainer = struct { cmd: ptr }

var global_adapter = make(ptr[], 1)
var global_device = make(ptr[], 1)

pub fn (self: &RenderSystem) Init(_w: &gecs.World) {
    self.clear_color = f32(0.3)
    
    // 1. Instance
    var desc = alloc WGPUInstanceDescriptor { nextInChain: ptr(0), requiredFeatureCount: 0, requiredFeatures: ptr(0), requiredLimits: ptr(0) }
    self.instance = wgpuCreateInstance(#desc)
    
    // 2. Surface
    self.surface = nr_create_surface(self.instance, self.window_handle)
    
    var adapter_opts = alloc WGPURequestAdapterOptions {
        nextInChain: ptr(0),
        compatibleSurface: self.surface,
        powerPreference: 0,
        forceFallbackAdapter: 0,
        featureLevel: 0,
        backendType: 0
    }
    
    var on_adapter = fn(status: i32, adapter: ptr, message: WGPUStringView) {
        global_adapter[0] = adapter
    }
    var adapter_wrapper = alloc AdapterCallback { cb: on_adapter }
    var adapter_cb = ptr(adapter_trampoline)
    var adapter_info = WGPURequestAdapterCallbackInfo {
        nextInChain: ptr(0),
        mode: 0,
        callback: adapter_cb,
        userdata1: ffi.IntoRaw[AdapterCallback](adapter_wrapper),
        userdata2: ptr(0)
    }
    wgpuInstanceRequestAdapter(self.instance, #adapter_opts, adapter_info)
    self.adapter = global_adapter[0]
    
    var err_cb = ptr(error_stateless_cb)
    var dev_desc = alloc WGPUDeviceDescriptor {
        nextInChain: ptr(0),
        label: WGPUStringView { data: ptr(0), length: 0 },
        requiredFeatureCount: 0,
        requiredFeatures: ptr(0),
        requiredLimits: ptr(0),
        defaultQueue: alloc WGPUQueueDescriptor {
            nextInChain: ptr(0),
            label: WGPUStringView { data: ptr(0), length: 0 }
        },
        deviceLostCallbackInfo: alloc WGPUDeviceLostCallbackInfo {
            nextInChain: ptr(0), mode: 0, callback: ptr(0), userdata1: ptr(0), userdata2: ptr(0)
        },
        uncapturedErrorCallbackInfo: WGPUUncapturedErrorCallbackInfo {
            nextInChain: ptr(0), callback: err_cb, userdata1: ptr(0), userdata2: ptr(0)
        }
    }
    
    var on_device = fn(status: i32, device: ptr, message: WGPUStringView) {
        global_device[0] = device
    }
    var device_wrapper = alloc DeviceCallback { cb: on_device }
    var device_cb = ptr(device_trampoline)
    var device_info = WGPURequestDeviceCallbackInfo {
        nextInChain: ptr(0), mode: 0, callback: device_cb, userdata1: ffi.IntoRaw[DeviceCallback](device_wrapper), userdata2: ptr(0)
    }
    wgpuAdapterRequestDevice(self.adapter, #dev_desc, device_info)
    self.device = global_device[0]
    self.queue = wgpuDeviceGetQueue(self.device)
    
    var config = alloc WGPUSurfaceConfiguration {
        nextInChain: ptr(0), device: self.device, format: 0x1B, usage: 0x00000010,
        width: 800, height: 600, viewFormatCount: 0, viewFormats: ptr(0),
        alphaMode: 1, presentMode: 1
    }
    wgpuSurfaceConfigure(self.surface, #config)

    self.depth_view = nr_create_depth_texture_view(self.device, 800, 600)
    
    var camera_layout = nr_create_bind_group_layout_1_uniform(self.device, 1) // Vertex (1)
    self.model_layout = nr_create_bind_group_layout_1_uniform(self.device, 1) // Vertex (1)
    var pipeline_layout = nr_create_pipeline_layout_2_bind_groups(self.device, camera_layout, self.model_layout)
    
    var wgsl_code = "@vertex\nfn vs_main(model_in: VertexInput) -> VertexOutput {\n    var out: VertexOutput;\n    out.color = model_in.color;\n    out.normal = model_in.normal;\n    out.clip_position = camera.view_proj * model.model * vec4<f32>(model_in.position, 1.0);\n    return out;\n}\n\n@fragment\nfn fs_main(in: VertexOutput) -> @location(0) vec4<f32> {\n    let light_dir = normalize(vec3<f32>(0.5, 1.0, 0.2));\n    let normal = normalize(in.normal);\n    let diffuse = max(dot(normal, light_dir), 0.1);\n    let final_color = in.color * diffuse;\n    return vec4<f32>(final_color, 1.0);\n}\n\nstruct CameraUniform { view_proj: mat4x4<f32> };\n@group(0) @binding(0) var<uniform> camera: CameraUniform;\nstruct ModelUniform { model: mat4x4<f32> };\n@group(1) @binding(0) var<uniform> model: ModelUniform;\nstruct VertexInput { @location(0) position: vec3<f32>, @location(1) color: vec3<f32>, @location(2) normal: vec3<f32> };\nstruct VertexOutput { @builtin(position) clip_position: vec4<f32>, @location(0) color: vec3<f32>, @location(1) normal: vec3<f32> };"
    
    var wgsl_desc = alloc WGPUShaderSourceWGSL {
        chain: WGPUChainedStruct { next: ptr(0), sType: 2 },
        code: WGPUStringView { data: ptr(wgsl_code), length: u64(len(wgsl_code)) }
    }
    wgsl_desc.code.length = u64(0xFFFFFFFFFFFFFFFF)
    
    var shader_desc = alloc WGPUShaderModuleDescriptor {
        nextInChain: #wgsl_desc, label: WGPUStringView { data: ptr(0), length: 0 }
    }
    var shader_module = wgpuDeviceCreateShaderModule(self.device, #shader_desc)

    self.pipeline = nr_create_render_pipeline(self.device, pipeline_layout, shader_module)

    self.camera_buffer = nr_create_buffer(self.device, 64, 0x00000040 | 0x00000008, 0) // Uniform | CopyDst
    self.camera_bind_group = nr_create_bind_group_1_uniform(self.device, camera_layout, self.camera_buffer, 64)

    wgpuShaderModuleRelease(shader_module)
    io.PrintLn("WebGPU Initialized.")
}

pub fn (self: &RenderSystem) CreateMeshRenderer() @MeshRenderer {
    var verts = CreateCubeVertices()
    var v_buf = nr_create_buffer(self.device, u64(36 * 36), 0x00000020 | 0x00000008, 0) // Vertex
    nr_queue_write_buffer(self.queue, v_buf, 0, ptr(#verts[0]), u64(36 * 36))

    var m_buf = nr_create_buffer(self.device, 64, 0x00000040 | 0x00000008, 0) // Uniform | CopyDst
    var m_bg = nr_create_bind_group_1_uniform(self.device, self.model_layout, m_buf, 64)

    return alloc MeshRenderer {
        vertex_buffer: v_buf,
        vertex_count: 36,
        model_buffer: m_buf,
        model_bind_group: m_bg
    }
}

pub fn UpdateCamera(queue: ptr, camera_buffer: ptr) {
    var _proj = core.Mat4Perspective(f32(1.5707), f32(800.0/600.0), f32(0.1), f32(100.0))
    var _view = core.Mat4LookAt(f32(0.0), f32(15.0), f32(30.0), f32(0.0), f32(0.0), f32(0.0), f32(0.0), f32(1.0), f32(0.0))
    var view_proj = core.Mat4Multiply(_view, _proj)
    nr_queue_write_buffer(queue, camera_buffer, 0, #view_proj, 64)
}

pub fn DrawMeshes(w: &gecs.World, pass: ptr, queue: ptr, pipeline: ptr, camera_bg: ptr) {

    

    wgpuRenderPassEncoderSetPipeline(pass, pipeline)
    

    wgpuRenderPassEncoderSetBindGroup(pass, 0, camera_bg, 0, ptr(0))

    var mesh_query = gecs.NewQuery2[gecs.Transform, MeshRenderer](w, 1, 3)

    var get_mesh = fn(entity: gecs.Entity, trans: &gecs.Transform, mesh: &MeshRenderer) {
        var _tr = core.Mat4Translation(f32(trans.position.x), f32(trans.position.y), f32(trans.position.z))
        var _sc = core.Mat4Scale(f32(trans.scale.x), f32(trans.scale.y), f32(trans.scale.z))
        var model_mat = core.Mat4Multiply(_sc, _tr)
        
        var q_copy = nr_duplicate_ptr(queue)
        nr_queue_write_buffer(q_copy, mesh.model_buffer, 0, #model_mat, 64)

        var p_copy = nr_duplicate_ptr(pass)
        wgpuRenderPassEncoderSetBindGroup(p_copy, 1, mesh.model_bind_group, 0, ptr(0))
        wgpuRenderPassEncoderSetVertexBuffer(p_copy, 0, mesh.vertex_buffer, 0, u64(mesh.vertex_count * 36))
        wgpuRenderPassEncoderDraw(p_copy, u32(mesh.vertex_count), 1, 0, 0)
    }

    mesh_query.Each[gecs.Transform, MeshRenderer](get_mesh)
}

pub fn (self: &RenderSystem) Update(w: &gecs.World) {
    var query = gecs.NewQuery1[gecs.Transform](w, 1)
    var get_pos = fn(entity: gecs.Entity, trans: &gecs.Transform) {

        self.clear_color = f32(trans.position.y) / f32(10.0)
    }
    query.Each[gecs.Transform](get_pos)

    UpdateCamera(nr_duplicate_ptr(self.queue), nr_duplicate_ptr(self.camera_buffer))

    var surface_tex = alloc WGPUSurfaceTexture { nextInChain: ptr(0), texture: ptr(0), suboptimal: 0, status: 0 }
    wgpuSurfaceGetCurrentTexture(nr_duplicate_ptr(self.surface), #surface_tex)
    if (surface_tex.texture == ptr(0)) { return }

    var view_tex = wgpuTextureCreateView(surface_tex.texture, ptr(0))
    var encoder_desc = alloc WGPUCommandEncoderDescriptor { nextInChain: ptr(0), label: WGPUStringView { data: ptr(0), length: 0 } }
    var encoder = wgpuDeviceCreateCommandEncoder(nr_duplicate_ptr(self.device), #encoder_desc)

    var color_att = alloc WGPURenderPassColorAttachment {
        nextInChain: ptr(0), view: view_tex, depthSlice: 0xFFFFFFFF, resolveTarget: ptr(0),
        loadOp: 2, storeOp: 1, clearValue: WGPUColor { r: f64(self.clear_color), g: f64(self.clear_color), b: 0.3, a: 1.0 }
    }
    
    var depth_att = alloc WGPURenderPassDepthStencilAttachment {
        nextInChain: ptr(0), view: nr_duplicate_ptr(self.depth_view),
        depthLoadOp: 2, depthStoreOp: 1, depthClearValue: f32(1.0), depthReadOnly: 0,
        stencilLoadOp: 0, stencilStoreOp: 0, stencilClearValue: 0, stencilReadOnly: 0
    }
    
    var pass_desc = alloc WGPURenderPassDescriptor {
        nextInChain: ptr(0), label: WGPUStringView { data: ptr(0), length: 0 },
        colorAttachmentCount: 1, colorAttachments: #color_att,
        depthStencilAttachment: #depth_att, occlusionQuerySet: ptr(0), timestampWrites: ptr(0)
    }

    var pass = wgpuCommandEncoderBeginRenderPass(encoder, #pass_desc)
    
    DrawMeshes(w, pass, nr_duplicate_ptr(self.queue), nr_duplicate_ptr(self.pipeline), nr_duplicate_ptr(self.camera_bind_group))
    
    wgpuRenderPassEncoderEnd(pass)
    var cmd_buffer = wgpuCommandEncoderFinish(encoder, ptr(0))
    var cmd_cont = alloc CmdBufferContainer { cmd: cmd_buffer }
    wgpuQueueSubmit(nr_duplicate_ptr(self.queue), 1, #cmd_cont)

    wgpuSurfacePresent(nr_duplicate_ptr(self.surface))
    wgpuTextureViewRelease(view_tex)
    wgpuCommandBufferRelease(cmd_buffer)
    wgpuRenderPassEncoderRelease(pass)
    wgpuCommandEncoderRelease(encoder)
    wgpuTextureRelease(surface_tex.texture)
}
