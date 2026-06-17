package plugin

import (
	"context"
	"fmt"
	"os"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

// NativeMacro is a Go function that implements a macro directly.
type NativeMacro func(requestJSON []byte) ([]byte, error)

// PluginManager handles loading and executing WebAssembly macros.
type PluginManager struct {
	runtime      wazero.Runtime
	modules      map[string]api.Module
	nativeMacros map[string]map[string]NativeMacro
}

// NewPluginManager creates a new sandboxed WebAssembly runtime.
func NewPluginManager() *PluginManager {
	ctx := context.Background()
	r := wazero.NewRuntime(ctx)

	// Instantiate WASI, which is required for many WASM compilers (like Go/Rust)
	// even if the plugin doesn't perform I/O.
	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	return &PluginManager{
		runtime:      r,
		modules:      make(map[string]api.Module),
		nativeMacros: make(map[string]map[string]NativeMacro),
	}
}

// RegisterNativeMacro registers a Go function as a macro handler.
func (m *PluginManager) RegisterNativeMacro(pluginName string, macroName string, handler NativeMacro) {
	if _, ok := m.nativeMacros[pluginName]; !ok {
		m.nativeMacros[pluginName] = make(map[string]NativeMacro)
	}
	m.nativeMacros[pluginName][macroName] = handler
}

// LoadPlugin reads a compiled .wasm file and initializes it in the sandbox.
func (m *PluginManager) LoadPlugin(name string, path string) error {
	ctx := context.Background()

	wasmBytes, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read plugin %s: %w", path, err)
	}

	compiled, err := m.runtime.CompileModule(ctx, wasmBytes)
	if err != nil {
		return fmt.Errorf("failed to compile plugin %s: %w", path, err)
	}

	module, err := m.runtime.InstantiateModule(ctx, compiled, wazero.NewModuleConfig().WithName(name))
	if err != nil {
		return fmt.Errorf("failed to instantiate plugin %s: %w", path, err)
	}

	// Initialize the WASI reactor module if it exports `_initialize`
	initFn := module.ExportedFunction("_initialize")
	if initFn != nil {
		if _, err := initFn.Call(ctx); err != nil {
			return fmt.Errorf("failed to initialize plugin %s: %w", path, err)
		}
	}

	m.modules[name] = module
	return nil
}

// Close cleans up the runtime.
func (m *PluginManager) Close() {
	m.runtime.Close(context.Background())
}

// ExecuteMacro writes a JSON payload into the WASM sandbox memory, executes the
// macro function, and returns the modified JSON payload.
func (m *PluginManager) ExecuteMacro(pluginName string, macroName string, requestJSON []byte) ([]byte, error) {
	// 1. Check for WASM modules first
	module, ok := m.modules[pluginName]
	if ok {
		ctx := context.Background()

		// 1. Get the allocation function
		allocFn := module.ExportedFunction("plugin_alloc")
		if allocFn == nil {
			return nil, fmt.Errorf("plugin %s does not export plugin_alloc", pluginName)
		}

		// 2. Allocate memory for the JSON string
		allocSize := uint64(len(requestJSON) + 1) // +1 for null terminator
		results, err := allocFn.Call(ctx, allocSize)
		if err != nil || len(results) == 0 {
			return nil, fmt.Errorf("failed to allocate memory in plugin: %w", err)
		}
		ptr := uint32(results[0])

		// 3. Write JSON to memory
		if !module.Memory().Write(ptr, requestJSON) {
			return nil, fmt.Errorf("failed to write JSON to plugin memory")
		}
		module.Memory().WriteByte(ptr+uint32(len(requestJSON)), 0) // Null terminate

		// 4. Call the macro function
		macroFn := module.ExportedFunction(macroName)
		if macroFn == nil {
			return nil, fmt.Errorf("plugin %s does not export macro %s", pluginName, macroName)
		}

		resResults, err := macroFn.Call(ctx, uint64(ptr))
		if err != nil || len(resResults) == 0 {
			return nil, fmt.Errorf("failed to execute macro %s: %w", macroName, err)
		}
		resPtr := uint32(resResults[0])

		// 5. Read the response JSON from memory (null-terminated string)
		resBytes, ok := module.Memory().Read(resPtr, uint32(module.Memory().Size())-resPtr)
		if !ok {
			return nil, fmt.Errorf("failed to read response from plugin memory")
		}

		// Find null terminator
		end := 0
		for i, b := range resBytes {
			if b == 0 {
				end = i
				break
			}
		}

		// 6. Reset heap for next call
		resetFn := module.ExportedFunction("plugin_reset")
		if resetFn != nil {
			resetFn.Call(ctx)
		}

		return resBytes[:end], nil
	}

	// 2. Fallback to Native (Go) Macros
	if p, ok := m.nativeMacros[pluginName]; ok {
		if macro, ok := p[macroName]; ok {
			return macro(requestJSON)
		}
	}
	return nil, fmt.Errorf("plugin %s not loaded and no native fallback found", pluginName)
}
