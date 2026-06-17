package codegen

import "runtime"

// Target defines the platform-specific behavior for C code emission
type Target interface {
	// Name returns the name of the target platform
	Name() string
	// Bootstrap returns the C code required to initialize the platform runtime
	Bootstrap() string
	// GetMainWrapper returns the C code for the entry point (main function)
	GetMainWrapper(mainFunc string) string
}

// GetTarget returns the Target implementation based on the generator's config
func (g *Generator) GetTarget() Target {
	target := g.Target
	if target == "" {
		target = runtime.GOOS
	}

	switch target {
	case "wasm":
		return &WasmTarget{}
	case "windows":
		return &WindowsTarget{}
	case "linux", "linux-amd64":
		return &LinuxTarget{}
	default:
		return &WindowsTarget{} // Fallback
	}
}
