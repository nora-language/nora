package rendering

type WGPUChainedStruct = struct {
    next: ptr,
    sType: u32,
}

type WGPUStringView = struct {
    data: ptr,
    length: u64,
}

type WGPUSurfaceSourceWindowsHWND = struct {
    chain: WGPUChainedStruct,
    hinstance: ptr,
    hwnd: ptr,
}

type WGPUSurfaceDescriptor = struct {
    nextInChain: ptr,
    label: WGPUStringView,
}

type WGPUInstanceDescriptor = struct {
    nextInChain: ptr,
    requiredFeatureCount: u64,
    requiredFeatures: ptr,
    requiredLimits: ptr,
}

type WGPURequestAdapterOptions = struct {
    nextInChain: ptr,
    featureLevel: u32,
    powerPreference: u32,
    forceFallbackAdapter: u32, // WGPUBool
    backendType: u32,
    compatibleSurface: ptr,
}

type WGPUDeviceLostCallbackInfo = struct {
    nextInChain: ptr,
    mode: u32,
    callback: ptr,
    userdata1: ptr,
    userdata2: ptr,
}

type WGPURequestAdapterCallbackInfo = struct {
    nextInChain: ptr,
    mode: u32,
    callback: ptr,
    userdata1: ptr,
    userdata2: ptr,
}

type WGPURequestDeviceCallbackInfo = struct {
    nextInChain: ptr,
    mode: u32,
    callback: ptr,
    userdata1: ptr,
    userdata2: ptr,
}

type WGPUUncapturedErrorCallbackInfo = struct {
    nextInChain: ptr,
    callback: ptr,
    userdata1: ptr,
    userdata2: ptr,
}

type WGPUQueueDescriptor = struct {
    nextInChain: ptr,
    label: WGPUStringView,
}

type WGPUDeviceDescriptor = struct {
    nextInChain: ptr,
    label: WGPUStringView,
    requiredFeatureCount: u64,
    requiredFeatures: ptr,
    requiredLimits: ptr,
    defaultQueue: WGPUQueueDescriptor,
    deviceLostCallbackInfo: WGPUDeviceLostCallbackInfo,
    uncapturedErrorCallbackInfo: WGPUUncapturedErrorCallbackInfo,
}

type WGPUSurfaceConfiguration = struct {
    nextInChain: ptr,
    device: ptr,
    format: u32,
    usage: u64,
    width: u32,
    height: u32,
    viewFormatCount: u64,
    viewFormats: ptr,
    alphaMode: u32,
    presentMode: u32,
}

type WGPUShaderSourceWGSL = struct {
    chain: WGPUChainedStruct,
    code: WGPUStringView,
}

type WGPUShaderModuleDescriptor = struct {
    nextInChain: ptr,
    label: WGPUStringView,
}

type WGPUColor = struct {
    r: f64,
    g: f64,
    b: f64,
    a: f64,
}

type WGPUBlendComponent = struct {
    operation: u32,
    srcFactor: u32,
    dstFactor: u32,
}

type WGPUBlendState = struct {
    color: WGPUBlendComponent,
    alpha: WGPUBlendComponent,
}

type WGPUColorTargetState = struct {
    nextInChain: ptr,
    format: u32,
    blend: ptr, // WGPUBlendState const *
    writeMask: u64,
}

type WGPUFragmentState = struct {
    nextInChain: ptr,
    module: ptr,
    entryPoint: WGPUStringView,
    constantCount: u64,
    constants: ptr,
    targetCount: u64,
    targets: ptr, // WGPUColorTargetState const *
}

type WGPUVertexState = struct {
    nextInChain: ptr,
    module: ptr,
    entryPoint: WGPUStringView,
    constantCount: u64,
    constants: ptr,
    bufferCount: u64,
    buffers: ptr,
}

type WGPUPrimitiveState = struct {
    nextInChain: ptr,
    topology: u32,
    stripIndexFormat: u32,
    frontFace: u32,
    cullMode: u32,
    unclippedDepth: u32, // WGPUBool
}

type WGPUMultisampleState = struct {
    nextInChain: ptr,
    count: u32,
    mask: u32,
    alphaToCoverageEnabled: u32, // WGPUBool
}

type WGPURenderPipelineDescriptor = struct {
    nextInChain: ptr,
    label: WGPUStringView,
    layout: ptr,
    vertex: WGPUVertexState,
    primitive: WGPUPrimitiveState,
    depthStencil: ptr,
    multisample: WGPUMultisampleState,
    fragment: ptr,
}

type WGPURenderPassColorAttachment = struct {
    nextInChain: ptr,
    view: ptr,
    depthSlice: u32,
    resolveTarget: ptr,
    loadOp: u32,
    storeOp: u32,
    clearValue: WGPUColor,
}

type WGPURenderPassDepthStencilAttachment = struct {
    nextInChain: ptr,
    view: ptr,
    depthLoadOp: u32,
    depthStoreOp: u32,
    depthClearValue: f32,
    depthReadOnly: u32,
    stencilLoadOp: u32,
    stencilStoreOp: u32,
    stencilClearValue: u32,
    stencilReadOnly: u32,
}

type WGPURenderPassDescriptor = struct {
    nextInChain: ptr,
    label: WGPUStringView,
    colorAttachmentCount: u64,
    colorAttachments: ptr,
    depthStencilAttachment: ptr,
    occlusionQuerySet: ptr,
    timestampWrites: ptr,
}

type WGPUCommandEncoderDescriptor = struct {
    nextInChain: ptr,
    label: WGPUStringView,
}

type WGPUTextureViewDescriptor = struct {
    nextInChain: ptr,
    label: WGPUStringView,
    format: u32,
    dimension: u32,
    baseMipLevel: u32,
    mipLevelCount: u32,
    baseArrayLayer: u32,
    arrayLayerCount: u32,
    aspect: u32,
    usage: u64,
}

type WGPUSurfaceTexture = struct {
    nextInChain: ptr,
    texture: ptr,
    suboptimal: u32,
    status: u32,
}

// OS Functions
extern fn nr_get_hinstance() ptr
extern fn nr_get_hwnd(window: #GLFWwindow) ptr
extern fn nr_create_surface(instance: ptr, window: #GLFWwindow) ptr
extern fn nr_print_surface(surface: ptr)

// Helper Functions
extern fn nr_create_buffer(device: ptr, size: u64, usage: u32, mappedAtCreation: u32) ptr
extern fn nr_queue_write_buffer(queue: ptr, buffer: ptr, bufferOffset: u64, data: ptr, size: u64)
extern fn nr_create_bind_group_layout_1_uniform(device: ptr, visibility: u32) ptr
extern fn nr_create_bind_group_1_uniform(device: ptr, layout: ptr, buffer: ptr, size: u64) ptr
extern fn nr_create_depth_texture_view(device: ptr, width: u32, height: u32) ptr
extern fn nr_create_pipeline_layout_2_bind_groups(device: ptr, layout0: ptr, layout1: ptr) ptr
extern fn nr_create_render_pipeline(device: ptr, layout: ptr, shader_module: ptr) ptr

// WebGPU API
extern fn wgpuInstanceRequestAdapter(instance: ptr, options: ptr, callbackInfo: WGPURequestAdapterCallbackInfo)
extern fn wgpuAdapterRequestDevice(adapter: ptr, descriptor: ptr, callbackInfo: WGPURequestDeviceCallbackInfo)

extern fn wgpuCreateInstance(descriptor: ptr) ptr
extern fn wgpuInstanceCreateSurface(instance: ptr, descriptor: ptr) ptr
extern fn wgpuDeviceGetQueue(device: ptr) ptr
extern fn wgpuSurfaceGetCapabilities(surface: ptr, adapter: ptr, capabilities: ptr) u32
extern fn wgpuSurfaceCapabilitiesFreeMembers(capabilities: ptr)
extern fn wgpuSurfaceConfigure(surface: ptr, config: ptr)
extern fn wgpuSurfaceGetCurrentTexture(surface: ptr, surfaceTexture: ptr)
extern fn wgpuTextureCreateView(texture: ptr, descriptor: ptr) ptr
extern fn wgpuDeviceCreateShaderModule(device: ptr, descriptor: ptr) ptr
extern fn wgpuDeviceCreateRenderPipeline(device: ptr, descriptor: ptr) ptr
extern fn wgpuDeviceCreateCommandEncoder(device: ptr, descriptor: ptr) ptr
extern fn wgpuCommandEncoderBeginRenderPass(encoder: ptr, descriptor: ptr) ptr
extern fn wgpuRenderPassEncoderSetPipeline(pass: ptr, pipeline: ptr)
extern fn wgpuRenderPassEncoderSetBindGroup(pass: ptr, groupIndex: u32, group: ptr, dynamicOffsetCount: u64, dynamicOffsets: ptr)
extern fn wgpuRenderPassEncoderSetVertexBuffer(pass: ptr, slot: u32, buffer: ptr, offset: u64, size: u64)
extern fn wgpuRenderPassEncoderDraw(pass: ptr, vertexCount: u32, instanceCount: u32, firstVertex: u32, firstInstance: u32)
extern fn wgpuRenderPassEncoderEnd(pass: ptr)
extern fn wgpuCommandEncoderFinish(encoder: ptr, descriptor: ptr) ptr
extern fn wgpuQueueSubmit(queue: ptr, commandCount: u64, commands: ptr)
extern fn wgpuSurfacePresent(surface: ptr)

extern fn wgpuTextureViewRelease(view: ptr)
extern fn wgpuTextureRelease(texture: ptr)
extern fn wgpuRenderPassEncoderRelease(pass: ptr)
extern fn wgpuCommandBufferRelease(cmd: ptr)
extern fn wgpuCommandEncoderRelease(encoder: ptr)
extern fn wgpuRenderPipelineRelease(pipeline: ptr)
extern fn wgpuShaderModuleRelease(module: ptr)
extern fn wgpuQueueRelease(queue: ptr)
extern fn wgpuDeviceRelease(device: ptr)
extern fn wgpuAdapterRelease(adapter: ptr)
extern fn wgpuSurfaceRelease(surface: ptr)
extern fn wgpuInstanceRelease(instance: ptr)


