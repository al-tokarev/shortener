package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

// ErrExitAnalyzer — анализатор, который запрещает использовать os.Exit в функции main пакета main.
var ErrExitAnalyzer = &analysis.Analyzer{
	Name: "errexitcheck",
	Doc:  "check use Exit function in main()",
	Run:  CheckExitInMain,
}

// CheckExitInMain проверяет, что в функции main пакета main нет вызова os.Exit.
func CheckExitInMain(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			funcDecl, ok := node.(*ast.FuncDecl)
			if !ok {
				return true
			}

			if funcDecl.Name.Name != "main" {
				return true
			}

			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				callExpr, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				selExpr, ok := callExpr.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := selExpr.X.(*ast.Ident)
				if !ok {
					return true
				}

				obj, ok := pass.TypesInfo.Uses[ident]
				if !ok {
					return true
				}

				pkgName, ok := obj.(*types.PkgName)
				if !ok {
					return true
				}

				if pkgName.Imported().Path() == "os" && selExpr.Sel.Name == "Exit" {
					pass.Reportf(callExpr.Pos(), "os.Exit is not allowed in main function main package")
				}

				return true
			})

			return false
		})
	}

	return nil, nil
}
