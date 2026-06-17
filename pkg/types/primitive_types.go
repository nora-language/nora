package types

// 1. Concrete Type for Primitives
type PrimitiveType struct {
	KindName string
	BitSize  int
	Methods  map[string]NRType
}

func (p *PrimitiveType) Name() string   { return p.KindName }
func (p *PrimitiveType) GetKind() Kind  { return KindPrimitive }
func (p *PrimitiveType) IsLeased() bool { return false } // Primitives are copied
func (p *PrimitiveType) Size() int      { return p.BitSize / 8 }

// 2. Global Registry of Primitives
var (
	Int    = &PrimitiveType{KindName: "int", BitSize: 64} // Default int
	I32    = &PrimitiveType{KindName: "i32", BitSize: 32}
	I64    = &PrimitiveType{KindName: "i64", BitSize: 64}
	U64    = &PrimitiveType{KindName: "u64", BitSize: 64}
	U32    = &PrimitiveType{KindName: "u32", BitSize: 32}
	U16    = &PrimitiveType{KindName: "u16", BitSize: 16}
	U8     = &PrimitiveType{KindName: "u8", BitSize: 8}
	I16    = &PrimitiveType{KindName: "i16", BitSize: 16}
	I8     = &PrimitiveType{KindName: "i8", BitSize: 8}
	Byte   = &PrimitiveType{KindName: "byte", BitSize: 8}
	F32    = &PrimitiveType{KindName: "f32", BitSize: 32}
	F64    = &PrimitiveType{KindName: "f64", BitSize: 64}
	Bool   = &PrimitiveType{KindName: "bool", BitSize: 8}
	String = &PrimitiveType{KindName: "str", BitSize: 0} // Special case
	Void   = &PrimitiveType{KindName: "void", BitSize: 0}
	Ptr    = &PrimitiveType{KindName: "ptr"}
	Fiber  = &PrimitiveType{KindName: "fiber"}

	// ErrorType is a placeholder when resolution fails
	ErrorType = &PrimitiveType{KindName: "<error>"}

	Any = &ProtocolType{
		ProtocolName: "any",
		Methods:      make(map[string]*FunctionType),
	}
)

// LookupPrimitive checks if a string name is a built-in type
func LookupPrimitive(name string) (NRType, bool) {
	switch name {
	case "any":
		return Any, true
	case "int":
		return Int, true
	case "i32":
		return I32, true
	case "i64":
		return I64, true
	case "u64":
		return U64, true
	case "u32":
		return U32, true
	case "u16":
		return U16, true
	case "u8":
		return U8, true
	case "i16":
		return I16, true
	case "i8":
		return I8, true
	case "byte":
		return Byte, true
	case "f32":
		return F32, true
	case "f64":
		return F64, true
	case "bool":
		return Bool, true
	case "str":
		return String, true
	case "void":
		return Void, true
	case "ptr":
		return Ptr, true
	case "fiber":
		return Fiber, true
	default:
		return nil, false
	}
}
