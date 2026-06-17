#include <windows.h>
#define GLFW_EXPOSE_NATIVE_WIN32
#include <GLFW/glfw3.h>
#include <GLFW/glfw3native.h>
#include <webgpu/webgpu.h>
#include <stdio.h>
#include <stdbool.h>
#include <string.h>

#define SV(str) (WGPUStringView){ .data = (str), .length = strlen(str) }

void* nr_get_hinstance() {
    return GetModuleHandleA(NULL);
}

void* nr_get_hwnd(GLFWwindow* window) {
    return glfwGetWin32Window(window);
}

void* nr_create_surface(void* instance, GLFWwindow* window) {
    HWND hwnd = glfwGetWin32Window(window);
    printf("nr_create_surface: window=%p, hwnd=%p\n", window, hwnd);
    fflush(stdout);

    WGPUSurfaceSourceWindowsHWND hwnd_desc = {
        .chain = {
            .next = NULL,
            .sType = WGPUSType_SurfaceSourceWindowsHWND
        },
        .hinstance = GetModuleHandleA(NULL),
        .hwnd = hwnd
    };

    WGPUSurfaceDescriptor surf_desc = {
        .nextInChain = (const WGPUChainedStruct*)&hwnd_desc,
        .label = {NULL, 0}
    };

    WGPUSurface surface = wgpuInstanceCreateSurface((WGPUInstance)instance, &surf_desc);
    printf("nr_create_surface: surface=%p\n", surface);
    fflush(stdout);
    return surface;
}

void nr_print_surface(void* surface) {
    printf("nr_print_surface: %p\n", surface);
    fflush(stdout);
}

void* nr_erase_lifetime(void* p) {
    return p;
}

void* nr_duplicate_ptr(void* p) {
    return p;
}

void* nr_create_buffer(void* device, uint64_t size, uint32_t usage, uint32_t mappedAtCreation) {
    WGPUBufferDescriptor desc = {
        .nextInChain = NULL,
        .label = SV("nr_buffer"),
        .usage = usage,
        .size = size,
        .mappedAtCreation = mappedAtCreation
    };
    return wgpuDeviceCreateBuffer((WGPUDevice)device, &desc);
}

void nr_queue_write_buffer(void* queue, void* buffer, uint64_t bufferOffset, void* data, uint64_t size) {
    wgpuQueueWriteBuffer((WGPUQueue)queue, (WGPUBuffer)buffer, bufferOffset, data, size);
}

void* nr_create_bind_group_layout_1_uniform(void* device, uint32_t visibility) {
    WGPUBindGroupLayoutEntry bgl_entry = {0};
    bgl_entry.binding = 0;
    bgl_entry.visibility = visibility;
    bgl_entry.buffer.type = WGPUBufferBindingType_Uniform;
    bgl_entry.buffer.hasDynamicOffset = false;
    bgl_entry.buffer.minBindingSize = 0;

    WGPUBindGroupLayoutDescriptor bgl_desc = {
        .nextInChain = NULL,
        .label = SV("bgl_1_uniform"),
        .entryCount = 1,
        .entries = &bgl_entry
    };

    return wgpuDeviceCreateBindGroupLayout((WGPUDevice)device, &bgl_desc);
}

void* nr_create_bind_group_1_uniform(void* device, void* layout, void* buffer, uint64_t size) {
    WGPUBindGroupEntry bg_entry = {
        .nextInChain = NULL,
        .binding = 0,
        .buffer = (WGPUBuffer)buffer,
        .offset = 0,
        .size = size,
        .sampler = NULL,
        .textureView = NULL
    };

    WGPUBindGroupDescriptor bg_desc = {
        .nextInChain = NULL,
        .label = SV("bg_1_uniform"),
        .layout = (WGPUBindGroupLayout)layout,
        .entryCount = 1,
        .entries = &bg_entry
    };

    return wgpuDeviceCreateBindGroup((WGPUDevice)device, &bg_desc);
}

void* nr_create_depth_texture_view(void* device, uint32_t width, uint32_t height) {
    WGPUTextureDescriptor depth_desc = {
        .nextInChain = NULL,
        .label = SV("DepthTexture"),
        .usage = WGPUTextureUsage_RenderAttachment,
        .dimension = WGPUTextureDimension_2D,
        .size = {width, height, 1},
        .format = WGPUTextureFormat_Depth24Plus,
        .mipLevelCount = 1,
        .sampleCount = 1,
        .viewFormatCount = 1,
        .viewFormats = (WGPUTextureFormat[]){WGPUTextureFormat_Depth24Plus}
    };
    WGPUTexture depth_texture = wgpuDeviceCreateTexture((WGPUDevice)device, &depth_desc);

    WGPUTextureViewDescriptor view_desc = {
        .nextInChain = NULL,
        .label = SV("DepthTextureView"),
        .format = WGPUTextureFormat_Depth24Plus,
        .dimension = WGPUTextureViewDimension_2D,
        .baseMipLevel = 0,
        .mipLevelCount = 1,
        .baseArrayLayer = 0,
        .arrayLayerCount = 1,
        .aspect = WGPUTextureAspect_DepthOnly
    };
    return wgpuTextureCreateView(depth_texture, &view_desc);
}

void* nr_create_pipeline_layout_2_bind_groups(void* device, void* layout0, void* layout1) {
    WGPUBindGroupLayout layouts[2] = {(WGPUBindGroupLayout)layout0, (WGPUBindGroupLayout)layout1};
    WGPUPipelineLayoutDescriptor pl_desc = {
        .nextInChain = NULL,
        .label = SV("pipeline_layout"),
        .bindGroupLayoutCount = 2,
        .bindGroupLayouts = layouts
    };
    return wgpuDeviceCreatePipelineLayout((WGPUDevice)device, &pl_desc);
}

void* nr_create_render_pipeline(void* device, void* layout, void* shader_module) {
    WGPUVertexAttribute vert_attrs[3];
    vert_attrs[0] = (WGPUVertexAttribute){
        .format = WGPUVertexFormat_Float32x3,
        .offset = 0,
        .shaderLocation = 0
    };
    vert_attrs[1] = (WGPUVertexAttribute){
        .format = WGPUVertexFormat_Float32x3,
        .offset = 12,
        .shaderLocation = 1
    };
    vert_attrs[2] = (WGPUVertexAttribute){
        .format = WGPUVertexFormat_Float32x3,
        .offset = 24,
        .shaderLocation = 2
    };

    WGPUVertexBufferLayout vb_layout = {
        .arrayStride = 36, // 3 floats for pos, 3 floats for color, 3 floats for normal
        .stepMode = WGPUVertexStepMode_Vertex,
        .attributeCount = 3,
        .attributes = vert_attrs
    };

    WGPUBlendState blend_state = {
        .color = {
            .operation = WGPUBlendOperation_Add,
            .srcFactor = WGPUBlendFactor_One,
            .dstFactor = WGPUBlendFactor_Zero
        },
        .alpha = {
            .operation = WGPUBlendOperation_Add,
            .srcFactor = WGPUBlendFactor_One,
            .dstFactor = WGPUBlendFactor_Zero
        }
    };

    WGPUColorTargetState color_target = {
        .nextInChain = NULL,
        .format = WGPUTextureFormat_BGRA8Unorm, // Window surface format
        .blend = &blend_state,
        .writeMask = WGPUColorWriteMask_All
    };

    WGPUFragmentState fragment_state = {
        .nextInChain = NULL,
        .module = (WGPUShaderModule)shader_module,
        .entryPoint = SV("fs_main"),
        .constantCount = 0,
        .constants = NULL,
        .targetCount = 1,
        .targets = &color_target
    };

    WGPUDepthStencilState depth_stencil = {
        .nextInChain = NULL,
        .format = WGPUTextureFormat_Depth24Plus,
        .depthWriteEnabled = true,
        .depthCompare = WGPUCompareFunction_Less,
        .stencilFront = { .compare = WGPUCompareFunction_Always, .failOp = WGPUStencilOperation_Keep, .depthFailOp = WGPUStencilOperation_Keep, .passOp = WGPUStencilOperation_Keep },
        .stencilBack = { .compare = WGPUCompareFunction_Always, .failOp = WGPUStencilOperation_Keep, .depthFailOp = WGPUStencilOperation_Keep, .passOp = WGPUStencilOperation_Keep },
        .stencilReadMask = 0xFFFFFFFF,
        .stencilWriteMask = 0xFFFFFFFF,
        .depthBias = 0,
        .depthBiasSlopeScale = 0.0f,
        .depthBiasClamp = 0.0f
    };

    WGPURenderPipelineDescriptor pipeline_desc = {
        .nextInChain = NULL,
        .label = SV("RenderPipeline"),
        .layout = (WGPUPipelineLayout)layout,
        .vertex = {
            .nextInChain = NULL,
            .module = (WGPUShaderModule)shader_module,
            .entryPoint = SV("vs_main"),
            .constantCount = 0,
            .constants = NULL,
            .bufferCount = 1,
            .buffers = &vb_layout
        },
        .primitive = {
            .nextInChain = NULL,
            .topology = WGPUPrimitiveTopology_TriangleList,
            .stripIndexFormat = WGPUIndexFormat_Undefined,
            .frontFace = WGPUFrontFace_CCW,
            .cullMode = WGPUCullMode_Back
        },
        .depthStencil = &depth_stencil,
        .multisample = {
            .nextInChain = NULL,
            .count = 1,
            .mask = 0xFFFFFFFF,
            .alphaToCoverageEnabled = false
        },
        .fragment = &fragment_state
    };

    return wgpuDeviceCreateRenderPipeline((WGPUDevice)device, &pipeline_desc);
}

void nr_print_string_view(const char* data, uint64_t length) {
    if (!data) {
        printf("WGPU Error: null message\n");
        return;
    }
    printf("WGPU Error: %.*s\n", (int)length, data);
    fflush(stdout);
}
