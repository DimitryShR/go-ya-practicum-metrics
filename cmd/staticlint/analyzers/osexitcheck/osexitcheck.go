// Package osexitcheck provides a static analyzer that checks for direct calls
// to os.Exit in the main function of the main package.
//
// The analyzer inspects Go AST to find function declarations named "main"
// in packages named "main" and reports any direct call to os.Exit within
// the body of that function.
//
// Usage:
//
//	mycheck ./...
package osexitcheck

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer defines the osexitcheck analyzer.
// It checks that os.Exit is not called directly in the main function
// of the main package, which can lead to issues with deferred functions
// not being executed.
var Analyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  "check for os.Exit calls in main function of main package",
	Run:  run,
}

// run performs the analysis on each package.
// It skips packages that are not named "main".
func run(pass *analysis.Pass) (interface{}, error) {
	// Only analyze packages named "main"
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			// Look for function declarations
			fn, ok := n.(*ast.FuncDecl)
			if !ok {
				return true
			}

			// Only check functions named "main"
			if fn.Name.Name != "main" {
				return true
			}

			// Inspect the body of the main function
			ast.Inspect(fn.Body, func(n2 ast.Node) bool {
				call, ok := n2.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				pkgIdent, ok := sel.X.(*ast.Ident)
				if !ok || pkgIdent.Name != "os" || sel.Sel.Name != "Exit" {
					return true
				}

				pass.Reportf(call.Pos(), "os.Exit in main function of main package is prohibited")
				return true
			})

			// Don't inspect nested functions - only top-level main
			return false
		})
	}

	return nil, nil
}
