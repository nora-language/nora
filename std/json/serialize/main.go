package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"unsafe"

	"github.com/DwiYI/Project-Nora/pkg/lexer"
	"github.com/DwiYI/Project-Nora/pkg/parser"
	"github.com/DwiYI/Project-Nora/pkg/parser/ast"
	"github.com/DwiYI/Project-Nora/pkg/plugin/api"
)

func main() {
	select {}
}

var allocs [][]byte

//go:wasmexport plugin_alloc
func plugin_alloc(size int32) int32 {
	buf := make([]byte, size)
	allocs = append(allocs, buf)
	return int32(uintptr(unsafe.Pointer(&buf[0])))
}

//go:wasmexport plugin_reset
func plugin_reset() {
	allocs = nil
}

//go:wasmexport macro_serialize
func macro_serialize(ptr int32) int32 {
	// 1. Read the null-terminated JSON string from memory
	var reqJSON []byte
	p := uintptr(ptr)
	for {
		b := *(*byte)(unsafe.Pointer(p))
		if b == 0 {
			break
		}
		reqJSON = append(reqJSON, b)
		p++
	}

	// 2. Unmarshal the request
	var req api.PluginRequest
	if err := json.Unmarshal(reqJSON, &req); err != nil {
		return returnError("failed to unmarshal request: " + err.Error())
	}

	if req.Node.Kind != api.KindType || req.Node.Type == nil {
		return returnError("serialize macro can only be applied to types")
	}

	sName := req.Node.Type.Name
	sValue := req.Node.Type.Value

	// 3. Parse the raw type value to get the AST
	l := lexer.New("type Dummy = "+sValue, "macro_input")
	p_ast := parser.New(l)
	file := p_ast.Parse("macro_input")

	if len(p_ast.Errors()) > 0 {
		return returnError(fmt.Sprintf("failed to parse struct body: %v", p_ast.Errors()))
	}

	var sl *ast.StructLiteral
	for _, stmt := range file.Statements {
		if ts, ok := stmt.(*ast.TypeStatement); ok {
			if s, ok := ts.Value.(*ast.StructLiteral); ok {
				sl = s
				break
			}
		}
	}

	if sl == nil {
		return returnError("serialize macro requires a struct literal")
	}

	// 4. Generate the Nora code
	code := generateNoraSerializationCode(sName, sl)

	// 5. Build the response
	res := api.PluginResponse{
		GeneratedCode: code,
		Node:          req.Node, // Return unmodified original node
	}

	resJSON, err := json.Marshal(res)
	if err != nil {
		return returnError("failed to marshal response: " + err.Error())
	}
	resJSON = append(resJSON, 0) // Null terminate

	// Allocate response buffer
	buf := make([]byte, len(resJSON))
	copy(buf, resJSON)
	allocs = append(allocs, buf)

	return int32(uintptr(unsafe.Pointer(&buf[0])))
}

func returnError(msg string) int32 {
	res := api.PluginResponse{Error: msg}
	resJSON, _ := json.Marshal(res)
	resJSON = append(resJSON, 0)
	buf := make([]byte, len(resJSON))
	copy(buf, resJSON)
	allocs = append(allocs, buf)
	return int32(uintptr(unsafe.Pointer(&buf[0])))
}

func generateNoraSerializationCode(sName string, sl *ast.StructLiteral) string {
	var sb strings.Builder

	sb.WriteString("\n// --- GENERATED SERIALIZATION METHODS ---\n")
	sb.WriteString("import \"serialize\"\n")
	sb.WriteString("import \"json\"\n")

	// 1. nr_serialize_json_T
	sb.WriteString(fmt.Sprintf("pub fn nr_serialize_json_%s(val: #%s) str {\n", sName, sName))
	sb.WriteString("    var res = \"{\"\n")
	first := true
	for _, f := range sl.Fields {
		if isFieldBorrowed(f) {
			continue
		}
		key := getJSONKeyOfField(f)
		if !first {
			sb.WriteString("    res = res + \",\"\n")
		}
		first = false

		sb.WriteString(fmt.Sprintf("    res = res + \"\\\"%s\\\":\"\n", key))

		baseType := getBaseTypeName(f.Type)
		if isArrayType(f.Type) {
			sb.WriteString("    res = res + \"[\"\n")
			sb.WriteString(fmt.Sprintf("    var i_%s = 0\n", f.Name.Value))
			sb.WriteString(fmt.Sprintf("    while (i_%s < len(val.%s)) {\n", f.Name.Value, f.Name.Value))
			sb.WriteString(fmt.Sprintf("        if (i_%s > 0) { res = res + \",\" }\n", f.Name.Value))
			elemBase := getBaseTypeName(f.Type)
			if isPrimitiveType(elemBase) {
				sb.WriteString(fmt.Sprintf("        res = res + serialize.nr_serialize_json_%s(#val.%s[i_%s])\n", elemBase, f.Name.Value, f.Name.Value))
			} else {
				sb.WriteString(fmt.Sprintf("        res = res + nr_serialize_json_%s(#val.%s[i_%s])\n", elemBase, f.Name.Value, f.Name.Value))
			}
			sb.WriteString(fmt.Sprintf("        i_%s = i_%s + 1\n    }\n", f.Name.Value, f.Name.Value))
			sb.WriteString("    res = res + \"]\"\n")
		} else if isPrimitiveType(baseType) {
			sb.WriteString(fmt.Sprintf("    res = res + serialize.nr_serialize_json_%s(#val.%s)\n", baseType, f.Name.Value))
		} else {
			sb.WriteString(fmt.Sprintf("    res = res + nr_serialize_json_%s(#val.%s)\n", baseType, f.Name.Value))
		}
	}
	sb.WriteString("    res = res + \"}\"\n")
	sb.WriteString("    return res\n")
	sb.WriteString("}\n\n")

	// 2. nr_deserialize_json_T
	sb.WriteString(fmt.Sprintf("pub fn nr_deserialize_json_%s(data: str) %s {\n", sName, sName))
	sb.WriteString("    var v_res = json.Parse(data)\n")
	sb.WriteString("    if (v_res.IsErr()) {\n")
	sb.WriteString("        var empty = json.JsonNull\n")
	sb.WriteString(fmt.Sprintf("        return nr_deserialize_json_%s_from_val(#empty)\n", sName))
	sb.WriteString("    }\n")
	sb.WriteString("    var v = v_res.Unwrap()\n")
	sb.WriteString(fmt.Sprintf("    return nr_deserialize_json_%s_from_val(#v)\n", sName))
	sb.WriteString("}\n\n")

	// 3. nr_deserialize_json_T_from_val
	sb.WriteString(fmt.Sprintf("pub fn nr_deserialize_json_%s_from_val(v: #json.JsonValue) %s {\n", sName, sName))

	for _, f := range sl.Fields {
		if isFieldBorrowed(f) {
			continue
		}
		var defVal string
		baseType := getBaseTypeName(f.Type)
		if isPrimitiveType(baseType) || isArrayType(f.Type) {
			defVal = getDefaultValueForField(f.Type)
			sb.WriteString(fmt.Sprintf("    var f_%s: %s = %s\n", f.Name.Value, f.Type.String(), defVal))
		} else {
			sb.WriteString(fmt.Sprintf("    var empty_json_%s = json.JsonNull\n", f.Name.Value))
			defVal = fmt.Sprintf("nr_deserialize_json_%s_from_val(#empty_json_%s)", baseType, f.Name.Value)
			sb.WriteString(fmt.Sprintf("    var f_%s: %s = %s\n", f.Name.Value, f.Type.String(), defVal))
		}
	}

	sb.WriteString("    match v {\n")
	sb.WriteString("        JsonObject(obj) => {\n")
	sb.WriteString("            var i = 0\n")
	sb.WriteString("            while (i < obj.Len[json.JsonProperty]()) {\n")
	sb.WriteString("                var prop = obj.Get[json.JsonProperty](i)\n")
	
	matchedAny := false
	for _, f := range sl.Fields {
		if isFieldBorrowed(f) {
			continue
		}
		key := getJSONKeyOfField(f)
		baseType := getBaseTypeName(f.Type)

		cond := "if"
		if matchedAny {
			cond = "} else if"
		}
		matchedAny = true
		sb.WriteString(fmt.Sprintf("        %s (prop.name == \"%s\") {\n", cond, key))

		if isArrayType(f.Type) {
			sb.WriteString("            var arr_val = prop.value.GetArray()\n")
			elemTypeStr := getArrayElementTypeString(f.Type)
			sb.WriteString(fmt.Sprintf("            var arr = alloc %s[arr_val.Len()]\n", elemTypeStr))
			sb.WriteString("            var idx = 0\n")
			sb.WriteString("            while (idx < arr_val.Len()) {\n")
			sb.WriteString("                var item = arr_val.Get(idx)\n")
			elemBase := getBaseTypeName(f.Type)
			if isPrimitiveType(elemBase) {
				if elemBase == "str" {
					sb.WriteString("                arr[idx] = \"\" + item.GetString()\n")
				} else if elemBase == "bool" {
					sb.WriteString("                arr[idx] = item.GetBool()\n")
				} else if elemBase == "f64" {
					sb.WriteString("                arr[idx] = item.GetNumber()\n")
				} else {
					sb.WriteString(fmt.Sprintf("                arr[idx] = %s(item.GetNumber())\n", elemBase))
				}
			} else {
				sb.WriteString(fmt.Sprintf("                arr[idx] = nr_deserialize_json_%s_from_val(#item)\n", elemBase))
			}
			sb.WriteString("                idx = idx + 1\n")
			sb.WriteString("            }\n")
			sb.WriteString(fmt.Sprintf("            f_%s = arr\n", f.Name.Value))
		} else if isPrimitiveType(baseType) {
			if baseType == "str" {
				sb.WriteString(fmt.Sprintf("            f_%s = \"\" + prop.value.GetString()\n", f.Name.Value))
			} else if baseType == "bool" {
				sb.WriteString(fmt.Sprintf("            f_%s = prop.value.GetBool()\n", f.Name.Value))
			} else if baseType == "f64" {
				sb.WriteString(fmt.Sprintf("            f_%s = prop.value.GetNumber()\n", f.Name.Value))
			} else {
				sb.WriteString(fmt.Sprintf("            f_%s = %s(prop.value.GetNumber())\n", f.Name.Value, baseType))
			}
		} else {
			sb.WriteString(fmt.Sprintf("            f_%s = nr_deserialize_json_%s_from_val(#prop.value)\n", f.Name.Value, baseType))
		}
	}
	if matchedAny {
		sb.WriteString("        }\n")
	}
	sb.WriteString("                i = i + 1\n")
	sb.WriteString("            }\n")
	sb.WriteString("        }\n")
	sb.WriteString("        _ => {}\n")
	sb.WriteString("    }\n\n")

	sb.WriteString(fmt.Sprintf("    return %s {\n", sName))
	for _, f := range sl.Fields {
		if isFieldBorrowed(f) {
			sb.WriteString(fmt.Sprintf("        %s: none,\n", f.Name.Value))
			continue
		}
		baseType := getBaseTypeName(f.Type)
		if isPrimitiveType(baseType) && !isArrayType(f.Type) {
			sb.WriteString(fmt.Sprintf("        %s: f_%s,\n", f.Name.Value, f.Name.Value))
		} else {
			sb.WriteString(fmt.Sprintf("        %s: @f_%s,\n", f.Name.Value, f.Name.Value))
		}
	}
	sb.WriteString("    }\n")
	sb.WriteString("}\n\n")

	return sb.String()
}

func getBaseTypeName(node ast.Node) string {
	switch n := node.(type) {
	case *ast.Identifier:
		return n.Value
	case *ast.PrefixExpression:
		return getBaseTypeName(n.Right)
	case *ast.IndexExpression:
		return getBaseTypeName(n.Left)
	}
	return ""
}

func isFieldBorrowed(f *ast.FieldDefinition) bool {
	if pref, ok := f.Type.(*ast.PrefixExpression); ok {
		return pref.Operator == "#" || pref.Operator == "&"
	}
	return false
}

func getJSONKeyOfField(f *ast.FieldDefinition) string {
	if attr := ast.GetAttribute(f.Attributes, "rename"); attr != nil {
		if len(attr.Args) > 0 {
			return stripQuotes(attr.Args[0])
		}
	}
	return f.Name.Value
}

func stripQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return s
}

func isPrimitiveType(name string) bool {
	switch name {
	case "i32", "u32", "i64", "u64", "f64", "f32", "bool", "str", "byte":
		return true
	}
	return false
}

func isArrayType(node ast.Node) bool {
	switch n := node.(type) {
	case *ast.IndexExpression:
		return len(n.Indices) == 0
	case *ast.PrefixExpression:
		return isArrayType(n.Right)
	}
	return false
}

func getArrayElementTypeString(node ast.Node) string {
	switch n := node.(type) {
	case *ast.IndexExpression:
		if len(n.Indices) == 0 {
			return n.Left.String()
		}
	case *ast.PrefixExpression:
		return getArrayElementTypeString(n.Right)
	}
	return ""
}

func getDefaultValueForField(node ast.TypeNode) string {
	base := getBaseTypeName(node)
	if isArrayType(node) {
		elemTypeStr := getArrayElementTypeString(node)
		return fmt.Sprintf("make(%s[], 0)", elemTypeStr)
	}
	switch base {
	case "i32", "u32", "i16", "u16", "i8", "u8", "byte":
		return "0"
	case "i64", "u64":
		return "i64(0)"
	case "f64":
		return "0.0"
	case "f32":
		return "f32(0.0)"
	case "bool":
		return "false"
	case "str":
		return "\"\""
	}
	// Object type
	return base + "{}"
}
