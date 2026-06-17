package types

// pkg/types/function.go
type FunctionType struct {
	Params        []NRType
	ParamLeases   []LeaseKind
	Return        NRType
	IsVariadic    bool
	IsMethod      bool
	Receiver      NRType
	ReceiverLease LeaseKind
	CapturesLease bool // True if the closure captured local leases
}

func (ft *FunctionType) Name() string   { return "fn(...)" }    // specific logic later
func (ft *FunctionType) GetKind() Kind  { return KindFunction } // Add KindFunction to Enum
func (ft *FunctionType) IsLeased() bool { return false }
func (ft *FunctionType) Size() int      { return 16 } // Fat pointer: 8 for env, 8 for fn ptr
