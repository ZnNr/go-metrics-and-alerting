package staticlint

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/analysis"
)

var ExitFromMainAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for os.Exit from main functions of package main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		// только пакеты main
		if file.Name.Name != "main" {
			continue
		}
		for _, decl := range file.Decls {
			// только функции main
			if funcName, ok := decl.(*ast.FuncDecl); ok && funcName.Name.Name == "main" {
				ast.Inspect(decl, func(n ast.Node) bool {
					// только вызовы функций
					if c, ok := n.(*ast.CallExpr); ok {
						if s, ok := c.Fun.(*ast.SelectorExpr); ok {
							// только функции Exit пакета os
							if s.Sel.Name == "Exit" && fmt.Sprintf("%s", s.X) == "os" {
								pass.Reportf(s.Pos(), "os.Exit from main function of main packages is denied")
							}
						}
					}
					return true
				})
			}
		}
	}
	return nil, nil
}
