package types

import "fmt"

type ChanType struct {
	Elem NRType
}

func (t *ChanType) Name() string {
	return fmt.Sprintf("chan[%s]", t.Elem.Name())
}

func (t *ChanType) Size() int {
	return 8 // Pointer to channel structure
}

func (t *ChanType) GetKind() Kind {
	return KindChan
}

func (t *ChanType) IsLeased() bool {
	return false // Channels are handles, not leases themselves
}

func (t *ChanType) Equals(other NRType) bool {
	o, ok := other.(*ChanType)
	return ok && Equals(t.Elem, o.Elem)
}
