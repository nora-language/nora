package codegen

import (
	"fmt"
	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/types"
	"sort"
)

var cStdlibFunctions = map[string]bool{
	"printf": true, "scanf": true, "snprintf": true, "sprintf": true,
	"strlen": true, "strcmp": true, "strncmp": true, "strstr": true, "strcpy": true, "strncpy": true,
	"toupper": true, "tolower": true, "isspace": true, "isdigit": true, "isalpha": true,
	"getchar": true, "putchar": true, "fopen": true, "fclose": true, "fputs": true, "fputc": true,
	"fgetc": true, "fgets": true, "fread": true, "fwrite": true, "fflush": true, "feof": true,
	"ferror": true, "rewind": true, "fseek": true, "ftell": true, "memset": true, "memcpy": true, "memmove": true, "memcmp": true,
	"malloc": true, "free": true, "realloc": true, "calloc": true, "exit": true, "abort": true, "Sleep": true,
}

func (g *Generator) emitExternPrototypes(file *ast.File) {
	for _, stmt := range file.Statements {
		if ext, ok := stmt.(*ast.ExternStatement); ok {
			fn := ext.Function
			if fn == nil {
				continue
			}
			if cStdlibFunctions[fn.Name.Value] {
				continue
			}

			g.emitLine(fn)
			retType := g.cType(g.SemanticInfo.Types[fn.ReturnType])
			params := ""
			for i, p := range fn.Parameters {
				if i > 0 {
					params += ", "
				}
				if p.IsVariadic {
					params += "..."
					continue
				}
				lease := types.LeaseRead
				params += g.cParamType(g.SemanticInfo.Types[p.Type], lease)
				if p.Name != nil {
					params += " " + p.Name.Value
				}
			}
			if params == "" {
				params = "void"
			}
			g.emit("extern %s %s(%s);", retType, fn.Name.Value, params)
		}
	}
	g.currentFile = ""
	g.currentLine = -1
}

func (g *Generator) emitPrototypes() {
	g.emit("// --- PROTOTYPES ---")
	for _, sym := range g.Functions {
		if cStdlibFunctions[sym.Name] && sym.Name == g.mangleName(sym) {
			continue
		}
		if sym.DefNode == nil {
			continue
		}
		fn, ok := sym.DefNode.(*ast.FunctionStatement)
		if !ok {
			continue
		}

		ft := sym.Type.(*types.FunctionType)
		retType := g.cType(ft.Return)
		name := g.mangleName(sym)

		params := ""
		if !fn.IsExtern && !fn.IsExport {
			params = "void* _env_ptr"
		}

		if fn.Receiver != nil {
			t := g.cParamType(ft.Receiver, ft.ReceiverLease)
			if params != "" {
				params += ", "
			}
			params += t + " " + fn.Receiver.Name.Value
		}
		for i, p := range ft.Params {
			if params != "" {
				params += ", "
			}
			lease := types.LeaseRead
			if i < len(ft.ParamLeases) {
				lease = ft.ParamLeases[i]
			}
			t := g.cParamType(p, lease)
			params += t
			if i < len(fn.Parameters) && fn.Parameters[i].Name != nil {
				params += " " + fn.Parameters[i].Name.Value
			}
		}
		if params == "" {
			params = "void"
		}
		if ft.IsVariadic {
			params += ", ..."
		}
		g.emit("%s %s(%s);", retType, name, params)
	}

	if g.hirProg != nil {
		g.emit("")
		g.emit("// --- HIR LAMBDA PROTOTYPES ---")
		for _, hf := range g.hirProg.Functions {
			if hf.LambdaExpr != nil {
				ft := hf.FuncSymbol.Type.(*types.FunctionType)
				retType := g.cType(ft.Return)
				name := hf.Name

				params := "void* _env_ptr"
				var lambdaParams []*ast.Parameter
				if hf.LambdaExpr != nil {
					lambdaParams = hf.LambdaExpr.Parameters
				}
				for i, p := range ft.Params {
					params += ", "
					lease := types.LeaseRead
					if i < len(ft.ParamLeases) {
						lease = ft.ParamLeases[i]
					}
					t := g.cParamType(p, lease)
					params += t
					if i < len(lambdaParams) && lambdaParams[i].Name != nil {
						params += " " + lambdaParams[i].Name.Value
					} else {
						if i < len(hf.Params) {
							params += " " + hf.Params[i]
						} else {
							params += fmt.Sprintf(" _param_%d", i)
						}
					}
				}
				g.emit("%s %s(%s);", retType, name, params)
			}
		}
	}
}

func (g *Generator) emitForwardDeclarations() {
	g.emit("// --- TYPE FORWARD DECLARATIONS ---")
	var structNames []string
	for name := range g.Structs {
		structNames = append(structNames, name)
	}
	sort.Strings(structNames)
	for _, name := range structNames {
		g.emit("typedef struct %s %s;", name, name)
	}

	var sumNames []string
	for name := range g.SumTypes {
		sumNames = append(sumNames, name)
	}
	sort.Strings(sumNames)
	for _, name := range sumNames {
		g.emit("typedef struct %s %s;", name, name)
	}
	g.emit("")
}

func (g *Generator) emitProtocolDefs() {
	if _, exists := g.Protocols["any"]; !exists {
		g.Protocols["any"] = &types.ProtocolType{ProtocolName: "any"}
	}

	g.emit("// --- PROTOCOL DEFINITIONS ---")
	var keys []string
	for name := range g.Protocols {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	for _, name := range keys {
		g.emit("typedef struct { void* data; void** vtable; } %s;", name)
	}
}

func (g *Generator) emitGlobalDecls() {
	g.emit("// --- GLOBAL DECLARATIONS ---")
	for _, sym := range g.Globals {
		g.emit("extern %s %s;", g.cType(sym.Type), g.mangleName(sym))
	}
}

func (g *Generator) emitGlobalDefs() {
	g.emit("// --- GLOBAL DEFINITIONS ---")
	for _, sym := range g.Globals {
		g.emit("%s %s;", g.cType(sym.Type), g.mangleName(sym))
	}
}

func (g *Generator) emitVariantConstructors() {
	g.emit("// --- VARIANT CONSTRUCTORS ---")
	var sumNames []string
	for name := range g.SumTypes {
		sumNames = append(sumNames, name)
	}
	for _, name := range sumNames {
		st := g.SumTypes[name]
		vNames := g.sortedVariantNames(st)
		for _, vName := range vNames {
			variant := st.Variants[vName]
			if len(variant.Fields) == 0 {
				g.emit("%s %s_%s_make() {", name, name, vName)
				g.emit("    %s res;", name)
				g.emit("    res.tag = %s_TAG_%s;", name, vName)
				g.emit("    return res;")
				g.emit("}")
			} else if len(variant.Fields) == 1 {
				var fType types.NRType
				for _, t := range variant.Fields {
					fType = t
					break
				}
				g.emit("%s %s_%s_make(%s val) {", name, name, vName, g.cType(fType))
				g.emit("    %s res;", name)
				g.emit("    res.tag = %s_TAG_%s;", name, vName)
				g.emit("    res.data.%s = val;", vName)
				g.emit("    return res;")
				g.emit("}")
			} else {
				g.emit("// TODO: multi-field constructor for %s_%s", name, vName)
			}
			g.emit("")
		}
	}
}

func (g *Generator) emitCombinedTypeDefs() {
	g.emit("// --- COMPINED TYPE DEFINITIONS ---")

	// Collect all names and build dependency map
	deps := make(map[string][]string)

	// Check struct dependencies
	for name, st := range g.Structs {
		for _, fType := range st.Fields {
			ut := types.UnwrapLease(fType)
			if ut != nil && (ut.GetKind() == types.KindStruct || ut.GetKind() == types.KindSum) {
				if !g.isPointerTypeInC(fType) {
					depName := g.cType(fType)
					deps[name] = append(deps[name], depName)
				}
			}
		}
	}

	// Check sum type dependencies
	for name, st := range g.SumTypes {
		for _, variant := range st.Variants {
			for _, fType := range variant.Fields {
				ut := types.UnwrapLease(fType)
				if ut != nil && (ut.GetKind() == types.KindStruct || ut.GetKind() == types.KindSum) {
					if !g.isPointerTypeInC(fType) {
						depName := g.cType(fType)
						deps[name] = append(deps[name], depName)
					}
				}
			}
		}
	}

	// Topological sort using DFS
	visited := make(map[string]bool)
	temp := make(map[string]bool)
	var order []string

	var visit func(string)
	visit = func(n string) {
		if temp[n] {
			return
		}
		if visited[n] {
			return
		}
		temp[n] = true
		for _, dep := range deps[n] {
			visit(dep)
		}
		temp[n] = false
		visited[n] = true
		order = append(order, n)
	}

	var keys []string
	for name := range g.Structs {
		keys = append(keys, name)
	}
	for name := range g.SumTypes {
		keys = append(keys, name)
	}
	sort.Strings(keys)

	for _, k := range keys {
		visit(k)
	}

	// Emit full definitions in top-sorted order
	for _, name := range order {
		if st, ok := g.Structs[name]; ok {
			g.emit("struct %s {", name)
			fieldNames := st.FieldNames
			for _, fName := range fieldNames {
				fType := st.Fields[fName]
				g.emit("    %s %s;", g.cType(fType), fName)
			}
			g.emit("};")
			g.emit("")
		} else if st, ok := g.SumTypes[name]; ok {
			g.emit("struct %s {", name)
			g.emit("    int tag;")
			g.emit("    union {")
			vNames := g.sortedVariantNames(st)
			for _, vName := range vNames {
				variant := st.Variants[vName]
				if len(variant.Fields) > 0 {
					if len(variant.Fields) == 1 {
						for _, fType := range variant.Fields {
							g.emit("        %s %s;", g.cType(fType), vName)
							break
						}
					} else {
						g.emit("        struct {")
						fNames := variant.FieldNames
						for _, fName := range fNames {
							g.emit("            %s %s;", g.cType(variant.Fields[fName]), fName)
						}
						g.emit("        } %s;", vName)
					}
				}
			}
			g.emit("    } data;")
			g.emit("};")
			g.emit("")

			// Emit TAG Macros
			for i, vName := range vNames {
				g.emit("#define %s_TAG_%s %d", name, vName, i)
			}
			g.emit("")

			// Emit Constructors
			for _, vName := range vNames {
				variant := st.Variants[vName]
				params := ""
				fNames := variant.FieldNames
				for j, fName := range fNames {
					if j > 0 {
						params += ", "
					}
					params += fmt.Sprintf("%s %s", g.cType(variant.Fields[fName]), fName)
				}
				if params == "" {
					params = "void"
				}

				g.emit("static inline %s %s_%s_make(%s) {", name, name, vName, params)
				g.emit("    %s _res;", name)
				g.emit("    memset(&_res, 0, sizeof(%s));", name)
				g.emit("    _res.tag = %s_TAG_%s;", name, vName)
				if len(variant.Fields) > 0 {
					if len(variant.Fields) == 1 {
						for fName := range variant.Fields {
							g.emit("    _res.data.%s = %s;", vName, fName)
						}
					} else {
						for _, fName := range fNames {
							g.emit("    _res.data.%s.%s = %s;", vName, fName, fName)
						}
					}
				}
				g.emit("    return _res;")
				g.emit("}")
			}
			g.emit("")
		}
	}
}
