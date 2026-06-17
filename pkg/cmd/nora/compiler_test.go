package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/DwiYI/Project-Nora/pkg/codegen"
	"github.com/DwiYI/Project-Nora/pkg/lexer"
	"github.com/DwiYI/Project-Nora/pkg/parser"
	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/plugin"
	"github.com/DwiYI/Project-Nora/pkg/semantic"
	"github.com/DwiYI/Project-Nora/pkg/topology"
)

func TestCompilerWithTestFolder(t *testing.T) {
	// Change working directory to project root so FileLoader can resolve paths correctly
	if _, err := os.Stat("go.mod"); err != nil {
		err := os.Chdir("../../../")
		if err != nil {
			t.Fatalf("Failed to change working directory: %v", err)
		}
	}
	initStdPath()
	initCorePath()

	// Initialize PluginManager once for all tests to avoid parsing WASM 50+ times
	pm := plugin.NewPluginManager()
	defer pm.Close()
	pm.RegisterBuiltinMacros()

	// Load standard IO and serialize macro plugins if they exist
	ioWasmPath := filepath.Join(StdPath, "io/io_macro.wasm")
	if _, err := os.Stat(ioWasmPath); err == nil {
		if err := pm.LoadPlugin("io_macro", ioWasmPath); err != nil {
			t.Logf("Warning: failed to load %s: %v", ioWasmPath, err)
		}
	}
	serializeWasmPath := filepath.Join(StdPath, "json/serialize.wasm")
	if _, err := os.Stat(serializeWasmPath); err == nil {
		if err := pm.LoadPlugin("serialize", serializeWasmPath); err != nil {
			t.Logf("Warning: failed to load %s: %v", serializeWasmPath, err)
		}
	}

	testDirs := []string{"pkg/cmd/test"}
	for _, testDir := range testDirs {
		err := filepath.Walk(testDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				t.Logf("Warning: walk encountered error on path %s: %v", path, err)
				return nil
			}
			if info.IsDir() || !strings.HasSuffix(info.Name(), ".nr") {
				return nil
			}

			t.Run(path, func(t *testing.T) {
				inputFile := path
				expectFail := strings.HasPrefix(info.Name(), "fail_") || strings.Contains(info.Name(), "violation")

				// 1. Read Source
				input, err := ioutil.ReadFile(inputFile)
				if err != nil {
					t.Fatalf("Error reading file '%s': %v", inputFile, err)
				}

				// 2. Parse
				l := lexer.New(string(input), inputFile)
				p := parser.New(l)
				parsedFile := p.Parse(inputFile)
				if len(p.Errors()) > 0 {
					if expectFail {
						fmt.Printf("Successfully caught expected parser error in: %s\n", path)
						return // Test passes, stop here
					}
					for _, msg := range p.Errors() {
						t.Logf("Parser Error: %s", msg)
					}
					t.Fatalf("Parser encountered %d errors", len(p.Errors()))
				}
				prog := &ast.Program{Files: []*ast.File{parsedFile}}

				for _, file := range prog.Files {
					if err := pm.ProcessMacroForFile(file); err != nil {
						t.Fatalf("Macro processing failed: %v", err)
					}
				}

				// 3. Analyze
				analyzer := semantic.NewAnalyzer()
				analyzer.AllowUnsafe = true // Allow unsafe for integration tests that compile stdlib
				parsedFiles := make(map[string]*ast.File)
				parsedFiles[filepath.Clean(inputFile)] = parsedFile

				// Auto-load core/prelude.nr if it exists
				preludePath := filepath.Join(CorePath, "prelude.nr")
				if _, err := os.Stat(preludePath); err == nil {
					preludeInput, err := ioutil.ReadFile(preludePath)
					if err == nil {
						preludeL := lexer.New(string(preludeInput), preludePath)
						preludeP := parser.New(preludeL)
						preludeFile := preludeP.Parse(preludePath)
						if !preludeP.Diagnostics.HasErrors() {
							prog.Files = append(prog.Files, preludeFile)
							parsedFiles[filepath.Clean(preludePath)] = preludeFile
						}
					}
				}
				loader := &FileLoader{
					Cache:       make(map[string]*semantic.Scope),
					ParsedFiles: parsedFiles,
					Program:     prog,
					Analyzer:    analyzer,
				}
				analyzer.Loader = loader
				loader.loadManifest(CorePath)
				loader.loadManifest(StdPath)

				analyzer.Analyze(prog)
				if analyzer.Diagnostics.HasErrors() {
					if expectFail {
						fmt.Printf("Successfully caught expected semantic error in: %s\n", path)
						return // Test passes, stop here
					}
					for _, d := range analyzer.Diagnostics.Diagnostics {
						t.Logf("Semantic Error: %s", d.String())
					}
					t.Fatalf("Semantic Analyzer encountered %d errors", len(analyzer.Diagnostics.Diagnostics))
				}

				// 4. Topology Solve
				solver := topology.NewSolver(&analyzer.SemanticInfo)
				solver.Solve(prog)
				if solver.Diagnostics.HasErrors() {
					if expectFail {
						fmt.Printf("Successfully caught expected topology error in: %s\n", path)
						return // Test passes, stop here
					}
					for _, d := range solver.Diagnostics.Diagnostics {
						t.Logf("Topology Error: %s", d.String())
					}
					t.Fatalf("Topology Solver encountered %d errors", len(solver.Diagnostics.Diagnostics))
				}

				// 5. Generate Code
				gen := codegen.NewGenerator(prog, &analyzer.SemanticInfo, solver, pm, analyzer.Loader.(*FileLoader).CollectedNative.Headers)

				// Enable memory tracking for ALL tests to catch leaks early
				gen.DebugMemory = true
				isExpectedLeak := strings.Contains(path, "leak_detection") || strings.Contains(path, "real_leak") || strings.Contains(path, "source_leak")

				bodyCode, err := gen.Generate()
				if err != nil {
					if expectFail {
						fmt.Printf("Successfully caught expected codegen error in: %s\n", path)
						return // Test passes
					}
					t.Fatalf("Codegen Error: %v", err)
				}

				// 6. Write to C file (in temp dir)
				tempDir := t.TempDir()
				cFileName := strings.TrimSuffix(info.Name(), ".nr") + ".c"
				cFilePath := filepath.Join(tempDir, cFileName)

				err = ioutil.WriteFile(cFilePath, []byte(bodyCode), 0644)
				if err != nil {
					t.Fatalf("Error writing C file: %v", err)
				}

				// 7. Compile with GCC (Only if main exists)
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
					fmt.Printf("Skipping clang for library: %s\n", path)
					return
				}

				exeFile := filepath.Join(tempDir, strings.TrimSuffix(info.Name(), ".nr"))
				if runtime.GOOS == "windows" {
					exeFile += ".exe"
				}
				args := []string{cFilePath, "-o", exeFile, "-Wno-pointer-sign", "-Wno-deprecated-declarations", "-Wno-parentheses-equality", "-Wno-unused-value"}
				for _, dir := range loader.CollectedNative.IncludeDirs {
					args = append(args, "-I"+dir)
				}
				for _, dir := range loader.CollectedNative.LibDirs {
					args = append(args, "-L"+dir)
				}
				for _, lib := range loader.CollectedNative.DynamicLibs {
					args = append(args, "-l"+lib)
				}
				for _, lib := range loader.CollectedNative.StaticLibs {
					args = append(args, lib)
				}
				for _, src := range loader.CollectedNative.SourceFiles {
					args = append(args, src)
				}
				
				// [NEW] Dynamically include any .c files found alongside the test file
				testDir := filepath.Dir(inputFile)
				localCFiles, _ := filepath.Glob(filepath.Join(testDir, "*.c"))
				for _, cFile := range localCFiles {
					args = append(args, cFile)
				}
				
				if gen.DebugMemory {
					args = append(args, "-DNR_DEBUG_MEM=1")
				}

				if runtime.GOOS == "windows" {
					args = append(args, "-ldbghelp", "-lws2_32")
				}
				if runtime.GOOS == "linux" {
					args = append(args, "-pthread", "-lm")
				}
				cmd := exec.Command("clang", args...)
				output, err := cmd.CombinedOutput()

				if err != nil {
					if expectFail {
						fmt.Printf("Successfully caught expected compilation error in: %s\n", path)
						return
					} else {
						t.Fatalf("clang Compilation Error: %v\nOutput: %s", err, string(output))
					}
				}

				if !expectFail && strings.Contains(string(output), "warning:") {
					t.Fatalf("clang Compilation Warning (treated as Error):\nOutput: %s", string(output))
				}

				if expectFail {
					t.Fatalf("Expected compilation to fail for %s, but it succeeded", path)
				}

				fmt.Printf("Successfully compiled: %s\n", path)

				// 8. Run the executable
				// Skip tests that are known to be interactive or don't have input provided yet
				skipRun := false
				interactivePrefixes := []string{"scan", "test_scanf", "4A", "71A", "112A", "158A", "231A", "236A", "263A", "282A", "339A", "50A"}
				for _, prefix := range interactivePrefixes {
					if strings.HasPrefix(info.Name(), prefix) {
						skipRun = true
						break
					}
				}

				if skipRun {
					fmt.Printf("Skipping execution for interactive/Codeforces test: %s\n", path)
					return
				}

				fmt.Printf("Running executable: %s\n", exeFile)
				runCmd := exec.Command(exeFile)
				var runOutput strings.Builder
				runCmd.Stdout = &runOutput
				runCmd.Stderr = &runOutput

				// Set a timeout for execution
				done := make(chan error, 1)
				go func() {
					done <- runCmd.Run()
				}()

				select {
				case <-time.After(60 * time.Second):
					if runCmd.Process != nil {
						runCmd.Process.Kill()
					}
					t.Fatalf("Execution timed out for %s", path)
				case err := <-done:
					if err != nil {
						isCrash := false
						if exitErr, ok := err.(*exec.ExitError); ok {
							exitCode := exitErr.ExitCode()
							if runtime.GOOS == "windows" {
								uCode := uint32(exitCode)
								if uCode == 0xC0000005 || uCode == 0xC000001D || uCode == 0xC0000094 || uCode == 0xC00000FD {
									isCrash = true
								}
							} else {
								if status, ok := exitErr.Sys().(interface{ Signaled() bool }); ok && status.Signaled() {
									isCrash = true
								}
							}
						}

						if isCrash {
							t.Fatalf("SEGMENTATION FAULT / CRASH for %s: %v\nOutput: %s", path, err, runOutput.String())
						}

						// If it's a panic or deadlock test, we expect a non-zero exit code
						if strings.Contains(info.Name(), "panic") || strings.Contains(info.Name(), "deadlock") {
							fmt.Printf("Successfully caught expected panic/deadlock in: %s\n", path)
						} else {
							t.Fatalf("Runtime Error for %s: %v\nOutput: %s", path, err, runOutput.String())
						}
					}
				}

				// Verify memory leaks
				outputStr := runOutput.String()
				hasLeakReport := strings.Contains(outputStr, "Nora MEMORY LEAK REPORT")

				if isExpectedLeak {
					if !hasLeakReport {
						t.Errorf("Leak test %s failed: Expected leak report but none found\nOutput: %s", path, outputStr)
					} else {
						fmt.Printf("Successfully detected expected leak in: %s\n", path)
					}
				} else {
					if hasLeakReport {
						t.Errorf("MEMORY LEAK DETECTED in %s:\n%s", path, outputStr)
					}
				}

				fmt.Printf("Successfully ran: %s\n", path)
			})
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to walk test directory %s: %v", testDir, err)
		}
	}
}
