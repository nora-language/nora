package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"context"

	"github.com/nora-language/nora/pkg/codegen"
	"github.com/nora-language/nora/pkg/diag"
	"github.com/nora-language/nora/pkg/docgen"
	"github.com/nora-language/nora/pkg/format"
	"github.com/nora-language/nora/pkg/lexer"
	"github.com/nora-language/nora/pkg/lsp"
	"github.com/nora-language/nora/pkg/parser"
	"github.com/nora-language/nora/pkg/parser/ast"
	"github.com/nora-language/nora/pkg/plugin"
	"github.com/nora-language/nora/pkg/semantic"
	"github.com/nora-language/nora/pkg/target"
	"github.com/nora-language/nora/pkg/topology"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"gopkg.in/yaml.v3"
)

var StdPath string
var CorePath string

const CompilerVersion = "0.1.0"
const LanguageVersion = "0.1.0"

func initCorePath() {
	if StdPath != "" {
		path := filepath.Join(filepath.Dir(StdPath), "core")
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			CorePath = path
			return
		}
	}

	// 1. Check Env
	if env := os.Getenv("NORA_CORE_PATH"); env != "" {
		CorePath = env
		return
	}

	// 2. Check project-level fallback
	projectFallbacks := []string{"lib/core", "libs/core"}
	for _, fb := range projectFallbacks {
		if info, err := os.Stat(fb); err == nil && info.IsDir() {
			CorePath = fb
			return
		}
	}

	// 3. Check relative to executable
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)

		// Try ./core relative to exe
		path := filepath.Join(dir, "core")
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			CorePath = path
			return
		}

		// Try ../core relative to exe
		path = filepath.Join(dir, "..", "core")
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			CorePath = path
			return
		}
	}

	// 4. Check system locations (Unix-like)
	systemPaths := []string{
		"/usr/local/lib/Nora/core",
		"/lib/Nora/core",
		"/usr/lib/Nora/core",
	}

	for _, p := range systemPaths {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			CorePath = p
			return
		}
	}

	// 5. Fallback to current working directory
	CorePath = "core"
}

func initStdPath() {
	// 1. Check Env
	if env := os.Getenv("NORA_STD_PATH"); env != "" {
		StdPath = env
		return
	}

	// 2. Check project-level fallback
	projectFallbacks := []string{"lib/std", "libs/std"}
	for _, fb := range projectFallbacks {
		if info, err := os.Stat(fb); err == nil && info.IsDir() {
			StdPath = fb
			return
		}
	}

	// 3. Check relative to executable
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)

		// Try ./std relative to exe
		path := filepath.Join(dir, "std")
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			StdPath = path
			return
		}

		// Try ../std relative to exe
		path = filepath.Join(dir, "..", "std")
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			StdPath = path
			return
		}
	}

	// 4. Check system locations (Unix-like)
	systemPaths := []string{
		"/usr/local/lib/Nora/std",
		"/lib/Nora/std",
		"/usr/lib/Nora/std",
	}
	if runtime.GOOS != "windows" {
		for _, p := range systemPaths {
			if info, err := os.Stat(p); err == nil && info.IsDir() {
				StdPath = p
				return
			}
		}
	}

	// 5. Fallback to current working directory
	StdPath = "std"
}

type Dependency struct {
	Path    string `yaml:"path"`
	Version string `yaml:"version"`
}

type CompilerConfig struct {
	Compiler     string   `yaml:"compiler"`      // E.g. "clang", "gcc", "cl", "tcc"
	OptRelease   string   `yaml:"opt_release"`   // E.g. "-O3" or "/O2"
	OptDebug     string   `yaml:"opt_debug"`     // E.g. "-O0" or "/Od"
	DebugSymbols string   `yaml:"debug_symbols"` // E.g. "-g" or "/Zi"
	OutFlag      string   `yaml:"out_flag"`      // E.g. "-o" or "/Fe:"
	IncFlag      string   `yaml:"inc_flag"`      // E.g. "-I" or "/I"
	DefineFlag   string   `yaml:"define_flag"`   // E.g. "-D" or "/D"
	LibDirFlag   string   `yaml:"lib_dir_flag"`  // E.g. "-L" or "/link /LIBPATH:"
	CFlags       []string `yaml:"cflags"`        // Custom/extra flags
}

type NativeConfig struct {
	CompilerConfig `yaml:",inline"`

	DynamicLibs []string `yaml:"dynamic_libs"`
	StaticLibs  []string `yaml:"static_libs"`
	IncludeDirs []string `yaml:"include_dirs"`
	LibDirs     []string `yaml:"lib_dirs"`
	Headers     []string `yaml:"headers"`
	SourceFiles []string `yaml:"source_files"`
}

type ProjectConfig struct {
	Name         string                `yaml:"name"`
	Version      string                `yaml:"version"`
	Language     string                `yaml:"language"`
	Entry        string                `yaml:"entry"`
	Output       string                `yaml:"output"`
	Plugins      []string              `yaml:"plugins"`
	Dependencies map[string]Dependency `yaml:"dependencies"`
	Native       NativeConfig          `yaml:"native"`
	NoStdlib     bool                  `yaml:"no_stdlib"`
	NoCore       bool                  `yaml:"no_core"`
	AllowUnsafe  bool                  `yaml:"allow_unsafe"`
}

type BuildOptions struct {
	Target           target.Platform
	Release          bool
	Debug            bool
	DebugMemory      bool
	DebugSemantic    bool
	DebugTopology    bool
	DebugFiber       bool
	G                bool
	WasmExperimental bool
	AllowUnsafe      bool // Allow [unsafe] attribute
	NoStdlib         bool // Skip stdlib injection
	NoCore           bool // Skip core injection
	Native           NativeConfig
	Compiler         string   // CLI override
	CFlags           []string // CLI override
	Verbose          bool     // Show verbose logs
}

func LoadProjectConfig() (*ProjectConfig, error) {
	data, err := os.ReadFile("nora.yaml")
	if err != nil {
		return nil, err
	}
	var config ProjectConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (c *ProjectConfig) Save() error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile("nora.yaml", data, 0644)
}

type FileLoader struct {
	Cache             map[string]*semantic.Scope
	ParsedFiles       map[string]*ast.File
	Program           *ast.Program
	Analyzer          *semantic.SemanticAnalyzer
	Dependencies      map[string]Dependency
	CollectedNative   NativeConfig
	CollectedPlugins  []string
	PluginManager     *plugin.PluginManager
	NoStdlib          bool
	NoCore            bool
	AllowedUnsafeDirs []string
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func (n *NativeConfig) Merge(other NativeConfig, baseDir string) {
	resolvePath := func(p string) string {
		if filepath.IsAbs(p) || strings.HasPrefix(p, "http") {
			return p
		}
		return filepath.Clean(filepath.Join(baseDir, p))
	}

	if n.Compiler == "" {
		n.Compiler = other.Compiler
	}
	if n.OptRelease == "" {
		n.OptRelease = other.OptRelease
	}
	if n.OptDebug == "" {
		n.OptDebug = other.OptDebug
	}
	if n.DebugSymbols == "" {
		n.DebugSymbols = other.DebugSymbols
	}
	if n.OutFlag == "" {
		n.OutFlag = other.OutFlag
	}
	if n.IncFlag == "" {
		n.IncFlag = other.IncFlag
	}
	if n.DefineFlag == "" {
		n.DefineFlag = other.DefineFlag
	}
	if n.LibDirFlag == "" {
		n.LibDirFlag = other.LibDirFlag
	}
	for _, cf := range other.CFlags {
		if !contains(n.CFlags, cf) {
			n.CFlags = append(n.CFlags, cf)
		}
	}

	for _, d := range other.DynamicLibs {
		if !contains(n.DynamicLibs, d) {
			n.DynamicLibs = append(n.DynamicLibs, d)
		}
	}
	for _, s := range other.StaticLibs {
		resolved := resolvePath(s)
		if !contains(n.StaticLibs, resolved) {
			n.StaticLibs = append(n.StaticLibs, resolved)
		}
	}
	for _, inc := range other.IncludeDirs {
		resolved := resolvePath(inc)
		if !contains(n.IncludeDirs, resolved) {
			n.IncludeDirs = append(n.IncludeDirs, resolved)
		}
	}
	for _, lib := range other.LibDirs {
		resolved := resolvePath(lib)
		if !contains(n.LibDirs, resolved) {
			n.LibDirs = append(n.LibDirs, resolved)
		}
	}
	for _, hdr := range other.Headers {
		if !contains(n.Headers, hdr) {
			n.Headers = append(n.Headers, hdr)
		}
	}
	for _, src := range other.SourceFiles {
		resolved := resolvePath(src)
		if !contains(n.SourceFiles, resolved) {
			n.SourceFiles = append(n.SourceFiles, resolved)
		}
	}
}

func (f *FileLoader) loadManifest(dirPath string) {
	manifestPath := filepath.Join(dirPath, "nora.yaml")
	if data, err := os.ReadFile(manifestPath); err == nil {
		var config ProjectConfig
		if err := yaml.Unmarshal(data, &config); err == nil {
			f.NoStdlib = config.NoStdlib
			f.NoCore = config.NoCore
			if config.AllowUnsafe {
				cleanDir := filepath.Clean(dirPath)
				if runtime.GOOS == "windows" {
					cleanDir = strings.ToLower(cleanDir)
				}
				if !contains(f.AllowedUnsafeDirs, cleanDir) {
					f.AllowedUnsafeDirs = append(f.AllowedUnsafeDirs, cleanDir)
				}
			}
			if f.Dependencies == nil {
				f.Dependencies = make(map[string]Dependency)
			}
			for name, dep := range config.Dependencies {
				if _, exists := f.Dependencies[name]; !exists {
					if !filepath.IsAbs(dep.Path) && !strings.HasPrefix(dep.Path, "http") {
						dep.Path = filepath.Join(dirPath, dep.Path)
					}
					f.Dependencies[name] = dep
				}
			}
			f.CollectedNative.Merge(config.Native, dirPath)
			for _, p := range config.Plugins {
				resolved := p
				if !filepath.IsAbs(p) {
					resolved = filepath.Join(dirPath, p)
				}
				if !contains(f.CollectedPlugins, resolved) {
					f.CollectedPlugins = append(f.CollectedPlugins, resolved)
					if f.PluginManager != nil {
						name := filepath.Base(resolved)
						name = strings.TrimSuffix(name, filepath.Ext(name))
						if err := f.PluginManager.LoadPlugin(name, resolved); err != nil {
							fmt.Printf("Warning: failed to load dynamic plugin %s: %v\n", resolved, err)
						}
					}
				}
			}
		}
	}
}

func (f *FileLoader) Load(path string) (*semantic.Scope, error) {
	path = filepath.Clean(path)
	// 1. Check Dependencies from nora.yaml
	if dep, ok := f.Dependencies[path]; ok {
		actualPath := dep.Path
		// Verify version and load transitive dependencies
		libConfigPath := filepath.Join(actualPath, "nora.yaml")
		if data, err := os.ReadFile(libConfigPath); err == nil {
			var libConfig ProjectConfig
			if err := yaml.Unmarshal(data, &libConfig); err == nil {
				// Version check
				if libConfig.Version != dep.Version && dep.Version != "" {
					fmt.Printf("Warning: dependency '%s' version mismatch. Expected %s, found %s at %s\n",
						path, dep.Version, libConfig.Version, actualPath)
				}

				// Transitive dependencies
				if f.Dependencies == nil {
					f.Dependencies = make(map[string]Dependency)
				}
				for transName, transDep := range libConfig.Dependencies {
					if existing, exists := f.Dependencies[transName]; exists {
						if existing.Version != transDep.Version && transDep.Version != "" && existing.Version != "" {
							fmt.Printf("Warning: version conflict for dependency '%s' found in '%s'. Using '%s' (v%s) but found '%s' (v%s) as transitive dependency.\n",
								transName, path, transName, existing.Version, transName, transDep.Version)
						}
					} else {
						// Resolve relative path
						if !filepath.IsAbs(transDep.Path) && !strings.HasPrefix(transDep.Path, "http") {
							transDep.Path = filepath.Join(actualPath, transDep.Path)
						}
						f.Dependencies[transName] = transDep
					}
				}
			}
		}
		path = actualPath
	} else if !filepath.IsAbs(path) &&
		!strings.HasPrefix(path, "./") &&
		!strings.HasPrefix(path, "../") {
		// 2. Check if it exists in core/
		coreCandidate := filepath.Join(CorePath, path)
		stdCandidate := filepath.Join(StdPath, path)

		if _, err := os.Stat(coreCandidate); err == nil {
			path = coreCandidate
		} else if _, err := os.Stat(coreCandidate + ".nr"); err == nil {
			path = coreCandidate + ".nr"
		} else if _, err := os.Stat(stdCandidate); err == nil {
			// 3. Check if it exists in std/
			path = stdCandidate
		} else if _, err := os.Stat(stdCandidate + ".nr"); err == nil {
			path = stdCandidate + ".nr"
		} else {
			// If not in core/ or std/, check if it exists locally, otherwise fallback to std/ just in case
			if _, err := os.Stat(path); err != nil {
				if _, err := os.Stat(path + ".nr"); err != nil {
					path = stdCandidate
				}
			}
		}
	}

	// Load package manifest if present
	if info, err := os.Stat(path); err == nil {
		if info.IsDir() {
			f.loadManifest(path)
		} else {
			f.loadManifest(filepath.Dir(path))
		}
	}

	if !strings.HasSuffix(path, ".nr") {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			// It's a directory, load all .nr files in it
			if scope, exists := f.Cache[path]; exists {
				return scope, nil
			}

			// Pre-prevent recursive loading by marking directory as loaded!
			dirClean := filepath.Clean(path)
			if runtime.GOOS == "windows" {
				dirClean = strings.ToLower(dirClean)
			}
			if f.Analyzer.LoadedDirs == nil {
				f.Analyzer.LoadedDirs = make(map[string]bool)
			}
			f.Analyzer.LoadedDirs[dirClean] = true

			files, err := os.ReadDir(path)
			if err != nil {
				return nil, err
			}

			var pkgScope *semantic.Scope
			for _, fileInfo := range files {
				if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".nr") {
					fullFilePath := filepath.Join(path, fileInfo.Name())
					if file, exists := f.ParsedFiles[fullFilePath]; exists {
						if pkgScope == nil {
							pkgName := f.Analyzer.GetPackageName(file)
							pkgScope = f.Analyzer.GetPackageScope(pkgName)
						}
						continue
					}
					input, err := os.ReadFile(fullFilePath)
					if err != nil {
						continue
					}

					l := lexer.New(string(input), fullFilePath)
					p := parser.New(l)
					p.AllowNoPackage = false
					file := p.Parse(fullFilePath)

					if f.PluginManager != nil {
						if err := f.PluginManager.ProcessMacroForFile(file); err != nil {
							fmt.Printf("Macro Error in %s: %v\n", fullFilePath, err)
						}
					}

					f.Program.Files = append(f.Program.Files, file)
					f.ParsedFiles[fullFilePath] = file

					// First pass: collect symbols
					f.Analyzer.CollectSymbols(file)

					// Get the scope for this file (should be shared for all files in this directory)
					if pkgScope == nil {
						pkgName := f.Analyzer.GetPackageName(file)
						pkgScope = f.Analyzer.GetPackageScope(pkgName)
					}
				}
			}

			if pkgScope == nil {
				return nil, fmt.Errorf("no .nr files found in directory %s", path)
			}

			f.Cache[path] = pkgScope
			return pkgScope, nil
		} else {
			path += ".nr"
		}
	}

	if scope, exists := f.Cache[path]; exists {
		return scope, nil
	}

	if parsed, ok := f.ParsedFiles[path]; ok {
		return f.Analyzer.GetPackageScope(f.Analyzer.GetPackageName(parsed)), nil
	}

	input, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	l := lexer.New(string(input), path)
	p := parser.New(l)
	p.AllowNoPackage = false
	file := p.Parse(path)

	if f.PluginManager != nil {
		if err := f.PluginManager.ProcessMacroForFile(file); err != nil {
			fmt.Printf("Macro Error in %s: %v\n", path, err)
		}
	}

	f.Program.Files = append(f.Program.Files, file)
	f.ParsedFiles[path] = file

	f.Analyzer.CollectSymbols(file)
	pkgName := f.Analyzer.GetPackageName(file)
	scope := f.Analyzer.GetPackageScope(pkgName)
	f.Cache[path] = scope
	return scope, nil
}

func main() {
	initStdPath()
	initCorePath()
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	for _, arg := range os.Args[1:] {
		if arg == "--version" || arg == "-v" {
			fmt.Printf("Nora Compiler v%s\n", CompilerVersion)
			fmt.Printf("Language Version v%s\n", LanguageVersion)
			os.Exit(0)
		}
	}

	command := os.Args[1]

	switch command {
	case "init":
		runInit(os.Args[2:])
	case "build":
		runBuild(os.Args[2:])
	case "run":
		runRun(os.Args[2:])
	case "check":
		runCheck()
	case "clean":
		runClean()
	case "lib":
		runLib(os.Args[2:])
	case "lsp":
		runLSP()
	case "doc":
		runDoc(os.Args[2:])
	case "fmt":
		runFmt(os.Args[2:])
	case "targets":
		runTargets()
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runInit(args []string) {
	initFlags := flag.NewFlagSet("init", flag.ExitOnError)
	libFlag := initFlags.Bool("lib", false, "Initialize a library project instead of an executable binary")
	initFlags.BoolVar(libFlag, "l", false, "Initialize a library project (shorthand)")
	// Pre-process args to allow flags after positional project name
	var cleanArgs []string
	var nonFlags []string
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			cleanArgs = append(cleanArgs, args[i])
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				// Check if this flag takes an argument
				isBinary := true
				initFlags.VisitAll(func(f *flag.Flag) {
					if "-"+f.Name == args[i] || "--"+f.Name == args[i] {
						if _, ok := f.Value.(interface{ IsBoolFlag() bool }); !ok {
							isBinary = false
						}
					}
				})
				if !isBinary {
					cleanArgs = append(cleanArgs, args[i+1])
					i++
				}
			}
		} else {
			nonFlags = append(nonFlags, args[i])
		}
	}
	cleanArgs = append(cleanArgs, nonFlags...)

	initFlags.Parse(cleanArgs)

	if initFlags.NArg() < 1 {
		fmt.Println("Error: missing project name")
		fmt.Println("Usage: Nora init [flags] <project-name>")
		fmt.Println("Flags:")
		initFlags.PrintDefaults()
		os.Exit(1)
	}

	projectName := initFlags.Arg(0)
	if err := os.Mkdir(projectName, 0755); err != nil {
		fmt.Printf("Error creating project directory: %v\n", err)
		os.Exit(1)
	}

	srcDir := filepath.Join(projectName, "src")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		fmt.Printf("Error creating src directory: %v\n", err)
		os.Exit(1)
	}

	var entryFile string
	var config ProjectConfig

	if *libFlag {
		entryFile = "src/lib.nr"

		// Map library path so examples/ can import the library via local dependency
		deps := make(map[string]Dependency)
		deps[projectName] = Dependency{
			Path: "src",
		}

		config = ProjectConfig{
			Name:         projectName,
			Version:      "1.0.0",
			Language:     LanguageVersion,
			Entry:        entryFile,
			Output:       "lib" + projectName,
			Plugins:      []string{},
			Dependencies: deps,
		}

		// Create src/lib.nr
		libNora := fmt.Sprintf(`package %s

import "io"

pub fn hello() {
    io.println("Hello from %s library!")
}
`, projectName, projectName)

		if err := os.WriteFile(filepath.Join(projectName, entryFile), []byte(libNora), 0644); err != nil {
			fmt.Printf("Error creating %s: %v\n", entryFile, err)
			os.Exit(1)
		}

		// Automatically create examples/ directory and examples/basic.nr for premium DX!
		examplesDir := filepath.Join(projectName, "examples")
		if err := os.Mkdir(examplesDir, 0755); err != nil {
			fmt.Printf("Error creating examples directory: %v\n", err)
			os.Exit(1)
		}

		exampleNora := fmt.Sprintf(`package main

import "io"
import "%s"

fn main() {
    io.println("Running example program...")
    %s.hello()
}
`, projectName, projectName)

		if err := os.WriteFile(filepath.Join(examplesDir, "basic.nr"), []byte(exampleNora), 0644); err != nil {
			fmt.Printf("Error creating examples/basic.nr: %v\n", err)
			os.Exit(1)
		}

		configData, _ := yaml.Marshal(config)
		if err := os.WriteFile(filepath.Join(projectName, "nora.yaml"), configData, 0644); err != nil {
			fmt.Printf("Error creating nora.yaml: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Library project '%s' initialized successfully.\n", projectName)
		fmt.Printf("Directory layout:\n")
		fmt.Printf("  %s/\n", projectName)
		fmt.Printf("    nora.yaml        (manifest)\n")
		fmt.Printf("    src/lib.nr       (library source code)\n")
		fmt.Printf("    examples/basic.nr (runnable example importing your library)\n\n")
		fmt.Printf("To test and run the library example:\n")
		fmt.Printf("  cd %s && Nora run --example basic\n", projectName)

	} else {
		entryFile = "src/main.nr"
		config = ProjectConfig{
			Name:         projectName,
			Version:      "1.0.0",
			Language:     LanguageVersion,
			Entry:        entryFile,
			Output:       projectName,
			Plugins:      []string{},
			Dependencies: make(map[string]Dependency),
		}

		configData, _ := yaml.Marshal(config)
		if err := os.WriteFile(filepath.Join(projectName, "nora.yaml"), configData, 0644); err != nil {
			fmt.Printf("Error creating nora.yaml: %v\n", err)
			os.Exit(1)
		}

		mainNora := `package main

import "io"

fn main() {
    io.println("Hello, Nora Project!")
}
`
		if err := os.WriteFile(filepath.Join(projectName, entryFile), []byte(mainNora), 0644); err != nil {
			fmt.Printf("Error creating %s: %v\n", entryFile, err)
			os.Exit(1)
		}

		fmt.Printf("Project '%s' initialized successfully.\n", projectName)
		fmt.Printf("To build: cd %s && Nora build\n", projectName)
	}
}

func runBuild(args []string) {
	buildFlags := flag.NewFlagSet("build", flag.ExitOnError)

	outputFile := buildFlags.String("o", "", "Output executable name")
	pluginFlag := buildFlags.String("p", "", "Comma-separated list of additional plugins")
	targetFlag := buildFlags.String("target", "", "Target platform (e.g., windows-amd64, wasm, wasi)")
	wasmFlag := buildFlags.Bool("wasm", false, "Target WebAssembly (shorthand for --target wasm)")
	wasiFlag := buildFlags.Bool("wasi", false, "Target WebAssembly WASI (shorthand for --target wasi)")
	releaseFlag := buildFlags.Bool("release", false, "Build in release mode")
	debugFlag := buildFlags.Bool("debug", false, "Build in debug mode")
	debugMemFlag := buildFlags.Bool("debug-memory", false, "Enable runtime memory leak tracking")
	debugSemanticFlag := buildFlags.Bool("debug-semantic", false, "Enable semantic analyzer debug logs")
	debugTopologyFlag := buildFlags.Bool("debug-topology", false, "Enable topology solver debug logs")
	fiberReportFlag := buildFlags.Bool("fiber-report", false, "Enable fiber lifecycle reporting at exit")
	gFlag := buildFlags.Bool("g", false, "Enable debug symbols and source mapping")
	wasmExpFlag := buildFlags.Bool("wasm-experimental", false, "Enable experimental Wasm features (Native Stack Switching)")
	allowUnsafeFlag := buildFlags.Bool("allow-unsafe", false, "Allow [unsafe] attribute usage")
	noStdlibFlag := buildFlags.Bool("no-stdlib", false, "Disable automatic injection of the standard library prelude")
	noCoreFlag := buildFlags.Bool("no-core", false, "Disable automatic injection of the core prelude")
	ccFlag := buildFlags.String("cc", "", "Specify C compiler override (e.g., clang, gcc, cl, tcc)")
	cflagsFlag := buildFlags.String("cflags", "", "Custom flags passed directly to the C compiler (quote multiple flags)")
	verboseFlag := buildFlags.Bool("verbose", false, "Show detailed compilation logs and commands")
	exampleFlag := buildFlags.String("example", "", "Build a specific example from the examples/ directory")

	// Shorthands
	buildFlags.BoolVar(releaseFlag, "r", false, "Build in release mode (shorthand)")
	buildFlags.BoolVar(debugFlag, "d", false, "Build in debug mode (shorthand)")

	// Pre-process args to allow flags after filename
	var cleanArgs []string
	var nonFlags []string
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			cleanArgs = append(cleanArgs, args[i])
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				// Check if this flag takes an argument
				isBinary := true
				buildFlags.VisitAll(func(f *flag.Flag) {
					if "-"+f.Name == args[i] || "--"+f.Name == args[i] {
						if _, ok := f.Value.(interface{ IsBoolFlag() bool }); !ok {
							isBinary = false
						}
					}
				})
				if !isBinary {
					cleanArgs = append(cleanArgs, args[i+1])
					i++
				}
			}
		} else {
			nonFlags = append(nonFlags, args[i])
		}
	}
	cleanArgs = append(cleanArgs, nonFlags...)

	buildFlags.Parse(cleanArgs)

	// Target logic
	var t target.Platform
	if *targetFlag != "" {
		var ok bool
		t, ok = target.Get(*targetFlag)
		if !ok {
			fmt.Printf("Error: unknown target '%s'. Run 'Nora targets' to see available platforms.\n", *targetFlag)
			os.Exit(1)
		}
	} else if *wasiFlag {
		t, _ = target.Get("wasi")
	} else if *wasmFlag {
		t, _ = target.Get("wasm")
	} else {
		t = target.Discover()
	}

	// Default mode logic
	isRelease := *releaseFlag
	enableDebug := *gFlag || *debugFlag || !isRelease

	var inputFile string
	var pluginPaths []string
	var exeName string

	var config *ProjectConfig

	var dependencies map[string]Dependency

	var nativeConfig NativeConfig

	if *exampleFlag != "" {
		candidate1 := filepath.Join("examples", *exampleFlag+".nr")
		candidate2 := filepath.Join("examples", *exampleFlag, "main.nr")
		if _, err := os.Stat(candidate1); err == nil {
			inputFile = candidate1
		} else if _, err := os.Stat(candidate2); err == nil {
			inputFile = candidate2
		} else {
			fmt.Printf("Error: example '%s' not found in examples/ directory. Checked: %s and %s\n", *exampleFlag, candidate1, candidate2)
			os.Exit(1)
		}
		var err error
		if config, err = LoadProjectConfig(); err == nil {
			dependencies = config.Dependencies
			nativeConfig = config.Native
			pluginPaths = append(pluginPaths, config.Plugins...)
			exeName = *exampleFlag
		}
		if *outputFile != "" {
			exeName = *outputFile
		}
	} else {
		var err error
		config, err = LoadProjectConfig()
		if err == nil {
			pluginPaths = append(pluginPaths, config.Plugins...)
			dependencies = config.Dependencies
			nativeConfig = config.Native
			exeName = config.Output
			inputFile = config.Entry
		}

		if buildFlags.NArg() >= 1 {
			inputFile = buildFlags.Arg(0)
			if err == nil && *outputFile == "" {
				// If building a specific file but we are in a project,
				// we still respect the project's output name unless overridden.
				exeName = config.Output
			} else if *outputFile == "" {
				// Fallback to source file basename without extension
				base := filepath.Base(inputFile)
				exeName = strings.TrimSuffix(base, filepath.Ext(base))
			}
		} else {
			if err != nil {
				fmt.Println("Error: no input file specified and nora.yaml not found")
				fmt.Println("Usage: Nora build [flags] <filename>")
				os.Exit(1)
			}
			if config.Language != "" && config.Language > LanguageVersion {
				fmt.Printf("Warning: project requires language version %s, but current version is %s. Build may fail.\n", config.Language, LanguageVersion)
			}
		}

		if *outputFile != "" {
			exeName = *outputFile
		}
		if *pluginFlag != "" {
			pluginPaths = append(pluginPaths, strings.Split(*pluginFlag, ",")...)
		}
	}

	if *outputFile != "" {
		exeName = *outputFile
	}
	if *pluginFlag != "" {
		pluginPaths = append(pluginPaths, strings.Split(*pluginFlag, ",")...)
	}

	var cliCFlags []string
	if *cflagsFlag != "" {
		cliCFlags = strings.Fields(*cflagsFlag)
	}

	opts := BuildOptions{
		Target:           t,
		Release:          isRelease,
		Debug:            !isRelease || *debugFlag,
		DebugMemory:      *debugMemFlag,
		DebugSemantic:    *debugSemanticFlag,
		DebugTopology:    *debugTopologyFlag,
		DebugFiber:       *fiberReportFlag,
		G:                enableDebug,
		WasmExperimental: *wasmExpFlag,
		AllowUnsafe:      *allowUnsafeFlag,
		NoStdlib:         *noStdlibFlag || (config != nil && config.NoStdlib),
		NoCore:           *noCoreFlag || (config != nil && config.NoCore),
		Native:           nativeConfig,
		Compiler:         *ccFlag,
		CFlags:           cliCFlags,
		Verbose:          *verboseFlag,
	}

	outC, finalExe, err := compile(inputFile, exeName, pluginPaths, dependencies, opts)
	if err != nil {
		fmt.Printf("Build Failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Success! Built '%s' (via %s)\n", finalExe, outC)
}

func runRun(args []string) {
	runFlags := flag.NewFlagSet("run", flag.ExitOnError)

	pluginFlag := runFlags.String("p", "", "Comma-separated list of additional plugins")
	targetFlag := runFlags.String("target", "", "Target platform")
	wasmFlag := runFlags.Bool("wasm", false, "Run as WebAssembly (via internal runner)")
	wasiFlag := runFlags.Bool("wasi", false, "Run as WebAssembly WASI")
	debugMemFlag := runFlags.Bool("debug-memory", false, "Enable runtime memory leak tracking")
	debugSemanticFlag := runFlags.Bool("debug-semantic", false, "Enable semantic analyzer debug logs")
	debugTopologyFlag := runFlags.Bool("debug-topology", false, "Enable topology solver debug logs")
	fiberReportFlag := runFlags.Bool("fiber-report", false, "Enable fiber lifecycle reporting at exit")
	gFlag := runFlags.Bool("g", false, "Enable debug symbols and source mapping")
	wasmExpFlag := runFlags.Bool("wasm-experimental", false, "Enable experimental Wasm features (Native Stack Switching)")
	allowUnsafeFlag := runFlags.Bool("allow-unsafe", false, "Allow [unsafe] attribute usage")
	noStdlibFlag := runFlags.Bool("no-stdlib", false, "Disable automatic injection of the standard library prelude")
	noCoreFlag := runFlags.Bool("no-core", false, "Disable automatic injection of the core prelude")
	ccFlag := runFlags.String("cc", "", "Specify C compiler override (e.g., clang, gcc, cl, tcc)")
	cflagsFlag := runFlags.String("cflags", "", "Custom flags passed directly to the C compiler (quote multiple flags)")
	verboseFlag := runFlags.Bool("verbose", false, "Show detailed compilation logs and commands")
	exampleFlag := runFlags.String("example", "", "Run a specific example from the examples/ directory")

	// Pre-process args to allow flags after filename
	var cleanArgs []string
	var nonFlags []string
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			cleanArgs = append(cleanArgs, args[i])
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				// Check if this flag takes an argument
				isBinary := true
				runFlags.VisitAll(func(f *flag.Flag) {
					if "-"+f.Name == args[i] || "--"+f.Name == args[i] {
						if _, ok := f.Value.(interface{ IsBoolFlag() bool }); !ok {
							isBinary = false
						}
					}
				})
				if !isBinary {
					cleanArgs = append(cleanArgs, args[i+1])
					i++
				}
			}
		} else {
			nonFlags = append(nonFlags, args[i])
		}
	}
	cleanArgs = append(cleanArgs, nonFlags...)

	runFlags.Parse(cleanArgs)

	var t target.Platform
	if *targetFlag != "" {
		var ok bool
		t, ok = target.Get(*targetFlag)
		if !ok {
			fmt.Printf("Error: unknown target '%s'\n", *targetFlag)
			os.Exit(1)
		}
	} else if *wasiFlag {
		t, _ = target.Get("wasi")
	} else if *wasmFlag {
		t, _ = target.Get("wasm")
	} else {
		t = target.Discover()
	}

	var inputFile string
	var pluginPaths []string

	var config *ProjectConfig

	var dependencies map[string]Dependency

	var nativeConfig NativeConfig

	if *exampleFlag != "" {
		candidate1 := filepath.Join("examples", *exampleFlag+".nr")
		candidate2 := filepath.Join("examples", *exampleFlag, "main.nr")
		if _, err := os.Stat(candidate1); err == nil {
			inputFile = candidate1
		} else if _, err := os.Stat(candidate2); err == nil {
			inputFile = candidate2
		} else {
			fmt.Printf("Error: example '%s' not found in examples/ directory. Checked: %s and %s\n", *exampleFlag, candidate1, candidate2)
			os.Exit(1)
		}
		var err error
		if config, err = LoadProjectConfig(); err == nil {
			dependencies = config.Dependencies
			nativeConfig = config.Native
			pluginPaths = append(pluginPaths, config.Plugins...)
		}
	} else {
		var err error
		config, err = LoadProjectConfig()
		if err == nil {
			pluginPaths = append(pluginPaths, config.Plugins...)
			dependencies = config.Dependencies
			nativeConfig = config.Native
			inputFile = config.Entry
		}

		if runFlags.NArg() >= 1 {
			inputFile = runFlags.Arg(0)
		} else {
			if err != nil {
				fmt.Println("Error: no input file specified and nora.yaml not found")
				fmt.Println("Usage: Nora run [flags] <filename>")
				os.Exit(1)
			}
			if config.Language != "" && config.Language > LanguageVersion {
				fmt.Printf("Warning: project requires language version %s, but current version is %s. Execution may fail.\n", config.Language, LanguageVersion)
			}
		}

		if *pluginFlag != "" {
			pluginPaths = append(pluginPaths, strings.Split(*pluginFlag, ",")...)
		}
	}

	if *pluginFlag != "" {
		pluginPaths = append(pluginPaths, strings.Split(*pluginFlag, ",")...)
	}

	var cliCFlags []string
	if *cflagsFlag != "" {
		cliCFlags = strings.Fields(*cflagsFlag)
	}

	opts := BuildOptions{
		Target:           t,
		Debug:            true,
		DebugMemory:      *debugMemFlag,
		DebugSemantic:    *debugSemanticFlag,
		DebugTopology:    *debugTopologyFlag,
		DebugFiber:       *fiberReportFlag,
		G:                *gFlag,
		WasmExperimental: *wasmExpFlag,
		AllowUnsafe:      *allowUnsafeFlag,
		NoStdlib:         *noStdlibFlag || (config != nil && config.NoStdlib),
		NoCore:           *noCoreFlag || (config != nil && config.NoCore),
		Native:           nativeConfig,
		Compiler:         *ccFlag,
		CFlags:           cliCFlags,
		Verbose:          *verboseFlag,
	}
	_, exeName, err := compile(inputFile, "", pluginPaths, dependencies, opts)
	if err != nil {
		fmt.Printf("Run Failed: %v\n", err)
		os.Exit(1)
	}

	runCmdStr := "./" + exeName
	if runtime.GOOS == "windows" {
		runCmdStr = ".\\" + exeName
	}

	var runArgs []string
	if runFlags.NArg() > 0 {
		runArgs = runFlags.Args()[1:]
	} else {
		runArgs = runFlags.Args()
	}

	if t.Wasm {
		// Try to use wasmtime if available (supports experimental features like stack-switching/atomics)
		if _, err := exec.LookPath("wasmtime"); err == nil {
			fmt.Printf("Running Wasm module '%s' via wasmtime...\n", exeName)

			// Base features
			wargs := []string{"run", "-W", "threads=y", "-W", "bulk-memory=y"}
			if opts.WasmExperimental {
				wargs = append(wargs, "-W", "stack-switching=y")
			}
			wargs = append(wargs, exeName)
			wargs = append(wargs, runArgs...)

			cmd := exec.Command("wasmtime", wargs...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Stdin = os.Stdin
			if err := cmd.Run(); err != nil {
				fmt.Printf("Wasmtime Execution Error: %v\n", err)
				os.Exit(1)
			}
			return
		}

		fmt.Printf("Running Wasm module '%s' via internal WASI runner...\n", exeName)
		if err := runWasmStandalone(exeName); err != nil {
			fmt.Printf("Wasm Execution Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	runCmd := exec.Command(runCmdStr, runArgs...)
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	runCmd.Stdin = os.Stdin
	runCmd.Run()
}

func compile(inputFile string, exeName string, pluginPaths []string, dependencies map[string]Dependency, opts BuildOptions) (string, string, error) {
	prog := &ast.Program{Files: []*ast.File{}}
	diags := &diag.Collection{}

	if opts.Verbose {
		fmt.Printf("[Nora] Parsing source files from '%s'...\n", inputFile)
	} else {
		fmt.Println("  [Nora] Parsing source files...")
	}

	info, err := os.Stat(inputFile)
	if err != nil {
		if !strings.HasSuffix(inputFile, ".nr") {
			inputFile += ".nr"
			info, err = os.Stat(inputFile)
		}
	}

	if err != nil {
		return "", "", fmt.Errorf("error accessing input path '%s': %v", inputFile, err)
	}

	if info.IsDir() {
		files, err := os.ReadDir(inputFile)
		if err != nil {
			return "", "", err
		}
		for _, fileInfo := range files {
			if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".nr") {
				fullPath := filepath.Join(inputFile, fileInfo.Name())
				input, err := os.ReadFile(fullPath)
				if err != nil {
					continue
				}
				l := lexer.New(string(input), fullPath)
				l.Diagnostics = diags
				p := parser.New(l)
				p.AllowNoPackage = false
				file := p.Parse(fullPath)
				if p.Diagnostics.HasErrors() {
					diag.Report(p.Diagnostics)
					return "", "", fmt.Errorf("parser errors in %s", fullPath)
				}
				prog.Files = append(prog.Files, file)
			}
		}
	} else {
		input, err := os.ReadFile(inputFile)
		if err != nil {
			return "", "", fmt.Errorf("error reading file '%s': %v", inputFile, err)
		}

		l := lexer.New(string(input), inputFile)
		l.Diagnostics = diags
		p := parser.New(l)
		p.AllowNoPackage = false
		file := p.Parse(inputFile)
		if p.Diagnostics.HasErrors() {
			diag.Report(p.Diagnostics)
			return "", "", fmt.Errorf("parser errors")
		}
		prog.Files = append(prog.Files, file)
	}

	// Auto-load core/prelude.nr if it exists and not disabled
	if !opts.NoCore {
		preludePath := filepath.Join(CorePath, "prelude.nr")
		if _, err := os.Stat(preludePath); err == nil {
			preludeInput, err := os.ReadFile(preludePath)
			if err == nil {
				l := lexer.New(string(preludeInput), preludePath)
				l.Diagnostics = diags
				p := parser.New(l)
				p.AllowNoPackage = false
				preludeFile := p.Parse(preludePath)
				if !p.Diagnostics.HasErrors() {
					prog.Files = append(prog.Files, preludeFile)
				}
			}
		}
	}

	if len(prog.Files) == 0 {
		return "", "", fmt.Errorf("no .nr files found to compile")
	}

	pluginMgr := plugin.NewPluginManager()
	defer pluginMgr.Close()
	pluginMgr.RegisterBuiltinMacros()

	// Auto-load plugins from std/ and its subdirectories
	filepath.Walk(StdPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".wasm") {
			name := strings.TrimSuffix(info.Name(), ".wasm")
			if err := pluginMgr.LoadPlugin(name, path); err != nil {
				fmt.Printf("Warning: failed to auto-load plugin %s: %v\n", path, err)
			}
		}
		return nil
	})

	// Load manual plugins
	for _, p := range pluginPaths {
		if p == "" {
			continue
		}
		name := filepath.Base(p)
		name = strings.TrimSuffix(name, filepath.Ext(name))
		if err := pluginMgr.LoadPlugin(name, p); err != nil {
			return "", "", fmt.Errorf("error loading plugin %s: %v", p, err)
		}
	}

	if opts.Verbose {
		fmt.Println("[Nora] Processing plugin macros...")
	}
	for _, file := range prog.Files {
		if err := pluginMgr.ProcessMacroForFile(file); err != nil {
			return "", "", fmt.Errorf("plugin error: %v", err)
		}
	}

	if opts.Verbose {
		fmt.Println("[Nora] Running semantic analyzer & scope resolution...")
	} else {
		fmt.Println("  [Nora] Running semantic analysis...")
	}

	analyzer := semantic.NewAnalyzer()
	analyzer.DebugMode = opts.DebugSemantic
	analyzer.AllowUnsafe = opts.AllowUnsafe
	analyzer.Diagnostics = diags // Share diagnostics collection
	loader := &FileLoader{
		Cache:             make(map[string]*semantic.Scope),
		ParsedFiles:       make(map[string]*ast.File),
		Program:           prog,
		Analyzer:          analyzer,
		Dependencies:      dependencies,
		PluginManager:     pluginMgr,
		AllowedUnsafeDirs: make([]string, 0),
	}
	if loader.Dependencies == nil {
		loader.Dependencies = make(map[string]Dependency)
	}
	analyzer.Loader = loader
	analyzer.AllowedUnsafeDirs = loader.AllowedUnsafeDirs

	if !opts.NoCore {
		loader.loadManifest(CorePath)
	}

	// Explicitly load the standard library's manifest to gather its Native bindings (e.g. runtime headers)
	if !opts.NoStdlib {
		loader.loadManifest(StdPath)
	}

	// Pre-populate ParsedFiles with the files we already loaded
	for _, f := range prog.Files {
		loader.ParsedFiles[filepath.Clean(f.Name)] = f
	}
	analyzer.AllowedUnsafeDirs = loader.AllowedUnsafeDirs
	analyzer.Analyze(prog)

	if len(analyzer.Diagnostics.Diagnostics) > 0 {
		diag.Report(analyzer.Diagnostics)
		if analyzer.Diagnostics.HasErrors() {
			return "", "", fmt.Errorf("semantic errors")
		}
	}

	if opts.Verbose {
		fmt.Println("[Nora] Solving declaration dependency topology...")
	} else {
		fmt.Println("  [Nora] Resolving dependency topology...")
	}

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.Diagnostics = diags // Share diagnostics collection
	solver.DebugMode = opts.DebugTopology
	solver.Solve(prog)

	if solver.Diagnostics.HasErrors() {
		diag.Report(solver.Diagnostics)
		return "", "", fmt.Errorf("topology errors")
	}

	// Resolve active C compiler name
	compilerName := opts.Compiler
	if compilerName == "" {
		compilerName = opts.Native.Compiler
	}
	if compilerName == "" {
		compilerName = opts.Target.Compiler
	}

	if err := checkToolchain(compilerName, opts.Target); err != nil {
		return "", "", err
	}

	// 1. Identify the compiler family / select default template
	compilerBase := strings.ToLower(filepath.Base(compilerName))
	isMSVC := strings.HasPrefix(compilerBase, "cl") && !strings.HasPrefix(compilerBase, "clang")

	var template CompilerConfig
	if isMSVC {
		template = CompilerConfig{
			Compiler:     compilerName,
			OptRelease:   "/O2",
			OptDebug:     "/Od",
			DebugSymbols: "/Zi",
			OutFlag:      "/Fe:",
			IncFlag:      "/I",
			DefineFlag:   "/D",
			LibDirFlag:   "/link /LIBPATH:",
		}
	} else {
		template = CompilerConfig{
			Compiler:     compilerName,
			OptRelease:   "-O3",
			OptDebug:     "-O0",
			DebugSymbols: "-g",
			OutFlag:      "-o",
			IncFlag:      "-I",
			DefineFlag:   "-D",
			LibDirFlag:   "-L",
		}
	}

	// 2. Override template with manifest/CLI custom fields
	activeConfig := template
	if opts.Native.Compiler != "" {
		activeConfig.Compiler = opts.Native.Compiler
	}
	if opts.Compiler != "" {
		activeConfig.Compiler = opts.Compiler
	}

	if opts.Native.OptRelease != "" {
		activeConfig.OptRelease = opts.Native.OptRelease
	}
	if opts.Native.OptDebug != "" {
		activeConfig.OptDebug = opts.Native.OptDebug
	}
	if opts.Native.DebugSymbols != "" {
		activeConfig.DebugSymbols = opts.Native.DebugSymbols
	}
	if opts.Native.OutFlag != "" {
		activeConfig.OutFlag = opts.Native.OutFlag
	}
	if opts.Native.IncFlag != "" {
		activeConfig.IncFlag = opts.Native.IncFlag
	}
	if opts.Native.DefineFlag != "" {
		activeConfig.DefineFlag = opts.Native.DefineFlag
	}
	if opts.Native.LibDirFlag != "" {
		activeConfig.LibDirFlag = opts.Native.LibDirFlag
	}

	// Merge cflags
	for _, cf := range opts.Native.CFlags {
		if !contains(activeConfig.CFlags, cf) {
			activeConfig.CFlags = append(activeConfig.CFlags, cf)
		}
	}
	for _, cf := range opts.CFlags {
		if !contains(activeConfig.CFlags, cf) {
			activeConfig.CFlags = append(activeConfig.CFlags, cf)
		}
	}

	// 3. Merge dynamically collected native configurations from package manifests
	for _, dir := range loader.CollectedNative.IncludeDirs {
		if !contains(opts.Native.IncludeDirs, dir) {
			opts.Native.IncludeDirs = append(opts.Native.IncludeDirs, dir)
		}
	}
	for _, dir := range loader.CollectedNative.LibDirs {
		if !contains(opts.Native.LibDirs, dir) {
			opts.Native.LibDirs = append(opts.Native.LibDirs, dir)
		}
	}
	for _, lib := range loader.CollectedNative.DynamicLibs {
		if !contains(opts.Native.DynamicLibs, lib) {
			opts.Native.DynamicLibs = append(opts.Native.DynamicLibs, lib)
		}
	}
	for _, lib := range loader.CollectedNative.StaticLibs {
		if !contains(opts.Native.StaticLibs, lib) {
			opts.Native.StaticLibs = append(opts.Native.StaticLibs, lib)
		}
	}
	for _, src := range loader.CollectedNative.SourceFiles {
		if !contains(opts.Native.SourceFiles, src) {
			opts.Native.SourceFiles = append(opts.Native.SourceFiles, src)
		}
	}

	buildMode := "debug"
	if opts.Release {
		buildMode = "release"
	}
	buildDir := filepath.Join("build", buildMode)
	os.MkdirAll(buildDir, 0755)

	if exeName == "" {
		exeName = "main"
		if opts.Target.Wasm {
			exeName = "main.wasm"
		} else {
			exeName = "main" + opts.Target.ExeSuffix
		}
	} else if opts.Target.Wasm {
		if !strings.HasSuffix(exeName, ".wasm") {
			exeName += ".wasm"
		}
	} else if opts.Target.ExeSuffix != "" && !strings.HasSuffix(exeName, opts.Target.ExeSuffix) {
		exeName += opts.Target.ExeSuffix
	}
	finalExe := filepath.Join(buildDir, exeName)

	if opts.Verbose {
		fmt.Println("[Nora] Transpiling source AST using Package-Scoped Splitting...")
	} else {
		fmt.Println("  [Nora] Transpiling Package-Scoped Splitting backend...")
	}

	// Instantiate multi-file generator
	gen := codegen.NewGenerator(prog, &analyzer.SemanticInfo, solver, pluginMgr, loader.CollectedNative.Headers)
	gen.EnableDebug = opts.G
	gen.DebugMemory = opts.DebugMemory
	gen.DebugSemantic = opts.DebugSemantic
	gen.NoStdlib = opts.NoStdlib
	gen.Target = opts.Target.Name
	gen.MultiFileMode = true

	// Pre-populate structural AST definitions
	gen.CollectDefinitions()

	// Gather unique packages
	packages := make(map[string]bool)
	packageFiles := make(map[string][]string)
	packageDeps := make(map[string]map[string]bool)
	for _, f := range prog.Files {
		pkg := analyzer.GetPackageName(f)
		if pkg != "" {
			packages[pkg] = true
			packageFiles[pkg] = append(packageFiles[pkg], f.Name)
			if packageDeps[pkg] == nil {
				packageDeps[pkg] = make(map[string]bool)
			}
			for _, stmt := range f.Statements {
				if importStmt, ok := stmt.(*ast.ImportStatement); ok {
					packageDeps[pkg][importStmt.PathValue()] = true
				}
			}
		}
	}

	cacheDir := filepath.Join(buildDir, "runtime_cache")
	os.MkdirAll(cacheDir, 0755)

	// Load Cache Catalog
	catalog, err := loadBuildCache(buildDir)
	if err != nil {
		catalog = &BuildCacheCatalog{Packages: make(map[string]PackageCacheEntry)}
	}

	cflagsStr := strings.Join(activeConfig.CFlags, " ")
	configChanged := catalog.Compiler != compilerName ||
		catalog.Target != opts.Target.Name ||
		catalog.Release != opts.Release ||
		catalog.DebugMemory != opts.DebugMemory ||
		catalog.DebugFiber != opts.DebugFiber ||
		catalog.CFlags != cflagsStr

	if configChanged {
		if opts.Verbose {
			fmt.Println("[Nora] Build configuration changed. Invalidating incremental cache...")
		}
		catalog = &BuildCacheCatalog{
			Compiler:    compilerName,
			Target:      opts.Target.Name,
			Release:     opts.Release,
			DebugMemory: opts.DebugMemory,
			DebugFiber:  opts.DebugFiber,
			CFlags:      cflagsStr,
			Packages:    make(map[string]PackageCacheEntry),
		}
	}

	// Transpile package C files first to dynamically discover and populate g.AutoDropMethods, g.AutoEqMethods, and g.VTables!
	packageCodeMap := make(map[string]string)
	for pkg := range packages {
		pkgCode, err := gen.GeneratePackageCode(pkg)
		if err != nil {
			return "", "", fmt.Errorf("codegen error for package %s: %v", pkg, err)
		}
		packageCodeMap[pkg] = pkgCode
	}

	// Always generate shared contract header out.h BEFORE compiling package files so that C compilation can resolve includes!
	headerCode, err := gen.GenerateHeader()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate out.h: %v", err)
	}
	outH := filepath.Join(buildDir, "out.h")
	if err := os.WriteFile(outH, []byte(headerCode), 0644); err != nil {
		return "", "", fmt.Errorf("error writing out.h: %v", err)
	}

	var allPackageObjects []string
	var recompiledAny bool

	computedHashes := make(map[string]string)
	var getPackageHash func(pkgName string, visited map[string]bool) string
	getPackageHash = func(pkgName string, visited map[string]bool) string {
		if hash, ok := computedHashes[pkgName]; ok {
			return hash
		}
		if visited[pkgName] {
			return "" // Break cycle silently; semantic analyzer catches real cyclic imports
		}
		visited[pkgName] = true

		var sb strings.Builder

		// 1. Hash the package's own files
		files := packageFiles[pkgName]
		sort.Strings(files)
		for _, file := range files {
			info, err := os.Stat(file)
			if err != nil {
				continue
			}
			sb.WriteString(fmt.Sprintf("%s_%d_%d;", file, info.Size(), info.ModTime().UnixNano()))
		}

		// 2. Hash the dependencies' hashes
		var deps []string
		for dep := range packageDeps[pkgName] {
			deps = append(deps, dep)
		}
		sort.Strings(deps)
		for _, dep := range deps {
			depHash := getPackageHash(dep, visited)
			sb.WriteString(fmt.Sprintf("dep:%s=%s;", dep, depHash))
		}

		hashBytes := sha256.Sum256([]byte(sb.String()))
		hashStr := fmt.Sprintf("%x", hashBytes)[:16]
		computedHashes[pkgName] = hashStr
		return hashStr
	}

	// Compile or skip package files
	for pkg := range packages {
		currentHash := getPackageHash(pkg, make(map[string]bool))

		objExt := ".o"
		if isMSVC {
			objExt = ".obj"
		}
		safePkgName := strings.ReplaceAll(pkg, "/", "_")
		safePkgName = strings.ReplaceAll(safePkgName, ".", "_")

		objPath := filepath.Join(cacheDir, fmt.Sprintf("cache_pkg_%s%s", safePkgName, objExt))

		// Cache HIT validation
		cachedEntry, hasCache := catalog.Packages[pkg]
		objExists := false
		if hasCache {
			if _, err := os.Stat(cachedEntry.ObjectPath); err == nil {
				objExists = true
			}
		}

		useCache := hasCache && objExists && cachedEntry.Hash == currentHash && !configChanged

		if useCache {
			if opts.Verbose {
				fmt.Printf("[Nora] Cache HIT for package '%s'. Reusing object: %s\n", pkg, cachedEntry.ObjectPath)
			}
			allPackageObjects = append(allPackageObjects, cachedEntry.ObjectPath)
		} else {
			if opts.Verbose {
				fmt.Printf("[Nora] Cache MISS for package '%s'. Transpiling & Compiling...\n", pkg)
			} else {
				fmt.Printf("  [Nora] Compiling package: %s...\n", pkg)
			}

			recompiledAny = true
			pkgCFile := filepath.Join(buildDir, fmt.Sprintf("out_pkg_%s.c", safePkgName))
			if err := os.WriteFile(pkgCFile, []byte(packageCodeMap[pkg]), 0644); err != nil {
				return "", "", fmt.Errorf("error writing package C file: %v", err)
			}

			err := compileCToObject(compilerName, pkgCFile, objPath, isMSVC, activeConfig, opts)
			if err != nil {
				return "", "", fmt.Errorf("compilation failed for package %s: %v", pkg, err)
			}

			catalog.Packages[pkg] = PackageCacheEntry{
				Hash:       currentHash,
				ObjectPath: objPath,
			}
			allPackageObjects = append(allPackageObjects, objPath)
		}
	}

	// Always generate shared globals file
	globalsCode, err := gen.GenerateSharedGlobals()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate out_globals.c: %v", err)
	}
	outGlobalsC := filepath.Join(buildDir, "out_globals.c")
	if err := os.WriteFile(outGlobalsC, []byte(globalsCode), 0644); err != nil {
		return "", "", fmt.Errorf("error writing out_globals.c: %v", err)
	}

	objExt := ".o"
	if isMSVC {
		objExt = ".obj"
	}
	outGlobalsObj := filepath.Join(cacheDir, "out_globals"+objExt)

	globalsObjExists := false
	if _, err := os.Stat(outGlobalsObj); err == nil {
		globalsObjExists = true
	}

	globalsHash := fmt.Sprintf("%x", sha256.Sum256([]byte(globalsCode)))[:16]
	useGlobalsCache := globalsObjExists && catalog.GlobalsHash == globalsHash && !configChanged && !recompiledAny

	if useGlobalsCache {
		if opts.Verbose {
			fmt.Printf("[Nora] Cache HIT for shared globals. Reusing object: %s\n", outGlobalsObj)
		}
	} else {
		if opts.Verbose {
			fmt.Println("[Nora] Compiling shared globals...")
		}
		err := compileCToObject(compilerName, outGlobalsC, outGlobalsObj, isMSVC, activeConfig, opts)
		if err != nil {
			return "", "", fmt.Errorf("failed to compile shared globals: %v", err)
		}
		catalog.GlobalsHash = globalsHash
		catalog.GlobalsObjectPath = outGlobalsObj
	}

	if err := catalog.Save(buildDir); err != nil {
		if opts.Verbose {
			fmt.Printf("Warning: failed to save build cache catalog: %v\n", err)
		}
	}

	// Define all input object files to link
	var allInputObjects []string
	allInputObjects = append(allInputObjects, allPackageObjects...)
	allInputObjects = append(allInputObjects, outGlobalsObj)

	outC := filepath.Join(buildDir, "out_pkg_main.c")
	var buildCmd *exec.Cmd

	// 3.5 Dynamic Native C Source File Compilation Cache

	for i, srcPath := range opts.Native.SourceFiles {
		// Clean and get absolute path of source file to avoid relative path confusion
		absSrcPath, err := filepath.Abs(srcPath)
		if err != nil {
			absSrcPath = filepath.Clean(srcPath)
		}

		info, err := os.Stat(absSrcPath)
		if err != nil {
			// If file does not exist, let the regular compiler handle the error
			continue
		}

		// Calculate a hash/key of the file properties (path, size, modification time)
		fileKey := fmt.Sprintf("%s_%d_%d", absSrcPath, info.Size(), info.ModTime().UnixNano())
		hashBytes := sha256.Sum256([]byte(fileKey))
		hashKey := fmt.Sprintf("%x", hashBytes)[:16]

		compilerBaseName := filepath.Base(activeConfig.Compiler)
		compilerBaseName = strings.TrimSuffix(compilerBaseName, filepath.Ext(compilerBaseName))

		mode := "debug"
		if opts.Release {
			mode = "release"
		}

		mem := "0"
		if opts.DebugMemory {
			mem = "1"
		}

		fiber := "0"
		if opts.DebugFiber {
			fiber = "1"
		}

		wasmExp := "0"
		if opts.WasmExperimental {
			wasmExp = "1"
		}

		targetName := opts.Target.Name

		objExt := ".o"
		if isMSVC {
			objExt = ".obj"
		}

		// E.g. cache_[src_basename]_[compiler]_[mode]_[target]_memX_[hash].o
		srcBase := filepath.Base(srcPath)
		srcBase = strings.TrimSuffix(srcBase, filepath.Ext(srcBase))

		cacheName := fmt.Sprintf("cache_%s_%s_%s_%s_mem%s_fib%s_wasm%s_%s%s",
			srcBase, compilerBaseName, targetName, mode, mem, fiber, wasmExp, hashKey, objExt)

		cachedObjPath := filepath.Join(cacheDir, cacheName)

		if _, err := os.Stat(cachedObjPath); os.IsNotExist(err) {
			if opts.Verbose {
				fmt.Printf("[Nora] Compiling C source dependency to cache: %s -> %s\n", srcPath, cachedObjPath)
			} else {
				fmt.Printf("  [Nora] Compiling C source: %s (one-time cache)...\n", srcPath)
			}

			var objArgs []string
			if isMSVC {
				objArgs = append(objArgs, "/c", absSrcPath)
				objOutFlag := "/Fo:"
				if strings.HasSuffix(objOutFlag, ":") {
					objArgs = append(objArgs, objOutFlag+cachedObjPath)
				} else {
					objArgs = append(objArgs, objOutFlag, cachedObjPath)
				}
			} else {
				objArgs = append(objArgs, "-c", absSrcPath)
				if strings.HasSuffix(activeConfig.OutFlag, ":") {
					objArgs = append(objArgs, activeConfig.OutFlag+cachedObjPath)
				} else {
					objArgs = append(objArgs, activeConfig.OutFlag, cachedObjPath)
				}
			}

			// Optimization and Debug symbols
			if opts.Release {
				if activeConfig.OptRelease != "" {
					objArgs = append(objArgs, activeConfig.OptRelease)
				}
			} else {
				if activeConfig.OptDebug != "" {
					objArgs = append(objArgs, activeConfig.OptDebug)
				}
			}
			if opts.G && activeConfig.DebugSymbols != "" {
				objArgs = append(objArgs, activeConfig.DebugSymbols)
			}

			// Macro defines
			definePrefix := activeConfig.DefineFlag
			if opts.DebugFiber {
				objArgs = append(objArgs, definePrefix+"NR_DEBUG_FIBER")
			}
			if opts.DebugMemory {
				objArgs = append(objArgs, definePrefix+"NR_DEBUG_MEM=1")
			}

			// Include directories
			for _, dir := range opts.Native.IncludeDirs {
				objArgs = append(objArgs, activeConfig.IncFlag+dir)
			}

			// Custom compiler flags
			objArgs = append(objArgs, activeConfig.CFlags...)

			// Wasm experimental or specific sysroots
			if opts.Target.Wasm && (strings.Contains(strings.ToLower(activeConfig.Compiler), "clang") || strings.Contains(strings.ToLower(activeConfig.Compiler), "wasi-sdk")) {
				if sdkPath := os.Getenv("WASI_SDK_PATH"); sdkPath != "" {
					objArgs = append(objArgs, "--sysroot="+filepath.Join(sdkPath, "share", "wasi-sysroot"))
					if opts.WasmExperimental {
						objArgs = append(objArgs, "-DNR_USE_FIBERS")
					}
				}
			}

			if opts.Verbose {
				fmt.Printf("[Nora] Compile command: %s %s\n", activeConfig.Compiler, strings.Join(objArgs, " "))
			}

			cmd := exec.Command(activeConfig.Compiler, objArgs...)
			var stdoutBuf, stderrBuf bytes.Buffer
			cmd.Stdout = &stdoutBuf
			cmd.Stderr = &stderrBuf
			runErr := cmd.Run()
			if opts.Verbose {
				os.Stdout.Write(stdoutBuf.Bytes())
				os.Stderr.Write(stderrBuf.Bytes())
			}
			if runErr != nil {
				if !opts.Verbose {
					os.Stdout.Write(stdoutBuf.Bytes())
					os.Stderr.Write(stderrBuf.Bytes())
				}
				return "", "", fmt.Errorf("failed to compile standard C runtime: %v", runErr)
			}
		} else {
			if opts.Verbose {
				fmt.Printf("[Nora] Using cached object for dependency: %s (%s)\n", srcPath, cachedObjPath)
			}
		}

		// Swap the source file path with the cached object file path
		opts.Native.SourceFiles[i] = cachedObjPath
	}

	// 4. Parse target CFlags to extract libraries/linker-paths for MSVC or preserve warning/target options for POSIX
	var filteredTargetFlags []string
	for _, flag := range opts.Target.CFlags {
		if flag == "-O3" || flag == "-O0" || flag == "-g" || flag == "-s" {
			continue
		}
		if isMSVC {
			if strings.HasPrefix(flag, "-l") {
				libName := strings.TrimPrefix(flag, "-l")
				if !contains(opts.Native.DynamicLibs, libName) {
					opts.Native.DynamicLibs = append(opts.Native.DynamicLibs, libName)
				}
			} else if strings.HasPrefix(flag, "-L") {
				dir := strings.TrimPrefix(flag, "-L")
				if !contains(opts.Native.LibDirs, dir) {
					opts.Native.LibDirs = append(opts.Native.LibDirs, dir)
				}
			}
		} else {
			filteredTargetFlags = append(filteredTargetFlags, flag)
		}
	}

	// 5. Construct args slice based on compiler family
	var args []string

	if isMSVC {
		// Input files (packages and globals)
		for _, obj := range allInputObjects {
			args = append(args, obj)
		}

		// Output flag and value
		if strings.HasSuffix(activeConfig.OutFlag, ":") {
			args = append(args, activeConfig.OutFlag+finalExe)
		} else {
			args = append(args, activeConfig.OutFlag, finalExe)
		}

		// Optimization and Debug Symbols
		if opts.Release {
			if activeConfig.OptRelease != "" {
				args = append(args, activeConfig.OptRelease)
			}
		} else {
			if activeConfig.OptDebug != "" {
				args = append(args, activeConfig.OptDebug)
			}
		}
		if opts.G && activeConfig.DebugSymbols != "" {
			args = append(args, activeConfig.DebugSymbols)
		}

		// Macro defines
		definePrefix := activeConfig.DefineFlag
		if opts.DebugFiber {
			args = append(args, definePrefix+"NR_DEBUG_FIBER")
		}
		if opts.DebugMemory {
			args = append(args, definePrefix+"NR_DEBUG_MEM=1")
		}

		// Include directories
		for _, dir := range opts.Native.IncludeDirs {
			args = append(args, activeConfig.IncFlag+dir)
		}

		// Native source files
		for _, src := range opts.Native.SourceFiles {
			args = append(args, src)
		}

		// Custom/Extra cflags
		args = append(args, activeConfig.CFlags...)

		// Linker arguments (MSVC link options go after /link)
		var linkerArgs []string
		pathFlag := "/LIBPATH:"
		parts := strings.Fields(activeConfig.LibDirFlag)
		if len(parts) > 0 {
			lastPart := parts[len(parts)-1]
			if strings.HasPrefix(lastPart, "/") || strings.HasPrefix(lastPart, "-") {
				pathFlag = lastPart
			}
		}

		for _, dir := range opts.Native.LibDirs {
			linkerArgs = append(linkerArgs, pathFlag+dir)
		}
		for _, lib := range opts.Native.DynamicLibs {
			if !strings.HasSuffix(strings.ToLower(lib), ".lib") {
				linkerArgs = append(linkerArgs, lib+".lib")
			} else {
				linkerArgs = append(linkerArgs, lib)
			}
		}
		for _, lib := range opts.Native.StaticLibs {
			linkerArgs = append(linkerArgs, lib)
		}

		if len(linkerArgs) > 0 {
			args = append(args, "/link")
			args = append(args, linkerArgs...)
		}
	} else {
		// Input files (packages and globals)
		for _, obj := range allInputObjects {
			args = append(args, obj)
		}

		// Output flag and value
		if strings.HasSuffix(activeConfig.OutFlag, ":") {
			args = append(args, activeConfig.OutFlag+finalExe)
		} else {
			args = append(args, activeConfig.OutFlag, finalExe)
		}

		// Split target-specific flags into compiler options and library/linker flags
		var targetOpts []string
		var targetLibs []string
		for _, flag := range filteredTargetFlags {
			if strings.HasPrefix(flag, "-l") || strings.HasPrefix(flag, "-L") || strings.HasSuffix(flag, ".a") || strings.HasSuffix(flag, ".lib") {
				targetLibs = append(targetLibs, flag)
			} else {
				targetOpts = append(targetOpts, flag)
			}
		}

		// Compiler Options
		args = append(args, targetOpts...)

		// Optimization and Debug Symbols
		if opts.Release {
			if activeConfig.OptRelease != "" {
				args = append(args, activeConfig.OptRelease)
			}
		} else {
			if activeConfig.OptDebug != "" {
				args = append(args, activeConfig.OptDebug)
			}
		}

		if opts.G {
			if activeConfig.DebugSymbols != "" {
				args = append(args, activeConfig.DebugSymbols)
			}
			if opts.Target.OS == "linux" {
				args = append(args, "-rdynamic")
			}
		} else {
			args = append(args, "-s")
		}

		// Macro defines
		definePrefix := activeConfig.DefineFlag
		if opts.DebugFiber {
			args = append(args, definePrefix+"NR_DEBUG_FIBER")
		}
		if opts.DebugMemory {
			args = append(args, definePrefix+"NR_DEBUG_MEM=1")
		}

		// Include directories
		for _, dir := range opts.Native.IncludeDirs {
			args = append(args, activeConfig.IncFlag+dir)
		}

		// Native source files (must precede library links for GCC/MinGW)
		for _, src := range opts.Native.SourceFiles {
			args = append(args, src)
		}

		// Target libraries (placed after sources)
		args = append(args, targetLibs...)

		// User Library directories
		for _, dir := range opts.Native.LibDirs {
			args = append(args, activeConfig.LibDirFlag+dir)
		}

		// User Dynamic and Static libraries
		for _, lib := range opts.Native.DynamicLibs {
			args = append(args, "-l"+lib)
		}
		for _, lib := range opts.Native.StaticLibs {
			args = append(args, lib)
		}

		// Custom/Extra cflags
		args = append(args, activeConfig.CFlags...)
	}

	// 6. Handle Wasm clang / WASI-SDK overrides if applicable
	if opts.Target.Wasm && (strings.Contains(strings.ToLower(activeConfig.Compiler), "clang") || strings.Contains(strings.ToLower(activeConfig.Compiler), "wasi-sdk")) {
		if sdkPath := os.Getenv("WASI_SDK_PATH"); sdkPath != "" {
			wasiClang := filepath.Join(sdkPath, "bin", "clang")
			if runtime.GOOS == "windows" {
				wasiClang += ".exe"
			}
			if _, err := os.Stat(wasiClang); err == nil {
				activeConfig.Compiler = wasiClang
			}
			args = append(args, "--sysroot="+filepath.Join(sdkPath, "share", "wasi-sysroot"))
			if opts.WasmExperimental {
				args = append(args, "-DNR_USE_FIBERS")
			}
		}
	}

	if opts.Verbose {
		fmt.Printf("[Nora] Invoking C toolchain: %s\n", activeConfig.Compiler)
		family := "POSIX (GCC/Clang/TCC)"
		if isMSVC {
			family = "MSVC (cl)"
		}
		fmt.Printf("[Nora] Compiler family: %s\n", family)
		fmt.Printf("[Nora] Command line arguments: %s %s\n", activeConfig.Compiler, strings.Join(args, " "))
	} else {
		fmt.Printf("  [Nora] Compiling with C toolchain: %s...\n", filepath.Base(activeConfig.Compiler))
	}

	buildCmd = exec.Command(activeConfig.Compiler, args...)

	var stdoutBuf, stderrBuf bytes.Buffer
	buildCmd.Stdout = &stdoutBuf
	buildCmd.Stderr = &stderrBuf
	runErr := buildCmd.Run()
	if opts.Verbose {
		os.Stdout.Write(stdoutBuf.Bytes())
		os.Stderr.Write(stderrBuf.Bytes())
	}
	if runErr != nil {
		if !opts.Verbose {
			os.Stdout.Write(stdoutBuf.Bytes())
			os.Stderr.Write(stderrBuf.Bytes())
		}
		fmt.Printf("\nDebug: Failed Command: %s %s\n", activeConfig.Compiler, strings.Join(args, " "))
		if sdkPath := os.Getenv("WASI_SDK_PATH"); sdkPath != "" {
			fmt.Printf("Debug: WASI_SDK_PATH is set to: %s\n", sdkPath)
		} else {
			fmt.Println("Debug: WASI_SDK_PATH is NOT set.")
		}
		return "", "", fmt.Errorf("%s build failed: %v", activeConfig.Compiler, runErr)
	}

	if err := copyDynamicLibraries(buildDir, opts); err != nil {
		return "", "", err
	}

	return outC, finalExe, nil
}

func checkToolchain(compiler string, t target.Platform) error {
	_, err := exec.LookPath(compiler)
	if err != nil {
		if t.Wasm {
			if compiler == "emcc" {
				return fmt.Errorf("emscripten (emcc) not found in PATH. Required for legacy Wasm builds.\nDownload it from: https://emscripten.org/docs/getting_started/downloads.html")
			}
			return fmt.Errorf("%s not found in PATH. Required for WASI/Wasm builds.\n1. Install LLVM/Clang.\n2. Download wasi-sdk (https://github.com/WebAssembly/wasi-sdk/releases).\n3. Set WASI_SDK_PATH environment variable.\n4. (Optional) Install wasmtime to run modules.", compiler)
		}
		return fmt.Errorf("%s not found in PATH. Required for %s builds.\nDownload it for your platform.", compiler, t.Name)
	}

	// Check for WASI sysroot if using clang for wasm
	if t.Wasm && strings.Contains(strings.ToLower(compiler), "clang") {
		if sdkPath := os.Getenv("WASI_SDK_PATH"); sdkPath == "" {
			fmt.Println("Warning: WASI_SDK_PATH not set. WASI builds may fail with 'stdio.h not found'.")
			fmt.Println("Please download wasi-sdk and set WASI_SDK_PATH to its root directory.")
		}
	}

	return nil
}

func runTargets() {
	fmt.Println("Available Targets:")
	host := target.Discover()
	for _, t := range target.List() {
		marker := "  "
		if t.Name == host.Name {
			marker = "* "
		}
		fmt.Printf("%s %-20s (OS: %-10s Arch: %-10s Compiler: %s)\n", marker, t.Name, t.OS, t.Arch, t.Compiler)
	}
	fmt.Println("\n* = detected host platform")
}

func runLSP() {
	s := lsp.NewServer()
	if err := s.Run(); err != nil {
		fmt.Printf("LSP Error: %v\n", err)
		os.Exit(1)
	}
}

func runTest(args []string) {
	testDir := "pkg/cmd/test"

	passed := 0
	failed := 0
	skipped := 0

	fmt.Printf("Running integration tests in %s...\n", testDir)
	startTime := time.Now()

	filepath.WalkDir(testDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".nr") {
			return nil
		}

		relPath, _ := filepath.Rel(testDir, path)
		fmt.Printf(" - %-30s ", relPath)

		testErr := runSingleTest(path)
		if testErr == nil {
			fmt.Println("[PASS]")
			passed++
		} else if strings.HasPrefix(testErr.Error(), "SKIP") {
			fmt.Printf("[SKIP] (%s)\n", strings.TrimPrefix(testErr.Error(), "SKIP: "))
			skipped++
		} else {
			fmt.Printf("[FAIL] %v\n", testErr)
			failed++
		}
		return nil
	})

	fmt.Printf("\nTest Summary:\n")
	fmt.Printf("  Passed:  %d\n", passed)
	fmt.Printf("  Failed:  %d\n", failed)
	fmt.Printf("  Skipped: %d\n", skipped)
	fmt.Printf("  Time:    %v\n", time.Since(startTime))

	if failed > 0 {
		os.Exit(1)
	}
}

func runSingleTest(path string) error {
	filename := filepath.Base(path)
	expectFail := strings.HasPrefix(filename, "fail_") || strings.Contains(filename, "violation")

	input, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	l := lexer.New(string(input), path)
	p := parser.New(l)
	p.AllowNoPackage = false
	file := p.Parse(path)
	if p.Diagnostics.HasErrors() {
		if expectFail {
			return nil
		}
		diag.Report(p.Diagnostics)
		return fmt.Errorf("parser errors")
	}
	prog := &ast.Program{Files: []*ast.File{file}}

	pm := plugin.NewPluginManager()
	defer pm.Close()
	pm.RegisterBuiltinMacros()

	// Try load io_macro for tests
	filepath.Walk(StdPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.Name() == "io_macro.wasm" {
			pm.LoadPlugin("io_macro", path)
		}
		if err == nil && info.Name() == "serialize.wasm" {
			pm.LoadPlugin("serialize", path)
		}
		return nil
	})

	for _, file := range prog.Files {
		if err := pm.ProcessMacroForFile(file); err != nil {
			fmt.Printf("Macro Error: %v\n", err)
		}
	}

	analyzer := semantic.NewAnalyzer()
	analyzer.Diagnostics = p.Diagnostics
	// We don't have opts here, it's Language Server. By default we might not allow unsafe, or maybe we allow it for LSP if it's the stdlib.
	// For now, let's allow it in LSP since stdlib uses it and we don't want false positives in IDE.
	analyzer.AllowUnsafe = true
	loader := &FileLoader{
		Cache:             make(map[string]*semantic.Scope),
		ParsedFiles:       make(map[string]*ast.File),
		Program:           prog,
		Analyzer:          analyzer,
		Dependencies:      make(map[string]Dependency),
		PluginManager:     pm,
		AllowedUnsafeDirs: make([]string, 0),
	}
	analyzer.Loader = loader
	analyzer.AllowedUnsafeDirs = loader.AllowedUnsafeDirs
	loader.loadManifest(StdPath)

	// Pre-populate ParsedFiles
	for _, f := range prog.Files {
		loader.ParsedFiles[filepath.Clean(f.Name)] = f
	}
	analyzer.AllowedUnsafeDirs = loader.AllowedUnsafeDirs
	analyzer.Analyze(prog)
	if analyzer.Diagnostics.HasErrors() {
		if expectFail {
			return nil
		}
		diag.Report(analyzer.Diagnostics)
		return fmt.Errorf("semantic errors")
	}

	solver := topology.NewSolver(&analyzer.SemanticInfo)
	solver.Diagnostics = analyzer.Diagnostics
	solver.Solve(prog)
	if solver.Diagnostics.HasErrors() {
		if expectFail {
			return nil
		}
		diag.Report(solver.Diagnostics)
		return fmt.Errorf("topology errors")
	}

	gen := codegen.NewGenerator(prog, &analyzer.SemanticInfo, solver, pm, loader.CollectedNative.Headers)
	gen.NoStdlib = loader.NoStdlib
	gen.Target = target.Discover().Name
	bodyCode, err := gen.Generate()
	if err != nil {
		if expectFail {
			return nil
		}
		return fmt.Errorf("codegen error: %v", err)
	}

	if expectFail {
		return fmt.Errorf("expected failure but passed analysis")
	}

	// GCC Compile check
	hasMain := false
	for _, f := range prog.Files {
		for _, stmt := range f.Statements {
			if fn, ok := stmt.(*ast.FunctionStatement); ok && fn.Name.Value == "main" {
				hasMain = true
				break
			}
		}
		if hasMain {
			break
		}
	}

	if !hasMain {
		return fmt.Errorf("SKIP: no main function")
	}

	tempC := "test_temp.c"
	tempExe := "test_temp"
	if runtime.GOOS == "windows" {
		tempExe += ".exe"
	}

	os.WriteFile(tempC, []byte(bodyCode), 0644)
	defer os.Remove(tempC)
	defer os.Remove(tempExe)

	args := []string{tempC, "-o", tempExe, "-I" + StdPath}
	if runtime.GOOS == "windows" {
		args = append(args, "-ldbghelp", "-lws2_32")
	}
	buildCmd := exec.Command("clang", args...)
	if out, err := buildCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("clang error: %v\n%s", err, string(out))
	}

	// Execution check (optional/skipped for Codeforces tests)
	interactivePrefixes := []string{"scan", "test_scanf", "4A", "71A", "112A", "158A", "231A", "236A", "263A", "282A", "339A", "50A"}
	for _, prefix := range interactivePrefixes {
		if strings.HasPrefix(filename, prefix) {
			return nil // Consider it a pass if it compiles
		}
	}

	runCmd := exec.Command("./" + tempExe)
	if runtime.GOOS == "windows" {
		runCmd = exec.Command(".\\" + tempExe)
	}

	done := make(chan error, 1)
	go func() { done <- runCmd.Run() }()

	select {
	case <-time.After(2 * time.Second):
		runCmd.Process.Kill()
		return fmt.Errorf("execution timeout")
	case err := <-done:
		if err != nil {
			if strings.Contains(filename, "panic") {
				return nil
			}
			return fmt.Errorf("runtime error: %v", err)
		}
	}

	return nil
}

func runClean() {
	files, _ := os.ReadDir(".")
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		name := f.Name()
		if name == "out.c" || name == "main.exe" || name == "main" || strings.HasSuffix(name, ".exe") && name != "nora.exe" {
			// Be careful not to delete source files
			if !strings.HasSuffix(name, ".nr") && !strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, ".mod") {
				fmt.Printf("Removing %s\n", name)
				os.Remove(name)
			}
		}
	}

	// Recursively remove the build directory if it exists
	if _, err := os.Stat("build"); err == nil {
		fmt.Println("Removing build directory...")
		if err := os.RemoveAll("build"); err != nil {
			fmt.Printf("Warning: failed to remove build directory: %v\n", err)
		}
	}

	fmt.Println("Cleaned up build artifacts.")
}

func runLib(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: Nora lib <add|remove|list> [args]")
		os.Exit(1)
	}

	subCommand := args[0]
	switch subCommand {
	case "add":
		runLibAdd(args[1:])
	case "remove":
		runLibRemove(args[1:])
	case "list":
		runLibList()
	default:
		fmt.Printf("Unknown lib command: %s\n", subCommand)
		os.Exit(1)
	}
}

func runLibAdd(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: Nora lib add <url|path> [name]")
		os.Exit(1)
	}

	target := args[0]
	requestedVersion := ""
	if strings.Contains(target, "@") {
		parts := strings.SplitN(target, "@", 2)
		target = parts[0]
		requestedVersion = parts[1]
	}

	name := ""
	if len(args) > 1 {
		name = args[1]
	}

	config, err := LoadProjectConfig()
	if err != nil {
		fmt.Printf("Error: not in a Nora project (nora.yaml not found)\n")
		os.Exit(1)
	}

	if config.Dependencies == nil {
		config.Dependencies = make(map[string]Dependency)
	}

	// Handle URLs and Git repos
	isGit := strings.HasPrefix(target, "http") || strings.HasPrefix(target, "git@") || strings.HasPrefix(target, "file://") || strings.HasSuffix(target, ".git")

	if isGit {
		if name == "" {
			// Infer name from URL/Path
			cleanTarget := strings.TrimSuffix(target, ".git")
			cleanTarget = strings.TrimSuffix(cleanTarget, "/")
			parts := strings.Split(cleanTarget, "/")
			name = parts[len(parts)-1]
			if name == "" {
				name = "lib"
			}
		}

		libDir := filepath.Join("libs", name)
		if _, err := os.Stat("libs"); os.IsNotExist(err) {
			os.Mkdir("libs", 0755)
		}

		fmt.Printf("Cloning library '%s' from %s...\n", name, target)
		// Handle file:// prefix for git clone
		cloneTarget := strings.TrimPrefix(target, "file://")
		cloneCmd := exec.Command("git", "clone", cloneTarget, libDir)
		cloneCmd.Stdout = os.Stdout
		cloneCmd.Stderr = os.Stderr
		if err := cloneCmd.Run(); err != nil {
			fmt.Printf("Error cloning library: %v\n", err)
			os.Exit(1)
		}

		// Try to read version from cloned lib
		libVersion := "1.0.0"
		libConfigPath := filepath.Join(libDir, "nora.yaml")
		if data, err := os.ReadFile(libConfigPath); err == nil {
			var libConfig ProjectConfig
			if err := yaml.Unmarshal(data, &libConfig); err == nil {
				libVersion = libConfig.Version
			}
		}

		if requestedVersion != "" && requestedVersion != libVersion {
			fmt.Printf("Warning: requested version '%s' does not match library version '%s'. Using detected version.\n", requestedVersion, libVersion)
		}

		config.Dependencies[name] = Dependency{Path: libDir, Version: libVersion}
	} else {
		// Handle local paths
		if name == "" {
			name = filepath.Base(target)
		}

		libVersion := "local"
		libConfigPath := filepath.Join(target, "nora.yaml")
		if data, err := os.ReadFile(libConfigPath); err == nil {
			var libConfig ProjectConfig
			if err := yaml.Unmarshal(data, &libConfig); err == nil {
				libVersion = libConfig.Version
			}
		}

		if requestedVersion != "" && requestedVersion != libVersion {
			fmt.Printf("Warning: requested version '%s' does not match local library version '%s'.\n", requestedVersion, libVersion)
		}

		config.Dependencies[name] = Dependency{Path: target, Version: libVersion}
		fmt.Printf("Added local dependency '%s' (v%s) -> %s\n", name, libVersion, target)
	}

	if err := config.Save(); err != nil {
		fmt.Printf("Error updating nora.yaml: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Library '%s' added successfully.\n", name)
}

func runLibRemove(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: Nora lib remove <name>")
		os.Exit(1)
	}

	name := args[0]
	config, err := LoadProjectConfig()
	if err != nil {
		fmt.Printf("Error: nora.yaml not found\n")
		os.Exit(1)
	}

	if _, ok := config.Dependencies[name]; !ok {
		fmt.Printf("Error: library '%s' not found in dependencies\n", name)
		os.Exit(1)
	}

	delete(config.Dependencies, name)
	if err := config.Save(); err != nil {
		fmt.Printf("Error updating nora.yaml: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Library '%s' removed from nora.yaml.\n", name)

	// Optional: delete from libs/ if it exists there
	libDir := filepath.Join("libs", name)
	if _, err := os.Stat(libDir); err == nil {
		fmt.Printf("Deleting local library files in %s...\n", libDir)
		os.RemoveAll(libDir)
	}
}

func runLibList() {
	config, err := LoadProjectConfig()
	if err != nil {
		fmt.Printf("Error: nora.yaml not found\n")
		os.Exit(1)
	}

	if len(config.Dependencies) == 0 {
		fmt.Println("No dependencies found.")
		return
	}

	fmt.Println("Project Dependencies:")
	fmt.Printf(" %-15s %-10s %s\n", "NAME", "VERSION", "PATH")
	fmt.Println(" " + strings.Repeat("-", 40))
	for name, dep := range config.Dependencies {
		fmt.Printf(" - %-15s %-10s %s\n", name, dep.Version, dep.Path)
	}
}

func printUsage() {
	fmt.Println("================================================================================")
	fmt.Println("  NORA COMPILER SYSTEM  -  v" + CompilerVersion)
	fmt.Println("================================================================================")
	fmt.Println("Usage: Nora <command> [arguments] [flags]")
	fmt.Println()
	fmt.Println("Core Commands:")
	fmt.Println("  init <name> [--lib] Initialize a premium Nora binary executable or library project")
	fmt.Println("  build [file/dir]    Compile Nora source files to a native binary or Wasm module")
	fmt.Println("                      (Use '--example <name>' to build library example files)")
	fmt.Println("  run   [file/dir]    Compile, link, and execute Nora source files instantly")
	fmt.Println("                      (Use '--example <name>' to run library example files)")
	fmt.Println("  fmt   [files...]    Enforce premium formatting style across source files")
	fmt.Println("  targets             Display all supported and active platform targets")
	fmt.Println("  check               Validate required toolchain dependencies in the system PATH")
	fmt.Println("  clean               Clear the local build artifact workspace")
	fmt.Println("  lib                 Manage external package dependencies (add, remove, list)")
	fmt.Println("  lsp                 Launch the Language Server Protocol (LSP) background daemon")
	fmt.Println("  doc <pkg>           Generate beautiful markdown documentation for a package")
	fmt.Println("  help                Display this system help menu")
	fmt.Println()
	fmt.Println("Premium Build & Run Options:")
	fmt.Println("  -o <name>           Specify output binary name (build only)")
	fmt.Println("  -p <plugins>        Comma-separated list of compiler/macro plugins (.wasm)")
	fmt.Println("  --target <t>        Cross-compile target (e.g. windows-amd64, wasm-wasi, linux-amd64)")
	fmt.Println("  --wasm              Shorthand for legacy WebAssembly target (--target wasm-unknown)")
	fmt.Println("  --wasi              Shorthand for standard WebAssembly WASI target (--target wasm-wasi)")
	fmt.Println("  --release, -r       Compile with aggressive optimization (O3 / O2)")
	fmt.Println("  --debug, -d         Compile with debug optimizations (O0 / Od) [default]")
	fmt.Println("  -g                  Embed full debug symbols and source mapping metadata")
	fmt.Println("  --wasm-experimental Enable next-gen Wasm stack-switching and async runtime")
	fmt.Println("  --example <name>    Specify a runnable library example from the examples/ directory")
	fmt.Println()
	fmt.Println("Decoupled Toolchain Overrides (CLI & nora.yaml):")
	fmt.Println("  --cc <compiler>     Override C compiler (e.g. clang, gcc, cl, tcc)")
	fmt.Println("  --cflags \"<flags>\"  Direct custom flags passed straight to the target C compiler")
	fmt.Println("                      (Multiple flags must be quoted, e.g. --cflags \"-O2 -Wall\")")
	fmt.Println()
	fmt.Println("Diagnostics & Debugging:")
	fmt.Println("  --debug-memory      Enable strict runtime memory leak tracking and allocator audit")
	fmt.Println("  --debug-semantic    Dump semantic analysis phase graphs and resolution logs")
	fmt.Println("  --debug-topology    Dump symbol resolution dependency topology tree")
	fmt.Println("  --fiber-report      Print high-resolution fiber lifecycle report at execution exit")
	fmt.Println()
	fmt.Println("Premium Formatting Options (fmt):")
	fmt.Println("  -w                  Write formatted result directly back to the source file")
	fmt.Println("                      (Without -w, formatted source is printed directly to stdout)")
	fmt.Println()
	fmt.Println("================================================================================")
}

func runCheck() {
	fmt.Println("================================================================================")
	fmt.Println("  NORA SYSTEM TOOLCHAIN VERIFICATION")
	fmt.Println("================================================================================")

	// 1. Check C Compiler Families
	fmt.Println("C Compiler Toolchains:")
	compilers := []struct {
		name string
		desc string
	}{
		{"clang", "LLVM/Clang Compiler (Standard Default)"},
		{"gcc", "GNU Compiler Collection (GCC)"},
		{"cl", "Microsoft Visual Studio Compiler (MSVC)"},
		{"tcc", "Tiny C Compiler (TCC)"},
	}

	foundCompilers := 0
	for _, c := range compilers {
		path, err := exec.LookPath(c.name)
		if err == nil {
			fmt.Printf("  [OK]   %-8s : %s\n", c.name, path)
			foundCompilers++
		} else {
			fmt.Printf("  [INFO] %-8s : Not found in PATH (%s)\n", c.name, c.desc)
		}
	}

	if foundCompilers == 0 {
		fmt.Println("\n  [FAIL] No C compiler toolchains found! You must install at least one C compiler")
		fmt.Println("         (e.g., clang, gcc, cl, or tcc) and add it to your system PATH.")
	} else {
		fmt.Printf("\n  [OK]   Found %d available C compiler toolchain(s)!\n", foundCompilers)
	}

	fmt.Println("\nWebAssembly (Wasm) & WASI Runtimes:")

	// Wasmtime check
	wasmtimePath, err := exec.LookPath("wasmtime")
	if err == nil {
		fmt.Printf("  [OK]   wasmtime : %s\n", wasmtimePath)
	} else {
		fmt.Println("  [WARN] wasmtime : Not found in PATH. Required to execute WebAssembly WASI modules.")
	}

	// WASI SDK check
	if sdkPath := os.Getenv("WASI_SDK_PATH"); sdkPath != "" {
		fmt.Printf("  [OK]   WASI SDK : %s\n", sdkPath)
	} else {
		fmt.Println("  [WARN] WASI SDK : WASI_SDK_PATH environment variable is not set.")
		fmt.Println("                    Required to cross-compile Wasm/WASI modules using clang.")
	}

	// Legacy Emscripten check
	emccPath, err := exec.LookPath("emcc")
	if err == nil {
		fmt.Printf("  [OK]   emcc     : %s\n", emccPath)
	} else {
		fmt.Println("  [INFO] emcc     : Not found in PATH. Only required for legacy Wasm builds.")
	}

	fmt.Println("\nSystem Core:")
	if info, err := os.Stat(StdPath); err == nil && info.IsDir() {
		absStd, _ := filepath.Abs(StdPath)
		fmt.Printf("  [OK]   Standard Library: %s\n", absStd)
	} else {
		fmt.Printf("  [FAIL] Standard Library path not found or invalid: %s\n", StdPath)
	}
	fmt.Println("================================================================================")
}

func runWasmStandalone(wasmPath string) error {
	ctx := context.Background()
	runtimeConfig := wazero.NewRuntimeConfig().
		WithCoreFeatures(api.CoreFeaturesV2)
	r := wazero.NewRuntimeWithConfig(ctx, runtimeConfig)
	defer r.Close(ctx)

	wasi_snapshot_preview1.MustInstantiate(ctx, r)

	// Provide 'env' module for Emscripten/Asyncify
	_, err := r.NewHostModuleBuilder("env").
		NewFunctionBuilder().WithFunc(func(ctx context.Context, p int32) {}).Export("asyncify_start_unwind").
		NewFunctionBuilder().WithFunc(func(ctx context.Context) {}).Export("asyncify_stop_unwind").
		NewFunctionBuilder().WithFunc(func(ctx context.Context, p int32) {}).Export("asyncify_start_rewind").
		NewFunctionBuilder().WithFunc(func(ctx context.Context) {}).Export("asyncify_stop_rewind").
		NewFunctionBuilder().WithFunc(func(ctx context.Context) int32 { return 0 }).Export("asyncify_get_state").
		Instantiate(ctx)
	if err != nil {
		return err
	}

	// Provide 'wasm_runtime' module for Native Stack Switching stubs
	_, err = r.NewHostModuleBuilder("wasm_runtime").
		NewFunctionBuilder().WithFunc(func(ctx context.Context, fn int32, arg int32) int32 { return 0 }).Export("cont_new").
		NewFunctionBuilder().WithFunc(func(ctx context.Context, cont int32) {}).Export("cont_resume").
		NewFunctionBuilder().WithFunc(func(ctx context.Context) {}).Export("cont_suspend").
		Instantiate(ctx)
	if err != nil {
		return err
	}

	wasmBytes, err := os.ReadFile(wasmPath)
	if err != nil {
		return err
	}

	config := wazero.NewModuleConfig().
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithStdin(os.Stdin).
		WithArgs(os.Args[1:]...)

	_, err = r.InstantiateWithConfig(ctx, wasmBytes, config)
	return err
}

func runDoc(args []string) {
	docFlags := flag.NewFlagSet("doc", flag.ExitOnError)
	serveFlag := docFlags.Bool("serve", false, "Start an HTTP server to view docs locally")
	portFlag := docFlags.String("port", "8080", "Port for the HTTP server (default: 8080)")
	docFlags.Parse(args)

	if docFlags.NArg() < 1 {
		fmt.Println("Usage: Nora doc [--serve] [--port=8080] <package-path>")
		os.Exit(1)
	}

	path := docFlags.Arg(0)
	files, err := docgen.ParsePackage(path)
	if err != nil {
		fmt.Printf("Failed to parse package: %v\n", err)
		os.Exit(1)
	}

	gen := docgen.NewDocGenerator(files)

	if *serveFlag {
		docgen.Serve(gen, *portFlag)
	} else {
		md := gen.GenerateMarkdown()

		outPath := "README.md"
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			outPath = filepath.Join(path, "README.md")
		} else {
			dir := filepath.Dir(path)
			outPath = filepath.Join(dir, "README.md")
		}

		if err := os.WriteFile(outPath, []byte(md), 0644); err != nil {
			fmt.Printf("Failed to write documentation: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Documentation generated at %s\n", outPath)
	}
}
func runFmt(args []string) {
	fmtFlags := flag.NewFlagSet("fmt", flag.ExitOnError)
	writeFlag := fmtFlags.Bool("w", false, "Write result to (source) file instead of stdout")
	fmtFlags.Parse(args)

	var files []string
	if fmtFlags.NArg() < 1 {
		// Format all .nr files in current directory and src/
		filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(path, ".nr") {
				files = append(files, path)
			}
			return nil
		})
	} else {
		files = fmtFlags.Args()
	}

	config, err := format.LoadConfig("nora-fmt.yaml")
	if err != nil {
		fmt.Printf("Warning: failed to load nora-fmt.yaml: %v. Using defaults.\n", err)
		config = format.DefaultConfig()
	}

	formatter := format.New(config)

	for _, file := range files {
		input, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", file, err)
			continue
		}

		l := lexer.New(string(input), file)
		p := parser.New(l)
		p.AllowNoPackage = false
		p.PreserveParentheses = true
		p.DisableMacros = true
		parsedFile := p.Parse(file)

		if p.Diagnostics.HasErrors() {
			fmt.Printf("Error: could not parse %s. Skipping.\n", file)
			diag.Report(p.Diagnostics)
			continue
		}

		formatted := formatter.Format(parsedFile)

		if *writeFlag {
			if string(input) != formatted {
				if err := os.WriteFile(file, []byte(formatted), 0644); err != nil {
					fmt.Printf("Error writing %s: %v\n", file, err)
				} else {
					fmt.Printf("Formatted %s\n", file)
				}
			}
		} else {
			fmt.Println(formatted)
		}
	}
}

type PackageCacheEntry struct {
	Hash       string `json:"hash"`
	ObjectPath string `json:"object_path"`
}

type BuildCacheCatalog struct {
	Compiler          string                       `json:"compiler"`
	Target            string                       `json:"target"`
	Release           bool                         `json:"release"`
	DebugMemory       bool                         `json:"debug_memory"`
	DebugFiber        bool                         `json:"debug_fiber"`
	CFlags            string                       `json:"cflags"`
	GlobalsHash       string                       `json:"globals_hash"`
	GlobalsObjectPath string                       `json:"globals_object_path"`
	Packages          map[string]PackageCacheEntry `json:"packages"`
}

func loadBuildCache(buildDir string) (*BuildCacheCatalog, error) {
	cachePath := filepath.Join(buildDir, ".nora_build_cache.json")
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return &BuildCacheCatalog{Packages: make(map[string]PackageCacheEntry)}, nil
	}
	var catalog BuildCacheCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return &BuildCacheCatalog{Packages: make(map[string]PackageCacheEntry)}, nil
	}
	if catalog.Packages == nil {
		catalog.Packages = make(map[string]PackageCacheEntry)
	}
	return &catalog, nil
}

func (c *BuildCacheCatalog) Save(buildDir string) error {
	cachePath := filepath.Join(buildDir, ".nora_build_cache.json")
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath, data, 0644)
}

func compileCToObject(compiler string, srcPath string, objPath string, isMSVC bool, activeConfig CompilerConfig, opts BuildOptions) error {
	absSrcPath, err := filepath.Abs(srcPath)
	if err != nil {
		absSrcPath = filepath.Clean(srcPath)
	}

	var objArgs []string
	if isMSVC {
		objArgs = append(objArgs, "/c", absSrcPath)
		objOutFlag := "/Fo:"
		objArgs = append(objArgs, objOutFlag+objPath)
	} else {
		objArgs = append(objArgs, "-c", absSrcPath)
		if strings.HasSuffix(activeConfig.OutFlag, ":") {
			objArgs = append(objArgs, activeConfig.OutFlag+objPath)
		} else {
			objArgs = append(objArgs, activeConfig.OutFlag, objPath)
		}
	}

	// Optimization and Debug symbols
	if opts.Release {
		if activeConfig.OptRelease != "" {
			objArgs = append(objArgs, activeConfig.OptRelease)
		}
	} else {
		if activeConfig.OptDebug != "" {
			objArgs = append(objArgs, activeConfig.OptDebug)
		}
	}
	if opts.G && activeConfig.DebugSymbols != "" {
		objArgs = append(objArgs, activeConfig.DebugSymbols)
	}

	// Macro defines
	definePrefix := activeConfig.DefineFlag
	if opts.DebugFiber {
		objArgs = append(objArgs, definePrefix+"NR_DEBUG_FIBER")
	}
	if opts.DebugMemory {
		objArgs = append(objArgs, definePrefix+"NR_DEBUG_MEM=1")
	}

	// Include directories
	buildMode := "debug"
	if opts.Release {
		buildMode = "release"
	}
	buildDir := filepath.Join("build", buildMode)
	objArgs = append(objArgs, activeConfig.IncFlag+buildDir)

	for _, dir := range opts.Native.IncludeDirs {
		objArgs = append(objArgs, activeConfig.IncFlag+dir)
	}

	// Custom compiler flags
	objArgs = append(objArgs, activeConfig.CFlags...)

	// Wasm experimental or specific sysroots
	if opts.Target.Wasm && (strings.Contains(strings.ToLower(activeConfig.Compiler), "clang") || strings.Contains(strings.ToLower(activeConfig.Compiler), "wasi-sdk")) {
		if sdkPath := os.Getenv("WASI_SDK_PATH"); sdkPath != "" {
			objArgs = append(objArgs, "--sysroot="+filepath.Join(sdkPath, "share", "wasi-sysroot"))
			if opts.WasmExperimental {
				objArgs = append(objArgs, "-DNR_USE_FIBERS")
			}
		}
	}

	if opts.Verbose {
		fmt.Printf("[Nora] Compiling C file: %s -> %s\n", srcPath, objPath)
		fmt.Printf("[Nora] Compile command: %s %s\n", activeConfig.Compiler, strings.Join(objArgs, " "))
	}

	cmd := exec.Command(activeConfig.Compiler, objArgs...)
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf
	err = cmd.Run()
	if opts.Verbose {
		os.Stdout.Write(stdoutBuf.Bytes())
		os.Stderr.Write(stderrBuf.Bytes())
	}
	if err != nil {
		if !opts.Verbose {
			os.Stdout.Write(stdoutBuf.Bytes())
			os.Stderr.Write(stderrBuf.Bytes())
		}
	}
	return err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = out.Sync()
	if err != nil {
		return err
	}

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return err
	}
	return os.Chtimes(dst, si.ModTime(), si.ModTime())
}

func copyDynamicLibraries(buildDir string, opts BuildOptions) error {
	if len(opts.Native.DynamicLibs) == 0 {
		return nil
	}

	// Determine OS target
	targetOS := opts.Target.OS
	if targetOS == "" {
		targetOS = runtime.GOOS
	}

	for _, lib := range opts.Native.DynamicLibs {
		libName := filepath.Base(lib)
		if libName == "" {
			continue
		}

		var candidates []string
		switch targetOS {
		case "windows":
			candidates = []string{
				libName,
				libName + ".dll",
				"lib" + libName + ".dll",
			}
			if strings.HasSuffix(strings.ToLower(libName), ".dll") {
				base := strings.TrimSuffix(libName, ".dll")
				candidates = append(candidates, base, "lib"+base+".dll")
			} else if strings.HasSuffix(strings.ToLower(libName), "dll") {
				base := libName[:len(libName)-3]
				candidates = append(candidates, base+".dll", "lib"+base+".dll")
			}
		case "linux":
			candidates = []string{
				libName,
				"lib" + libName + ".so",
				libName + ".so",
			}
			if strings.HasSuffix(strings.ToLower(libName), ".so") {
				base := strings.TrimSuffix(libName, ".so")
				candidates = append(candidates, base, "lib"+base+".so")
			}
		case "darwin":
			candidates = []string{
				libName,
				"lib" + libName + ".dylib",
				libName + ".dylib",
			}
			if strings.HasSuffix(strings.ToLower(libName), ".dylib") {
				base := strings.TrimSuffix(libName, ".dylib")
				candidates = append(candidates, base, "lib"+base+".dylib")
			}
		default:
			candidates = []string{
				libName,
				libName + ".dll",
				"lib" + libName + ".dll",
				"lib" + libName + ".so",
				libName + ".so",
				"lib" + libName + ".dylib",
				libName + ".dylib",
			}
		}

		found := false
		var srcPath string
		var matchedName string

		for _, dir := range opts.Native.LibDirs {
			for _, candidate := range candidates {
				path := filepath.Join(dir, candidate)
				if info, err := os.Stat(path); err == nil && !info.IsDir() {
					srcPath = path
					matchedName = candidate
					found = true
					break
				}
			}
			if found {
				break
			}
		}

		if !found {
			if opts.Verbose {
				fmt.Printf("[Nora] Dynamic library '%s' not found in lib_dirs. Assuming system library.\n", lib)
			}
			continue
		}

		destPath := filepath.Join(buildDir, matchedName)

		// Optimize: skip if destination exists and has same mod time and size
		srcInfo, err := os.Stat(srcPath)
		if err != nil {
			continue
		}
		destInfo, err := os.Stat(destPath)
		if err == nil && destInfo.Size() == srcInfo.Size() && destInfo.ModTime().Equal(srcInfo.ModTime()) {
			if opts.Verbose {
				fmt.Printf("[Nora] Dynamic library '%s' is up-to-date in build folder: %s\n", matchedName, destPath)
			}
			continue
		}

		if opts.Verbose {
			fmt.Printf("[Nora] Copying dynamic library: %s -> %s\n", srcPath, destPath)
		} else {
			fmt.Printf("  [Nora] Copying dynamic library: %s...\n", matchedName)
		}

		if err := copyFile(srcPath, destPath); err != nil {
			return fmt.Errorf("failed to copy dynamic library %s to %s: %v", srcPath, destPath, err)
		}
	}
	return nil
}
