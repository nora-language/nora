package ast

import (
	"github.com/DwiYI/Project-Nora/pkg/token"
	"strings"
)

// Comment represents a single /// or // comment
type Comment struct {
	Token token.Token // The DOC_COMMENT or COMMENT token
	Text  string      // The comment text (excluding /// or //)
}

func (c *Comment) expressionNode()      {}
func (c *Comment) TokenLiteral() string { return c.Token.Literal }
func (c *Comment) String() string {
	if c.Token.Type == token.DOC_COMMENT {
		return "/// " + c.Text
	}
	return "// " + c.Text
}
func (c *Comment) Pos() token.Position { return c.Token.Position }

// CommentGroup represents a sequence of comments with no empty lines between them.
type CommentGroup struct {
	List []*Comment
}

func (cg *CommentGroup) expressionNode() {}
func (cg *CommentGroup) TokenLiteral() string {
	if len(cg.List) > 0 {
		return cg.List[0].TokenLiteral()
	}
	return ""
}

func (cg *CommentGroup) Pos() token.Position {
	if len(cg.List) > 0 {
		return cg.List[0].Pos()
	}
	return token.Position{}
}

func (cg *CommentGroup) String() string {
	var out strings.Builder
	for _, c := range cg.List {
		out.WriteString(c.String())
		out.WriteString("\n")
	}
	return out.String()
}

// Text returns the combined text of the comments
func (cg *CommentGroup) Text() string {
	if cg == nil {
		return ""
	}
	var out strings.Builder
	for i, c := range cg.List {
		out.WriteString(c.Text)
		if i < len(cg.List)-1 {
			out.WriteString("\n")
		}
	}
	return out.String()
}
