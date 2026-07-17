package semantic

import (
	"strings"

	"github.com/nora-language/nora/pkg/parser/ast"
)

// FilterCfg filters out AST nodes that do not match the current build configuration.
func FilterCfg(program *ast.Program, targetOS string, targetArch string, targetFeatures []string) {
	for _, file := range program.Files {
		FilterFileCfg(file, targetOS, targetArch, targetFeatures)
	}
}

// FilterFileCfg filters out AST nodes for a single file.
func FilterFileCfg(file *ast.File, targetOS string, targetArch string, targetFeatures []string) {
	var filtered []ast.Statement
	for _, stmt := range file.Statements {
		if evaluateStmtCfg(stmt, targetOS, targetArch, targetFeatures) {
			filtered = append(filtered, stmt)
		}
	}
	file.Statements = filtered
}

func evaluateStmtCfg(stmt ast.Statement, targetOS string, targetArch string, targetFeatures []string) bool {
	var attributes []ast.Attribute
	switch n := stmt.(type) {
	case *ast.FunctionStatement:
		attributes = n.Attributes
	case *ast.TypeStatement:
		attributes = n.Attributes
	case *ast.ExportStatement:
		if stmt, ok := n.Node.(ast.Statement); ok {
			return evaluateStmtCfg(stmt, targetOS, targetArch, targetFeatures)
		}
		return true
	}

	cfgAttr := ast.GetAttribute(attributes, "cfg")
	if cfgAttr == nil {
		return true // Include by default
	}

	return evaluateCfgArgs(cfgAttr.Args, targetOS, targetArch, targetFeatures)
}

func evaluateCfgArgs(args []string, targetOS string, targetArch string, targetFeatures []string) bool {
	if len(args) == 0 {
		return true
	}

	// For now, handle simple key=value pairs or "not(key=value)" passed as strings.
	// E.g., [cfg("target_feature=avx")] or [cfg("target_arch=x86_64")] or [cfg("not(target_feature=avx)")]
	
	condition := args[0]
	isNot := strings.HasPrefix(condition, "not(") && strings.HasSuffix(condition, ")")
	if isNot {
		condition = condition[4 : len(condition)-1]
	}

	parts := strings.SplitN(condition, "=", 2)
	if len(parts) != 2 {
		return true // invalid cfg format, just include it
	}

	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])

	match := false
	switch key {
	case "target_os", "os":
		match = targetOS == val
	case "target_arch", "arch":
		match = targetArch == val
	case "target_feature":
		for _, f := range targetFeatures {
			if f == val {
				match = true
				break
			}
		}
	}

	if isNot {
		return !match
	}
	return match
}
