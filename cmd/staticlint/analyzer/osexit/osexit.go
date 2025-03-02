// Модуль osexit реализует поиск вызова функции os.Exit
// Критерий для поиска - функция main в пакете main.
// Результат поиска - запрет использования прямого вызова os.Exit
package osexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer описывает функцию анализа
var Analyzer = &analysis.Analyzer{
	Name: "osExitCheck",
	Doc:  "reports usage of osExit in main functions of main packages",
	Run:  run,
}

// run запускает ast анализ исходного кода
func run(pass *analysis.Pass) (interface{}, error) {
	funcDecl := func(x *ast.FuncDecl) {
		for _, stmt := range x.Body.List {
			if exprStmt, ok := stmt.(*ast.ExprStmt); ok {
				callExpr := exprStmt.X.(*ast.CallExpr)
				seExpr := callExpr.Fun.(*ast.SelectorExpr)
				ident := seExpr.X.(*ast.Ident)
				if ident.Name == "os" && seExpr.Sel.Name == "Exit" {
					pass.Report(analysis.Diagnostic{
						Pos:     callExpr.Pos(),
						Message: "os.Exit calls not allowed",
					})
				}
			}
		}
	}

	for _, file := range pass.Files {
		if file.Name.Name == "main" {
			ast.Inspect(file, func(node ast.Node) bool {
				switch n := node.(type) {
				case *ast.FuncDecl:
					if n.Name.Name == "main" {
						funcDecl(n)
					}
				}
				return true
			})
		}
	}

	return nil, nil
}
