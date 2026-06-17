package rendering

[include("<GLFW/glfw3.h>")]
pub type GLFWwindow = struct {}

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

pub fn make_window(w: i32, h: i32, title: ptr) Window {
    glfwInit()
    glfwWindowHint(0x00022001, 0) // GLFW_CLIENT_API = GLFW_NO_API
    var handle = glfwCreateWindow(w, h, title, ptr(0), ptr(0))
    return Window {
        handle: handle
    }
}

pub fn (self: #Window) should_close() bool {
    return glfwWindowShouldClose(self.handle) != 0
}

pub fn (self: #Window) swap_buffers() {
    glfwSwapBuffers(self.handle)
}

pub fn (_self: #Window) poll_events() {
    glfwPollEvents()
}

pub fn (self: #Window) drop() {
    glfwDestroyWindow(self.handle)
    glfwTerminate()
}
