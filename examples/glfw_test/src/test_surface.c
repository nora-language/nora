#include <windows.h>
#define GLFW_EXPOSE_NATIVE_WIN32
#include <GLFW/glfw3.h>
#include <GLFW/glfw3native.h>
#include <webgpu/webgpu.h>
#include <stdio.h>

int main() {
    if (!glfwInit()) return 1;
    glfwWindowHint(GLFW_CLIENT_API, GLFW_NO_API);
    GLFWwindow* window = glfwCreateWindow(800, 600, "Test", NULL, NULL);

    WGPUInstanceDescriptor inst_desc = {0};
    WGPUInstance instance = wgpuCreateInstance(&inst_desc);
    printf("Instance: %p\n", instance);

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

    WGPUSurface surface = wgpuInstanceCreateSurface(instance, &surf_desc);
    printf("Surface: %p\n", surface);

    return 0;
}
