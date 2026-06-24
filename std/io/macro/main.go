package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"unsafe"

	"github.com/nora-language/nora/pkg/plugin/api"
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

func readRequest(ptr int32) (*api.CallRequest, error) {
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

	var req api.CallRequest
	if err := json.Unmarshal(reqJSON, &req); err != nil {
		return nil, err
	}
	return &req, nil
}

func writeResponse(res api.CallResponse) int32 {
	resJSON, _ := json.Marshal(res)
	resJSON = append(resJSON, 0) // Null terminate
	buf := make([]byte, len(resJSON))
	copy(buf, resJSON)
	allocs = append(allocs, buf)
	return int32(uintptr(unsafe.Pointer(&buf[0])))
}

func getFmtSpec(typeName string, valueRaw string) string {
	if typeName == "" {
		return "%s"
	}
	t := strings.TrimLeft(typeName, "#&@*")
	switch t {
	case "str", "bool":
		return "%s"
	case "i8", "i16", "i32", "int":
		return "%d"
	case "ptr":
		return "%p"
	case "u8", "u16", "u32":
		return "%u"
	case "i64":
		return "%lld"
	case "u64":
		return "%llu"
	case "f32":
		return "%f"
	case "f64":
		return "%lf"
	}
	return "%s"
}

func buildPrintfImpl(req *api.CallRequest, newline bool, useStderr bool) int32 {
	target := "printf("
	if useStderr {
		target = "fprintf(stderr, "
	}

	if len(req.Arguments) == 0 {
		if newline {
			if useStderr {
				return writeResponse(api.CallResponse{ReplacementCode: `fprintf(stderr, "\n");`})
			} else {
				return writeResponse(api.CallResponse{ReplacementCode: `printf("\n");`})
			}
		} else {
			return writeResponse(api.CallResponse{ReplacementCode: `(void)0;`})
		}
	}

	var sb strings.Builder
	sb.WriteString("{ ")

	for i, arg := range req.Arguments {
		val := arg.Value
		if val == "" {
			val = `""`
		}

		spec := getFmtSpec(arg.Type, val)
		sep := " "
		if i == len(req.Arguments)-1 {
			if newline {
				sep = "\\n"
			} else {
				sep = ""
			}
		}

		if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") && len(val) >= 2 {
			s := val[1 : len(val)-1]
			sb.WriteString(fmt.Sprintf(`%s"%%s%%s", "%s", "%s"); `, target, s, sep))
		} else if strings.TrimLeft(arg.Type, "#&@*") == "bool" {
			sb.WriteString(fmt.Sprintf(`%s"%%s%%s", (%s) ? "true" : "false", "%s"); `, target, val, sep))
		} else {
			sb.WriteString(fmt.Sprintf(`%s"%s%%s", %s, "%s"); `, target, spec, val, sep))
		}
	}
	sb.WriteString("}")

	return writeResponse(api.CallResponse{ReplacementCode: sb.String()})
}

//go:wasmexport expand_println
func expand_println(ptr int32) int32 {
	req, err := readRequest(ptr)
	if err != nil {
		return writeResponse(api.CallResponse{Error: err.Error()})
	}
	return buildPrintfImpl(req, true, false)
}

//go:wasmexport expand_print
func expand_print(ptr int32) int32 {
	req, err := readRequest(ptr)
	if err != nil {
		return writeResponse(api.CallResponse{Error: err.Error()})
	}
	return buildPrintfImpl(req, false, false)
}

//go:wasmexport expand_eprintln
func expand_eprintln(ptr int32) int32 {
	req, err := readRequest(ptr)
	if err != nil {
		return writeResponse(api.CallResponse{Error: err.Error()})
	}
	return buildPrintfImpl(req, true, true)
}

//go:wasmexport expand_eprint
func expand_eprint(ptr int32) int32 {
	req, err := readRequest(ptr)
	if err != nil {
		return writeResponse(api.CallResponse{Error: err.Error()})
	}
	return buildPrintfImpl(req, false, true)
}

func getScanFmtSpec(typeName string) string {
	if typeName == "" {
		return "%s"
	}
	t := strings.TrimLeft(typeName, "#&@*")
	switch t {
	case "str":
		return "%s"
	case "i8", "i16", "i32", "int", "bool":
		return "%d"
	case "u8", "u16", "u32":
		return "%u"
	case "i64":
		return "%lld"
	case "u64":
		return "%llu"
	case "f32":
		return "%f"
	case "f64":
		return "%lf"
	case "ptr":
		return "%p"
	}
	return "%s"
}

func isPassedByValue(typeName string) bool {
	if strings.HasPrefix(typeName, "#") {
		t := typeName[1:]
		switch t {
		case "i8", "i16", "i32", "int", "u8", "u16", "u32", "i64", "u64", "f32", "f64", "bool":
			return true
		}
	}
	return false
}

func buildScanf(req *api.CallRequest, scanln bool) int32 {
	var fmtStr strings.Builder
	var argsStr strings.Builder

	for i, arg := range req.Arguments {
		val := arg.Value
		typeName := arg.Type

		spec := getScanFmtSpec(typeName)
		fmtStr.WriteString(spec)

		if i > 0 {
			argsStr.WriteString(", ")
		}

		isStr := typeName == "str" || strings.TrimLeft(typeName, "#&@*") == "str"
		isPtr := strings.HasPrefix(typeName, "*") || strings.HasPrefix(typeName, "#") ||
			strings.HasPrefix(typeName, "&") || strings.HasPrefix(typeName, "@") ||
			typeName == "ptr"

		if isPassedByValue(typeName) {
			isPtr = false
		}

		if !isStr && !isPtr {
			argsStr.WriteString("&")
		}
		argsStr.WriteString(val)
	}

	if scanln && fmtStr.Len() > 0 {
		fmtStr.WriteString(" ")
	}

	var cCode string
	if argsStr.Len() > 0 {
		cCode = fmt.Sprintf(`scanf("%s", %s)`, fmtStr.String(), argsStr.String())
	} else {
		if scanln {
			cCode = `scanf(" ");`
		} else {
			cCode = `(void)0;`
		}
	}

	return writeResponse(api.CallResponse{ReplacementCode: cCode})
}

//go:wasmexport expand_scan
func expand_scan(ptr int32) int32 {
	req, err := readRequest(ptr)
	if err != nil {
		return writeResponse(api.CallResponse{Error: err.Error()})
	}
	return buildScanf(req, false)
}

//go:wasmexport expand_scanln
func expand_scanln(ptr int32) int32 {
	req, err := readRequest(ptr)
	if err != nil {
		return writeResponse(api.CallResponse{Error: err.Error()})
	}
	return buildScanf(req, true)
}
