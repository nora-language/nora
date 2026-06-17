package target

import (
	"fmt"
	"runtime"
	"strings"
)

type Platform struct {
	Name      string
	OS        string
	Arch      string
	Compiler  string
	CFlags    []string
	ExeSuffix string
	Wasm      bool
}

var (
	WindowsAmd64 = Platform{
		Name:      "windows-amd64",
		OS:        "windows",
		Arch:      "amd64",
		Compiler:  "clang",
		CFlags:    []string{"-O3", "-Wno-pointer-sign", "-Wno-deprecated-declarations", "-Wno-parentheses-equality", "-Wno-unused-value", "-ldbghelp", "-lws2_32"},
		ExeSuffix: ".exe",
	}
	LinuxAmd64 = Platform{
		Name:      "linux-amd64",
		OS:        "linux",
		Arch:      "amd64",
		Compiler:  "clang",
		CFlags:    []string{"-target", "x86_64-linux-gnu", "-O3", "-Wno-pointer-sign", "-Wno-deprecated-declarations", "-Wno-parentheses-equality", "-Wno-unused-value", "-pthread", "-lm", "-ldl"},
		ExeSuffix: "",
	}
	WasmUnknown = Platform{
		Name:      "wasm-unknown",
		OS:        "unknown",
		Arch:      "wasm",
		Compiler:  "emcc",
		CFlags:    []string{"-O3", "-s", "STANDALONE_WASM", "-s", "ASYNCIFY", "-Wno-pointer-sign", "-Wno-deprecated-declarations", "-Wno-parentheses-equality", "-Wno-unused-value"},
		ExeSuffix: ".wasm",
		Wasm:      true,
	}
	WasmWasi = Platform{
		Name:      "wasm-wasi",
		OS:        "wasi",
		Arch:      "wasm32",
		Compiler:  "clang",
		CFlags:    []string{"--target=wasm32-wasip1", "-matomics", "-mbulk-memory", "-O3", "-Wno-pointer-sign", "-Wno-deprecated-declarations", "-Wno-parentheses-equality", "-Wno-unused-value"},
		ExeSuffix: ".wasm",
		Wasm:      true,
	}
)

var Registry = map[string]Platform{
	"windows-amd64": WindowsAmd64,
	"linux-amd64":   LinuxAmd64,
	"wasm-unknown":  WasmUnknown,
	"wasm-wasi":     WasmWasi,
	"wasi":          WasmWasi,
	"wasm":          WasmUnknown, // Alias for legacy
}

func Get(name string) (Platform, bool) {
	p, ok := Registry[strings.ToLower(name)]
	return p, ok
}

func Discover() Platform {
	os := runtime.GOOS
	arch := runtime.GOARCH

	name := fmt.Sprintf("%s-%s", os, arch)
	if p, ok := Get(name); ok {
		return p
	}

	// Fallback/Generic
	return Platform{
		Name:     name,
		OS:       os,
		Arch:     arch,
		Compiler: "clang",
		CFlags:   []string{"-O3"},
	}
}

func List() []Platform {
	return []Platform{WindowsAmd64, LinuxAmd64, WasmWasi, WasmUnknown}
}
