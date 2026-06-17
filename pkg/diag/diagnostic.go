package diag

import "fmt"

type Severity int

const (
	Error Severity = iota
	Warning
	Info
)

func (s Severity) String() string {
	switch s {
	case Error:
		return "Error"
	case Warning:
		return "Warning"
	case Info:
		return "Info"
	default:
		return "Unknown"
	}
}

type Position struct {
	Line   int
	Column int
	Offset int
}

type Range struct {
	Start Position
	End   Position
}

type Diagnostic struct {
	Range    Range
	Severity Severity
	Message  string
	Code     string
	Source   string // Stage: Lexer, Parser, Semantic
	File     string
	Notes    []string // Additional context
	Hint     string   // Suggested fix
}

func (d Diagnostic) String() string {
	return fmt.Sprintf("%s:%d:%d: [%s] %s", d.File, d.Range.Start.Line, d.Range.Start.Column, d.Source, d.Message)
}

type Collection struct {
	Diagnostics []Diagnostic
}

func (c *Collection) Add(d Diagnostic) {
	c.Diagnostics = append(c.Diagnostics, d)
}

func (c *Collection) HasErrors() bool {
	for _, d := range c.Diagnostics {
		if d.Severity == Error {
			return true
		}
	}
	return false
}

func (c *Collection) ErrorMessages() []string {
	var msgs []string
	for _, d := range c.Diagnostics {
		msgs = append(msgs, d.String())
	}
	return msgs
}
