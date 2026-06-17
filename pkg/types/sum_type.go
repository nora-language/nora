package types

type SumType struct {
	TypeName   string
	TypeParams []*TypeParam // For generics: Option[T]
	TypeArgs   []NRType
	BaseType   *SumType
	Variants   map[string]*Variant
	Methods    map[string]NRType // Added to support receiver methods on SumTypes!
	CoreIntrinsic string         // Extracted from [core_intrinsic("name")]
}

type Variant struct {
	Name       string
	Tag        int               // Added for dispatch
	Fields     map[string]NRType // Optional data
	FieldNames []string          // Preserves field order for positional matching
}

func (s *SumType) Name() string   { return s.TypeName }
func (s *SumType) GetKind() Kind  { return KindSum }
func (s *SumType) IsLeased() bool { return false } // Owned by default
func (s *SumType) Size() int {
	// Size is: Tag size + Max(Variant data size)
	// For simplicity in C, we'll treat it as a heap-allocated box or a fat struct.
	return 24 // Placeholder: {int tag, char data[16]}
}
