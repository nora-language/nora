package topology

import (
	"github.com/nora-language/nora/pkg/parser/ast"
	"github.com/nora-language/nora/pkg/semantic"
)

type CFGBlockKind int

const (
	BlockSimple          CFGBlockKind = iota
	BlockBranchCondition              // If or Match condition
	BlockLoopCondition                // While or For condition
	BlockJoin                         // Join point after branches/loops
	BlockExit                         // Function exit node
)

// CFGNode represents a single statement or control node in the flow graph
type CFGNode struct {
	ID    int
	Kind  CFGBlockKind
	Stmt  ast.Statement
	Expr  ast.Expression // Sub-expression (e.g., condition expression)
	Preds []*CFGNode
	Succs []*CFGNode

	// Liveness sets
	LiveIn  map[*semantic.Symbol]bool
	LiveOut map[*semantic.Symbol]bool

	// Data-flow sets
	Gen  map[*semantic.Symbol]bool
	Kill map[*semantic.Symbol]bool
}

func NewCFGNode(id int, kind CFGBlockKind, stmt ast.Statement, expr ast.Expression) *CFGNode {
	return &CFGNode{
		ID:      id,
		Kind:    kind,
		Stmt:    stmt,
		Expr:    expr,
		Preds:   make([]*CFGNode, 0),
		Succs:   make([]*CFGNode, 0),
		LiveIn:  make(map[*semantic.Symbol]bool),
		LiveOut: make(map[*semantic.Symbol]bool),
		Gen:     make(map[*semantic.Symbol]bool),
		Kill:    make(map[*semantic.Symbol]bool),
	}
}

// AddEdge creates a directed control-flow edge from fromNode to toNode
func AddEdge(fromNode, toNode *CFGNode) {
	if fromNode == nil || toNode == nil {
		return
	}
	// Avoid duplicate edges
	for _, succ := range fromNode.Succs {
		if succ == toNode {
			return
		}
	}
	fromNode.Succs = append(fromNode.Succs, toNode)
	toNode.Preds = append(toNode.Preds, fromNode)
}

// CFGBuilder maintains the context needed to construct the control-flow graph
type CFGBuilder struct {
	sem       *semantic.SemanticInfo
	nodeCount int
	allNodes  []*CFGNode
	exitNode  *CFGNode

	// Active loop contexts for resolving break and continue statements
	loopConds []*CFGNode
	loopExits []*CFGNode
}

func NewCFGBuilder(sem *semantic.SemanticInfo) *CFGBuilder {
	exit := &CFGNode{
		ID:      -1,
		Kind:    BlockExit,
		Preds:   make([]*CFGNode, 0),
		Succs:   make([]*CFGNode, 0),
		LiveIn:  make(map[*semantic.Symbol]bool),
		LiveOut: make(map[*semantic.Symbol]bool),
		Gen:     make(map[*semantic.Symbol]bool),
		Kill:    make(map[*semantic.Symbol]bool),
	}
	return &CFGBuilder{
		sem:       sem,
		nodeCount: 0,
		allNodes:  make([]*CFGNode, 0),
		exitNode:  exit,
		loopConds: make([]*CFGNode, 0),
		loopExits: make([]*CFGNode, 0),
	}
}

func (b *CFGBuilder) nextID() int {
	b.nodeCount++
	return b.nodeCount
}

func (b *CFGBuilder) createNode(kind CFGBlockKind, stmt ast.Statement, expr ast.Expression) *CFGNode {
	node := NewCFGNode(b.nextID(), kind, stmt, expr)
	b.allNodes = append(b.allNodes, node)
	b.populateGenKill(node)
	return node
}

// Build constructs the control flow graph for a block statement (e.g. function body)
func (b *CFGBuilder) Build(block *ast.BlockStatement) (*CFGNode, *CFGNode) {
	if block == nil || len(block.Statements) == 0 {
		return b.exitNode, b.exitNode
	}

	start, end := b.buildBlock(block)
	if end != nil {
		AddEdge(end, b.exitNode)
	}
	return start, b.exitNode
}

func (b *CFGBuilder) buildBlock(block *ast.BlockStatement) (start, end *CFGNode) {
	if block == nil || len(block.Statements) == 0 {
		return nil, nil
	}

	var firstNode *CFGNode
	var prevNode *CFGNode

	for _, stmt := range block.Statements {
		if stmt == nil {
			continue
		}

		s, e := b.buildStatement(stmt)
		if s == nil {
			continue
		}

		if firstNode == nil {
			firstNode = s
		}

		if prevNode != nil {
			AddEdge(prevNode, s)
		}

		prevNode = e
		// If prevNode is nil, it means control flow was interrupted (e.g., Return, Break, Continue)
		if prevNode == nil {
			break
		}
	}

	return firstNode, prevNode
}

func (b *CFGBuilder) buildStatement(stmt ast.Statement) (start, end *CFGNode) {
	switch s := stmt.(type) {
	case *ast.BlockStatement:
		return b.buildBlock(s)

	case *ast.ReturnStatement:
		node := b.createNode(BlockSimple, s, nil)
		AddEdge(node, b.exitNode)
		return node, nil // No sequential fallthrough

	case *ast.BreakStatement:
		node := b.createNode(BlockSimple, s, nil)
		if len(b.loopExits) > 0 {
			AddEdge(node, b.loopExits[len(b.loopExits)-1])
		} else {
			AddEdge(node, b.exitNode)
		}
		return node, nil

	case *ast.ContinueStatement:
		node := b.createNode(BlockSimple, s, nil)
		if len(b.loopConds) > 0 {
			AddEdge(node, b.loopConds[len(b.loopConds)-1])
		} else {
			AddEdge(node, b.exitNode)
		}
		return node, nil

	case *ast.WhileStatement:
		condNode := b.createNode(BlockLoopCondition, s, s.Condition)
		joinNode := b.createNode(BlockJoin, s, nil)

		// Push loop context
		b.loopConds = append(b.loopConds, condNode)
		b.loopExits = append(b.loopExits, joinNode)

		bodyStart, bodyEnd := b.buildBlock(s.Body)

		// Pop loop context
		b.loopConds = b.loopConds[:len(b.loopConds)-1]
		b.loopExits = b.loopExits[:len(b.loopExits)-1]

		// Edges:
		// 1. Condition leads to Loop body (true)
		if bodyStart != nil {
			AddEdge(condNode, bodyStart)
		}
		// 2. Condition leads to Join/Exit (false)
		AddEdge(condNode, joinNode)
		// 3. Body end loops back to Condition
		if bodyEnd != nil {
			AddEdge(bodyEnd, condNode)
		}

		return condNode, joinNode

	case *ast.ExpressionStatement:
		// If expressions can be blocks of code themselves. If it is an IfExpression or MatchExpression,
		// we treat it as structured control flow.
		if ifExpr, ok := s.Expression.(*ast.IfExpression); ok {
			return b.buildIfExpression(s, ifExpr)
		}
		if matchExpr, ok := s.Expression.(*ast.MatchExpression); ok {
			return b.buildMatchExpression(s, matchExpr)
		}
		node := b.createNode(BlockSimple, s, nil)
		return node, node

	default:
		node := b.createNode(BlockSimple, s, nil)
		return node, node
	}
}

func (b *CFGBuilder) buildExpressionBranch(expr ast.Expression) (start, end *CFGNode) {
	if expr == nil {
		return nil, nil
	}
	switch ex := expr.(type) {
	case *ast.BlockStatement:
		return b.buildBlock(ex)
	case *ast.IfExpression:
		// Wrap IfExpression in a dummy statement context or pass nil
		return b.buildIfExpression(nil, ex)
	case *ast.MatchExpression:
		return b.buildMatchExpression(nil, ex)
	default:
		// Fallback for simple expressions
		node := b.createNode(BlockSimple, nil, ex)
		return node, node
	}
}

func (b *CFGBuilder) buildIfExpression(stmt ast.Statement, ifExpr *ast.IfExpression) (start, end *CFGNode) {
	condNode := b.createNode(BlockBranchCondition, stmt, ifExpr.Condition)
	joinNode := b.createNode(BlockJoin, stmt, nil)

	thenStart, thenEnd := b.buildBlock(ifExpr.Consequence)
	var elseStart, elseEnd *CFGNode
	if ifExpr.Alternative != nil {
		elseStart, elseEnd = b.buildExpressionBranch(ifExpr.Alternative)
	}

	// Connect Condition to branches
	if thenStart != nil {
		AddEdge(condNode, thenStart)
	} else {
		AddEdge(condNode, joinNode)
	}

	if elseStart != nil {
		AddEdge(condNode, elseStart)
	} else {
		AddEdge(condNode, joinNode)
	}

	// Connect branches to Join
	if thenEnd != nil {
		AddEdge(thenEnd, joinNode)
	}
	if elseEnd != nil {
		AddEdge(elseEnd, joinNode)
	}

	return condNode, joinNode
}

func (b *CFGBuilder) buildMatchExpression(stmt ast.Statement, matchExpr *ast.MatchExpression) (start, end *CFGNode) {
	condNode := b.createNode(BlockBranchCondition, stmt, matchExpr.Target)
	joinNode := b.createNode(BlockJoin, stmt, nil)

	for _, cas := range matchExpr.Cases {
		caseStart, caseEnd := b.buildBlock(cas.Body)
		if caseStart != nil {
			AddEdge(condNode, caseStart)
		} else {
			AddEdge(condNode, joinNode)
		}

		if caseEnd != nil {
			AddEdge(caseEnd, joinNode)
		}
	}

	return condNode, joinNode
}

// populateGenKill inspects the statements / expressions to calculate Gen (Uses) and Kill (Defs) sets
func (b *CFGBuilder) populateGenKill(node *CFGNode) {
	if node == nil {
		return
	}

	var nodesToInspect []ast.Node
	if node.Expr != nil {
		nodesToInspect = append(nodesToInspect, node.Expr)
	} else if node.Stmt != nil {
		nodesToInspect = append(nodesToInspect, node.Stmt)
	}

	for _, root := range nodesToInspect {
		ast.Inspect(root, func(n ast.Node) bool {
			if id, ok := n.(*ast.Identifier); ok {
				// Defs (Kill)
				if sym := b.sem.Defs[id]; sym != nil {
					node.Kill[sym] = true
				}
				// Uses (Gen)
				if sym := b.sem.Uses[id]; sym != nil {
					// A symbol should not be in Gen if it is being defined in the same statement,
					// unless it is read on the RHS before definition (handled gracefully by standard liveness rules).
					if !node.Kill[sym] {
						node.Gen[sym] = true
					}
				}
			}
			return true
		})
	}
}
