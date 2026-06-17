package types

import "bytes"

type StructType struct {
	TypeName   string
	Fields     map[string]NRType // Resolved field types
	FieldNames []string          // Keys of Fields map in definition order
	Methods    map[string]NRType // Resolved method types (FunctionType)
	TypeParams []*TypeParam      // For generic structs
	TypeArgs   []NRType          // Arguments used for specialization
	BaseType   *StructType       // Pointer to the base generic struct
	IsShared   bool              // Marked with [shared] attribute
	CoreIntrinsic string         // Extracted from [core_intrinsic("name")]
}

func NewStructType(name string) *StructType {
	return &StructType{
		TypeName:   name,
		Fields:     make(map[string]NRType),
		FieldNames: []string{},
		Methods:    make(map[string]NRType),
	}
}

// Implement NRType Interface
func (s *StructType) Name() string   { return s.TypeName }
func (s *StructType) GetKind() Kind  { return KindStruct }
func (s *StructType) IsLeased() bool { return false } // Owned by default
func (s *StructType) Size() int      { return 0 }     // Calculate based on fields later

func (s *StructType) String() string {
	var out bytes.Buffer
	out.WriteString("struct " + s.TypeName + " {")
	// (String logic for fields...)
	out.WriteString("}")
	return out.String()
}
