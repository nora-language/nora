package types

type Kind int

const (
	KindPrimitive Kind = iota
	KindStruct
	KindSum      // Enums
	KindProtocol // Interfaces
	KindList
	KindMap
	KindModule // <--- Add this
	KindFunction
	KindChan
	KindGeneric
	KindPointer
)

type LeaseKind int

const (
	LeaseRead  LeaseKind = iota // Default
	LeaseWrite                  // #
	LeaseMove                   // @
)

// NRType is the interface that ALL types (including Modules) must implement.
type NRType interface {
	Name() string
	GetKind() Kind
	IsLeased() bool
	Size() int
}

// Equals checks if two types are compatible.
func Equals(t1, t2 NRType) bool {
	// 1. Identity Check (Fast path)
	if t1 == t2 {
		return true
	}

	// 2. Nil Check
	if t1 == nil || t2 == nil {
		return false
	}



	// 4. Kind Check (Optimization: "struct" can never equal "int")
	if t1.GetKind() != t2.GetKind() {
		return false
	}

	// 4. Specific Logic
	switch t1.GetKind() {
	case KindPrimitive:
		// "i32" == "i32"
		return t1.Name() == t2.Name()

	case KindStruct:
		s1 := t1.(*StructType)
		s2 := t2.(*StructType)

		// A. Nominal Check: If they are named types (e.g. "User"), names MUST match.
		// "User" != "Account" even if fields are same.
		if s1.Name() != "" || s2.Name() != "" {
			return s1.Name() == s2.Name()
		}

		// B. Structural Check: If both are Anonymous (struct { x: int }), compare fields.
		if len(s1.Fields) != len(s2.Fields) {
			return false
		}

		for name, fieldType1 := range s1.Fields {
			fieldType2, exists := s2.Fields[name]
			if !exists {
				return false // Field name missing
			}
			// Recursive Check: Checks if "i32" == "i32" inside the struct
			if !Equals(fieldType1, fieldType2) {
				return false
			}
		}
		return true

	case KindModule:
		// Modules match if they refer to the same package path
		return t1.Name() == t2.Name()

	// Future: KindList, KindMap
	case KindList:
		return Equals(t1.(*ListType).ElementType, t2.(*ListType).ElementType)

	case KindMap:
		m1 := t1.(*MapType)
		m2 := t2.(*MapType)
		return Equals(m1.Key, m2.Key) && Equals(m1.Value, m2.Value)

	case KindSum:
		return t1.Name() == t2.Name()

	case KindPointer:
		p1 := t1.(*PointerType)
		p2 := t2.(*PointerType)
		return p1.IsArray == p2.IsArray && p1.Leased == p2.Leased && p1.Kind == p2.Kind && Equals(p1.Base, p2.Base)

	case KindChan:
		return Equals(t1.(*ChanType).Elem, t2.(*ChanType).Elem)

	case KindFunction:
		f1, ok1 := t1.(*FunctionType)
		f2, ok2 := t2.(*FunctionType)
		if !ok1 || !ok2 {
			return false
		}
		if len(f1.Params) != len(f2.Params) {
			return false
		}
		for i := range f1.Params {
			t1Base := f1.Params[i]
			if pt, ok := t1Base.(*PointerType); ok && pt.Leased {
				t1Base = pt.Base
			}
			t2Base := f2.Params[i]
			if pt, ok := t2Base.(*PointerType); ok && pt.Leased {
				t2Base = pt.Base
			}
			if !Equals(t1Base, t2Base) {
				return false
			}
		}
		t1Ret := f1.Return
		if pt, ok := t1Ret.(*PointerType); ok && pt.Leased {
			t1Ret = pt.Base
		}
		t2Ret := f2.Return
		if pt, ok := t2Ret.(*PointerType); ok && pt.Leased {
			t2Ret = pt.Base
		}
		if !Equals(t1Ret, t2Ret) {
			return false
		}
		if f1.IsVariadic != f2.IsVariadic {
			return false
		}
		return true

	case KindProtocol:
		// Structural equality for interfaces (protocols)
		p1 := t1.(*ProtocolType)
		p2 := t2.(*ProtocolType)
		if p1.ProtocolName != "" || p2.ProtocolName != "" {
			return p1.ProtocolName == p2.ProtocolName
		}
		if len(p1.Methods) != len(p2.Methods) {
			return false
		}
		for name, m1 := range p1.Methods {
			m2, exists := p2.Methods[name]
			if !exists || !Equals(m1, m2) {
				return false
			}
		}
		return true

	default:
		// Fallback for types without special internal structure
		return t1.Name() == t2.Name()
	}
}

// IsPointerLike returns true for types that are raw pointers in C (str, ptr, *[]T)
func IsPointerLike(t NRType) bool {
	if t == nil {
		return false
	}
	name := t.Name()
	if name == "str" || name == "ptr" {
		return true
	}
	if pt, ok := t.(*PointerType); ok {
		// Leased primitive types (except 'str' and 'ptr') under LeaseRead are passed by value in C, so they are not pointer-like.
		if pt.Leased && !pt.IsArray {
			if pt.Base.GetKind() == KindPrimitive && pt.Base.Name() != "str" && pt.Base.Name() != "ptr" {
				if pt.Kind == LeaseRead {
					return false
				}
			}
		}
		// Protocol types (interfaces) are fat pointers (16-byte structs) in C, so they cannot be type-erased to void*.
		if _, isProtocol := pt.Base.(*ProtocolType); isProtocol {
			return false
		}
		return true
	}
	return false
}

func underlyingStructOrBase(t NRType) NRType {
	for {
		if pt, ok := t.(*PointerType); ok {
			if pt.IsArray {
				break
			}
			t = pt.Base
		} else {
			break
		}
	}
	return t
}

func unwrapListOrArray(t NRType) (NRType, bool) {
	if t == nil {
		return nil, false
	}
	t = underlyingStructOrBase(t)
	if lt, ok := t.(*ListType); ok {
		return lt.ElementType, true
	}
	if pt, ok := t.(*PointerType); ok && pt.IsArray {
		return pt.Base, true
	}
	return nil, false
}

func getLeaseInfo(t NRType) (bool, LeaseKind) {
	if pt, ok := t.(*PointerType); ok && pt.Leased {
		return true, pt.Kind
	}
	return false, LeaseRead
}

func checkLeaseCompatibility(toLeased, fromLeased bool, toKind, fromKind LeaseKind) bool {
	if toLeased {
		if toKind == LeaseRead {
			return true
		}
		if toKind == LeaseWrite {
			return !fromLeased || fromKind == LeaseWrite || fromKind == LeaseMove
		}
		if toKind == LeaseMove {
			return !fromLeased || fromKind == LeaseMove
		}
	} else {
		return !fromLeased || fromKind == LeaseMove
	}
	return false
}

func isByteType(t NRType) bool {
	if t == nil {
		return false
	}
	name := t.Name()
	return name == "byte" || name == "i8"
}

func isBytePointer(t NRType) bool {
	t = underlyingStructOrBase(t)
	if pt, ok := t.(*PointerType); ok {
		return isByteType(pt.Base)
	}
	if lt, ok := t.(*ListType); ok {
		return isByteType(lt.ElementType)
	}
	return false
}

// isIntLike returns true for integer-compatible primitive types
func isIntLike(t NRType) bool {
	if t == nil || t.GetKind() != KindPrimitive {
		return false
	}
	name := t.Name()
	return name == "int" || name == "i64" || name == "i32" || name == "i16" || name == "i8" ||
		name == "u64" || name == "u32" || name == "u16" || name == "u8" || name == "byte"
}

// IsOwnedType returns true for types that follow move/drop semantics (Strings, Structs, SumTypes, Lists, Maps, Channels)
func IsOwnedType(t NRType) bool {
	if t == nil {
		return false
	}
	kind := t.GetKind()
	if kind == KindStruct || kind == KindSum || kind == KindList || kind == KindMap || kind == KindChan || kind == KindGeneric || kind == KindProtocol || kind == KindFunction {
		return true
	}
	// Strings are primitive but owned in Nora
	if t.Name() == "str" {
		return true
	}
	// Pointers are owned if they are not leases
	if pt, ok := t.(*PointerType); ok {
		if pt.IsArray {
			return true
		}
		// All non-leased pointers are owned in Nora
		if !pt.Leased {
			return true
		}
		// Move leases (@T) are also owned
		if pt.Kind == LeaseMove {
			return true
		}
	}
	return false
}

// IsAssignable checks if `from` can be assigned to `to`, including C-compatible coercions.
// More permissive than Equals — allows implicit coercions safe in C.
func IsAssignable(to, from NRType) bool {
	if Equals(to, from) {
		return true
	}
	if to == nil || from == nil {
		return false
	}

	if protoTo, ok := to.(*ProtocolType); ok {
		if protoFrom, ok := from.(*ProtocolType); ok {
			if len(protoTo.Methods) == 0 {
				return true
			}
			for name, mTo := range protoTo.Methods {
				mFrom, exists := protoFrom.Methods[name]
				if !exists || !Equals(mTo, mFrom) {
					return false
				}
			}
			return true
		}
	}

	// [NEW] Allow assigning base type T to lease #T (implicit borrow)
	if pt, ok := to.(*PointerType); ok && pt.Leased && Equals(pt.Base, from) {
		return true
	}

	// [NEW] Allow assigning move lease @T to base type T (implicit move-load),
	// or allowing leased strings/pointers to be implicitly unwrapped when assigned to primitive str/ptr
	// Also allow assigning any lease of a non-owned (copy-by-value) type to its base type.
	if pt, ok := from.(*PointerType); ok && pt.Leased {
		if Equals(underlyingStructOrBase(pt.Base), underlyingStructOrBase(to)) {
			if to.Name() == "str" || to.Name() == "ptr" || pt.Kind == LeaseMove || !IsOwnedType(to) {
				return true
			}
		}
	}

	// 1. Check list/array compatibility (e.g. #[]T, #@[]T, []T, *[]T)
	elTo, okToArr := unwrapListOrArray(to)
	elFrom, okFromArr := unwrapListOrArray(from)
	if okToArr && okFromArr {
		if Equals(elTo, elFrom) {
			toLeased, toKind := getLeaseInfo(to)
			fromLeased, fromKind := getLeaseInfo(from)
			if checkLeaseCompatibility(toLeased, fromLeased, toKind, fromKind) {
				return true
			}
		}
	}

	// 2. Check structured pointer lease and base-type compatibility
	pTo, okTo := to.(*PointerType)
	pFrom, okFrom := from.(*PointerType)
	if okTo && okFrom && !pTo.IsArray && !pFrom.IsArray {
		baseTo := underlyingStructOrBase(pTo.Base)
		baseFrom := underlyingStructOrBase(pFrom.Base)
		if Equals(baseTo, baseFrom) {
			toLeased, toKind := getLeaseInfo(to)
			fromLeased, fromKind := getLeaseInfo(from)
			if checkLeaseCompatibility(toLeased, fromLeased, toKind, fromKind) {
				return true
			}
		}
	}

	// 3. Allow raw 'ptr' assignments, but enforce strict pointer-like properties
	if IsPointerLike(to) && IsPointerLike(from) {
		toBase := underlyingStructOrBase(to)
		fromBase := underlyingStructOrBase(from)
		if toBase.Name() == "ptr" || fromBase.Name() == "ptr" {
			return true
		}
		// Allow string <-> byte-array conversions
		if toBase.Name() == "str" && isBytePointer(from) {
			return true
		}
		if fromBase.Name() == "str" && isBytePointer(to) {
			return true
		}
	}

	// 4. Allow implicit assignment of function pointers to 'ptr'
	if to.Name() == "ptr" && from.GetKind() == KindFunction {
		return true
	}

	// Allow integer widening/narrowing: byte, i8, i32, int are all compatible
	if isIntLike(to) && isIntLike(from) {
		return true
	}
	return false
}

type GenericType struct {
	TypeParam  string
	Constraint NRType
}

func (g *GenericType) Name() string   { return g.TypeParam }
func (g *GenericType) GetKind() Kind  { return KindGeneric }
func (g *GenericType) IsLeased() bool { return false }
func (g *GenericType) Size() int      { return 8 }

type ProtocolType struct {
	ProtocolName string
	Methods      map[string]*FunctionType
	TypeParams   []*TypeParam
	TypeArgs     []NRType
	BaseType     *ProtocolType
	IsShared     bool
}

func (p *ProtocolType) Name() string   { return p.ProtocolName }
func (p *ProtocolType) GetKind() Kind  { return KindProtocol }
func (p *ProtocolType) IsLeased() bool { return false }
func (p *ProtocolType) Size() int      { return 16 } // Fat pointer: {void* data, void* vtable}

type PointerType struct {
	Base    NRType
	IsArray bool      // If true, this is a heap-allocated array with a size header
	Leased  bool      // If true, this represents a #T or @T lease
	Kind    LeaseKind // For leases: LeaseRead (#) or LeaseMove (@)
}

func (p *PointerType) Name() string {
	prefix := ""
	if p.Leased {
		switch p.Kind {
		case LeaseRead:
			prefix = "#"
		case LeaseWrite:
			prefix = "&"
		case LeaseMove:
			prefix = "@"
		}
	}
	if p.IsArray {
		return prefix + "(" + p.Base.Name() + ")[]"
	}
	return prefix + p.Base.Name()
}
func (p *PointerType) GetKind() Kind { return KindPointer }
func (p *PointerType) IsLeased() bool {
	return p.Leased
}
func (p *PointerType) Size() int { return 8 }

type TypeParam struct {
	Name       string
	Constraint NRType
}

func UnwrapLease(t NRType) NRType {
	for {
		if pt, ok := t.(*PointerType); ok && pt.Leased && !pt.IsArray && (pt.Kind == LeaseRead || pt.Kind == LeaseWrite || pt.Kind == LeaseMove) {
			t = pt.Base
		} else {
			break
		}
	}
	return t
}
