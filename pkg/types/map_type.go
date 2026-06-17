package types

type MapType struct {
	Key   NRType
	Value NRType
}

func (m *MapType) Name() string   { return "map[" + m.Key.Name() + "]" + m.Value.Name() }
func (m *MapType) GetKind() Kind  { return KindMap }
func (m *MapType) IsLeased() bool { return false }
func (m *MapType) Size() int      { return 8 } // Pointer to hash map
