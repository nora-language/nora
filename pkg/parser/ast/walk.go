package ast

// Inspector is a function type that returns true if we should keep drilling down
type Inspector func(Node) bool

// Inspect traverses an AST in depth-first order:
// It calls f(node); node must not be nil. If f(node) returns true, Inspect invokes f
// recursively for each of the non-nil children of node, followed by a call of f(nil).
func Inspect(node Node, f Inspector) {
	if node == nil || IsNil(node) {
		return
	}

	// 1. Visit the current node
	if !f(node) {
		return
	}

	// 2. Traverse Children based on Node Type
	switch n := node.(type) {
	case *Program:
		for _, file := range n.Files {
			Inspect(file, f)
		}

	case *ImportStatement:
		if n.Alias != nil {
			Inspect(n.Alias, f)
		}
		if n.Path != nil {
			Inspect(n.Path, f)
		}

	case *File:
		for _, stmt := range n.Statements {
			Inspect(stmt, f)
		}

	case *FunctionStatement:
		// Visit attributes
		for i := range n.Attributes {
			Inspect(&n.Attributes[i], f)
		}
		if n.Receiver != nil {
			Inspect(n.Receiver, f)
		}
		Inspect(n.Name, f)
		for _, tp := range n.TypeParameters {
			Inspect(tp, f)
		}
		for _, p := range n.Parameters {
			Inspect(p, f)
		}
		if n.ReturnType != nil {
			Inspect(n.ReturnType, f)
		}
		Inspect(n.Body, f)

	case *BlockStatement:
		for _, stmt := range n.Statements {
			Inspect(stmt, f)
		}

	case *VarStatement:
		Inspect(n.Name, f)  // Variable Name
		Inspect(n.Value, f) // The Expression

	case *AssignmentStatement:
		Inspect(n.Left, f)

		Inspect(n.Value, f)

	case *ReturnStatement:
		Inspect(n.ReturnValue, f)

	case *ExpressionStatement:
		Inspect(n.Expression, f)

	case *GroupedExpression:
		Inspect(n.Expression, f)

	case *InfixExpression:
		Inspect(n.Left, f)
		Inspect(n.Right, f)

	case *RangeExpression:
		Inspect(n.Start, f)
		Inspect(n.End, f)

	case *CallExpression:
		Inspect(n.Function, f)
		for _, arg := range n.Arguments {
			Inspect(arg, f)
		}

	case *ArgumentsExpression:
		Inspect(n.Value, f)

		// --- ADD THIS CASE ---
	case *StructLiteral:
		// 1. Visit the Struct Name (e.g. "User")
		if n.Name != nil {
			Inspect(n.Name, f)
		}

		// 2. Visit the Fields (e.g. name: "test", age: x)
		for _, field := range n.Fields {
			if field.Name != nil {
				Inspect(field.Name, f)
			}
			if field.Type != nil {
				Inspect(field.Type, f)
			}
			if field.Value != nil {
				Inspect(field.Value, f)
			}
		}

	case *IfExpression:
		Inspect(n.Condition, f)
		Inspect(n.Consequence, f)
		if n.Alternative != nil {
			Inspect(n.Alternative, f)
		}

	case *WhileStatement:
		Inspect(n.Condition, f)
		Inspect(n.Body, f)

	case *ForStatement:
		if n.Key != nil {
			Inspect(n.Key, f)
		}
		if n.Value != nil {
			Inspect(n.Value, f)
		}
		if n.Iterable != nil {
			Inspect(n.Iterable, f)
		}
		Inspect(n.Body, f)

	case *PinStatement:
		for _, target := range n.Targets {
			Inspect(target, f)
		}

	case *MatchExpression:
		Inspect(n.Target, f)
		for _, c := range n.Cases {
			Inspect(c.Pattern, f)
			Inspect(c.Body, f)
		}

	case *IndexExpression:
		Inspect(n.Left, f)
		for _, idx := range n.Indices {
			Inspect(idx, f)
		}

	case *PrefixExpression:
		Inspect(n.Right, f)

	case *SelectorExpression:
		Inspect(n.Left, f)
		Inspect(n.Field, f)

	case *SelectStatement:
		for _, c := range n.Cases {
			if c.Condition != nil {
				Inspect(c.Condition, f)
			}
			Inspect(c.Body, f)
		}

	case *SpawnExpression:
		if n.Call != nil {
			Inspect(n.Call, f)
		}
		if n.Body != nil {
			Inspect(n.Body, f)
		}

	case *StructInstantiation:
		Inspect(n.Type, f)
		for _, expr := range n.Fields {
			Inspect(expr, f)
		}

	case *ArrayLiteral:
		for _, el := range n.Elements {
			Inspect(el, f)
		}

	case *AllocExpression:
		Inspect(n.Value, f)

	case *MapLiteral:
		for k, v := range n.Pairs {
			Inspect(k, f)
			Inspect(v, f)
		}

	case *InterpolatedString:
		for _, part := range n.Parts {
			Inspect(part, f)
		}

	case *SendExpression:
		Inspect(n.Left, f)
		Inspect(n.Right, f)

	case *ReceiveExpression:
		Inspect(n.Value, f)

	case *TypeParameter:
		Inspect(n.Name, f)
		if n.Constraint != nil {
			Inspect(n.Constraint, f)
		}

	case *TypeStatement:
		// Visit attributes
		for i := range n.Attributes {
			Inspect(&n.Attributes[i], f)
		}
		Inspect(n.Name, f)
		for _, tp := range n.TypeParameters {
			Inspect(tp, f)
		}
		Inspect(n.Value, f)

	case *ParallelExpression:
		if n.Body != nil {
			Inspect(n.Body, f)
		}

	case *ScopeExpression:
		if n.Body != nil {
			Inspect(n.Body, f)
		}

	case *LambdaExpression:
		for _, p := range n.Parameters {
			Inspect(p, f)
		}
		if n.ReturnType != nil {
			Inspect(n.ReturnType, f)
		}
		if n.Body != nil {
			Inspect(n.Body, f)
		}

	case *DeferStatement:
		if n.Call != nil {
			Inspect(n.Call, f)
		}

	case *SumTypeLiteral:
		for _, v := range n.Variants {
			Inspect(v, f)
		}

	case *VariantDefinition:
		if n.Name != nil {
			Inspect(n.Name, f)
		}
		for _, field := range n.Fields {
			Inspect(field, f)
		}

	case *SliceExpression:
		if n.Start != nil {
			Inspect(n.Start, f)
		}
		if n.End != nil {
			Inspect(n.End, f)
		}

	case *TryExpression:
		Inspect(n.Value, f)

	case *InterfaceLiteral:
		for _, m := range n.Methods {
			Inspect(m, f)
		}
		for _, e := range n.Embedded {
			Inspect(e, f)
		}

	case *Parameter:
		if n.Name != nil {
			Inspect(n.Name, f)
		}
		if n.Type != nil {
			Inspect(n.Type, f)
		}

	case *PackageStatement:
		if n.Name != nil {
			Inspect(n.Name, f)
		}

	case *ExternStatement:
		if n.Function != nil {
			Inspect(n.Function, f)
		}

	case *ExportStatement:
		if n.Node != nil {
			Inspect(n.Node, f)
		}

	case *FieldDefinition:
		// Visit attributes
		for i := range n.Attributes {
			Inspect(&n.Attributes[i], f)
		}
		if n.Name != nil {
			Inspect(n.Name, f)
		}
		if n.Type != nil {
			Inspect(n.Type, f)
		}
		if n.Value != nil {
			Inspect(n.Value, f)
		}

	case *FunctionType:
		for _, p := range n.Parameters {
			Inspect(p, f)
		}
		if n.ReturnType != nil {
			Inspect(n.ReturnType, f)
		}

	case *ChanType:
		if n.Value != nil {
			Inspect(n.Value, f)
		}

	// Leaf nodes (no children to traverse)
	case *Identifier, *IntegerLiteral, *StringLiteral, *Boolean, *FloatLiteral, *ImaginaryLiteral, *NoneLiteral, *RuneLiteral, *BreakStatement, *ContinueStatement, *Comment, *CommentGroup, *Attribute:
		// Do nothing, just visited
	}

	// 3. Post-visit (optional standard behavior)
	f(nil)
}
