#include <windows.h>
#define GLFW_EXPOSE_NATIVE_WIN32
#include <GLFW/glfw3.h>
#include <GLFW/glfw3native.h>
#include <webgpu/webgpu.h>
#include <stdio.h>

void* nr_get_hinstance() {
    return GetModuleHandleA(NULL);
}

void* nr_get_hwnd(GLFWwindow* window) {
    return glfwGetWin32Window(window);
}

void* nr_create_surface(void* instance, GLFWwindow* window) {
    WGPUSurfaceSourceWindowsHWND hwnd_desc = {
        .chain = {
            .next = NULL,
            .sType = WGPUSType_SurfaceSourceWindowsHWND
        },
        .hinstance = GetModuleHandleA(NULL),
        .hwnd = glfwGetWin32Window(window)
    };

    WGPUSurfaceDescriptor surf_desc = {
        .nextInChain = (const WGPUChainedStruct*)&hwnd_desc,
        .label = {NULL, 0}
    };

    WGPUSurface surface = wgpuInstanceCreateSurface((WGPUInstance)instance, &surf_desc);
    return surface;
}

void* nr_create_buffer(void* device, uint64_t size, uint32_t usage, uint32_t mappedAtCreation) {
    WGPUBufferDescriptor desc = {0};
    desc.size = size;
    desc.usage = usage;
    desc.mappedAtCreation = mappedAtCreation;
    return wgpuDeviceCreateBuffer((WGPUDevice)device, &desc);
}

void nr_queue_write_buffer(void* queue, void* buffer, uint64_t bufferOffset, void* data, uint64_t size) {
    wgpuQueueWriteBuffer((WGPUQueue)queue, (WGPUBuffer)buffer, bufferOffset, data, size);
}

void* nr_create_bind_group_layout_1_uniform(void* device, uint32_t visibility) {
    WGPUBindGroupLayoutEntry entry = {0};
    entry.binding = 0;
    entry.visibility = visibility;
    entry.buffer.type = WGPUBufferBindingType_Uniform;
    entry.buffer.minBindingSize = 0;

    WGPUBindGroupLayoutDescriptor desc = {0};
    desc.entryCount = 1;
    desc.entries = &entry;
    
    return wgpuDeviceCreateBindGroupLayout((WGPUDevice)device, &desc);
}

void* nr_create_bind_group_1_uniform(void* device, void* layout, void* buffer, uint64_t size) {
    WGPUBindGroupEntry entry = {0};
    entry.binding = 0;
    entry.buffer = (WGPUBuffer)buffer;
    entry.offset = 0;
    entry.size = size;

    WGPUBindGroupDescriptor desc = {0};
    desc.layout = (WGPUBindGroupLayout)layout;
    desc.entryCount = 1;
    desc.entries = &entry;
    
    return wgpuDeviceCreateBindGroup((WGPUDevice)device, &desc);
}

void* nr_create_depth_texture_view(void* device, uint32_t width, uint32_t height) {
    WGPUTextureDescriptor desc = {0};
    desc.usage = WGPUTextureUsage_RenderAttachment;
    desc.dimension = WGPUTextureDimension_2D;
    desc.size.width = width;
    desc.size.height = height;
    desc.size.depthOrArrayLayers = 1;
    desc.format = WGPUTextureFormat_Depth24Plus;
    desc.mipLevelCount = 1;
    desc.sampleCount = 1;
    WGPUTexture texture = wgpuDeviceCreateTexture((WGPUDevice)device, &desc);
    
    WGPUTextureViewDescriptor viewDesc = {0};
    viewDesc.format = WGPUTextureFormat_Depth24Plus;
    viewDesc.dimension = WGPUTextureViewDimension_2D;
    viewDesc.baseMipLevel = 0;
    viewDesc.mipLevelCount = 1;
    viewDesc.baseArrayLayer = 0;
    viewDesc.arrayLayerCount = 1;
    viewDesc.aspect = WGPUTextureAspect_DepthOnly;
    return wgpuTextureCreateView(texture, &viewDesc);
}

void* nr_create_pipeline_layout_2_bind_groups(void* device, void* layout0, void* layout1) {
    WGPUBindGroupLayout layouts[2] = { (WGPUBindGroupLayout)layout0, (WGPUBindGroupLayout)layout1 };
    WGPUPipelineLayoutDescriptor desc = {0};
    desc.bindGroupLayoutCount = 2;
    desc.bindGroupLayouts = layouts;
    return wgpuDeviceCreatePipelineLayout((WGPUDevice)device, &desc);
}

void* nr_create_render_pipeline(void* device, void* layout, void* shader_module) {
    // Vertex Attributes for Vertex: pos(f32[3]), col(f32[3]), norm(f32[3])
    WGPUVertexAttribute attrs[3] = {0};
    attrs[0].format = WGPUVertexFormat_Float32x3;
    attrs[0].offset = 0;
    attrs[0].shaderLocation = 0;
    
    attrs[1].format = WGPUVertexFormat_Float32x3;
    attrs[1].offset = 12;
    attrs[1].shaderLocation = 1;
    
    attrs[2].format = WGPUVertexFormat_Float32x3;
    attrs[2].offset = 24;
    attrs[2].shaderLocation = 2;

    WGPUVertexBufferLayout buf_layout = {0};
    buf_layout.arrayStride = 36;
    buf_layout.stepMode = WGPUVertexStepMode_Vertex;
    buf_layout.attributeCount = 3;
    buf_layout.attributes = attrs;

    WGPUBlendState blend = {0};
    blend.color.operation = WGPUBlendOperation_Add;
    blend.color.srcFactor = WGPUBlendFactor_One;
    blend.color.dstFactor = WGPUBlendFactor_Zero;
    blend.alpha.operation = WGPUBlendOperation_Add;
    blend.alpha.srcFactor = WGPUBlendFactor_One;
    blend.alpha.dstFactor = WGPUBlendFactor_Zero;

    WGPUColorTargetState color_target = {0};
    color_target.format = WGPUTextureFormat_BGRA8Unorm;
    color_target.blend = &blend;
    color_target.writeMask = WGPUColorWriteMask_All;

    WGPUFragmentState fragment = {0};
    fragment.module = (WGPUShaderModule)shader_module;
    fragment.entryPoint = (WGPUStringView){"fs_main", 7};
    fragment.targetCount = 1;
    fragment.targets = &color_target;

    WGPUDepthStencilState depth = {0};
    depth.format = WGPUTextureFormat_Depth24Plus;
    depth.depthWriteEnabled = 1; // true
    depth.depthCompare = WGPUCompareFunction_Less;

    WGPURenderPipelineDescriptor desc = {0};
    desc.layout = (WGPUPipelineLayout)layout;
    desc.vertex.module = (WGPUShaderModule)shader_module;
    desc.vertex.entryPoint = (WGPUStringView){"vs_main", 7};
    desc.vertex.bufferCount = 1;
    desc.vertex.buffers = &buf_layout;
    
    desc.primitive.topology = WGPUPrimitiveTopology_TriangleList;
    desc.primitive.cullMode = WGPUCullMode_Back;
    desc.primitive.frontFace = WGPUFrontFace_CCW;
    
    desc.multisample.count = 1;
    desc.multisample.mask = 0xFFFFFFFF;
    
    desc.fragment = &fragment;
    desc.depthStencil = &depth;

    return wgpuDeviceCreateRenderPipeline((WGPUDevice)device, &desc);
}