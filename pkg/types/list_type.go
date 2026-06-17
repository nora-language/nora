package types

type ListType struct {
	ElementType NRType
}

func (l *ListType) Name() string   { return "List<" + l.ElementType.Name() + ">" }
func (l *ListType) GetKind() Kind  { return KindList }
func (l *ListType) IsLeased() bool { return false } // Owned by default
func (l *ListType) Size() int      { return 8 }     // Pointer size in C11
