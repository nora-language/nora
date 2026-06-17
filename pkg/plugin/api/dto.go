package api

// PluginRequest represents the payload sent to a WASM macro.
type PluginRequest struct {
	AttributeName string   `json:"attribute_name"`
	Args          []string `json:"args"`
	Node          NodeDTO  `json:"node"`
}

// PluginResponse represents the payload returned by a WASM macro.
type PluginResponse struct {
	Error         string  `json:"error,omitempty"`
	Node          NodeDTO `json:"node,omitempty"`
	GeneratedCode string  `json:"generated_code,omitempty"`
}

// NodeKind identifies the type of the DTO node.
type NodeKind string

const (
	KindFunction NodeKind = "Function"
	KindType     NodeKind = "Type"
)

// NodeDTO is the base data transfer object for all AST nodes.
type NodeDTO struct {
	Kind NodeKind `json:"kind"`

	// Only one of these will be populated based on the Kind.
	Function *FunctionDTO `json:"function,omitempty"`
	Type     *TypeDTO     `json:"type,omitempty"`
}

// FunctionDTO represents a function declaration.
// For now, we pass the body as a raw string of statements to avoid deeply serializing
// the entire expression/statement tree, which makes the API much stabler.
type FunctionDTO struct {
	Name       string   `json:"name"`
	Parameters []string `json:"parameters"`
	ReturnType string   `json:"return_type"`
	IsExport   bool     `json:"is_export"`
	IsExtern   bool     `json:"is_extern"`
	Body       string   `json:"body_raw"` // Optional: raw string representation of the body
}

// TypeDTO represents a type alias or struct declaration.
type TypeDTO struct {
	Name  string `json:"name"`
	Value string `json:"value_raw"` // Raw string representation of the type value
}
