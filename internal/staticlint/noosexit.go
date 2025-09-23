package staticlint

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// NoOsExitAnalyzer checks that os.Exit is not used in main function of main package
var NoOsExitAnalyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      "forbids usage of os.Exit in main function of main package",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runNoOsExit,
}

func runNoOsExit(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		pos := pass.Fset.Position(file.Pos())
		if strings.Contains(pos.Filename, "/.cache/go-build/") {
			return nil, nil
		}
	}

	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	inspct := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Filter for function declarations
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspct.Preorder(nodeFilter, func(n ast.Node) {
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok {
			return
		}

		// Check if this is the main function (no receiver, name is "main")
		if funcDecl.Name.Name != "main" || funcDecl.Recv != nil {
			return
		}

		// Inspect the function body for os.Exit calls
		ast.Inspect(funcDecl.Body, func(node ast.Node) bool {
			callExpr, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			if isOsExitCall(callExpr) {
				pass.Reportf(callExpr.Pos(),
					"os.Exit usage in main function is forbidden. Use return instead of os.Exit")
			}

			return true
		})
	})

	return nil, nil
}

// isOsExitCall checks if the call expression is os.Exit()
func isOsExitCall(callExpr *ast.CallExpr) bool {
	selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	ident, ok := selExpr.X.(*ast.Ident)
	if !ok {
		return false
	}

	return ident.Name == "os" && selExpr.Sel.Name == "Exit"
}
