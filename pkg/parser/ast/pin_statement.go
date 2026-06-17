package ast

import (
	"fmt"

	"github.com/nora-language/nora/pkg/token"
)

type PinStatement struct {
	Token   token.Token   // the 'pin' token
	Targets []*Identifier // The variables to pin
}

func (ps *PinStatement) statementNode()       {}
func (ps *PinStatement) TokenLiteral() string { return ps.Token.Literal }
func (ps *PinStatement) Pos() token.Position  { return ps.Token.Position }

func (ps *PinStatement) String() string {
	targets := ""
	for i, t := range ps.Targets {
		if i > 0 {
			targets += ", "
		}
		targets += t.String()
	}
	return fmt.Sprintf("pin %s", targets)
}
