package main

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
)

func main() {
	fmt.Println("Starting goidents...")
	if len(os.Args) < 2 {
		panic(errors.New("No filename specified"))
	}

	fname := os.Args[1]
	err := process(fname)
	if err != nil {
		panic(err)
	}
}

var fset *token.FileSet

func process(fname string) error {
	fmt.Println("Processing: ", fname)

	fset = token.NewFileSet()

	f, err := parser.ParseFile(fset, fname, nil, 0)
	if err != nil {
		panic(err)
	}

	for _, decl := range f.Decls {
		switch decl.(type) {
		case *ast.FuncDecl:
			parseFn(decl.(*ast.FuncDecl))
		}
	}
	var buf bytes.Buffer
	err = format.Node(&buf, fset, f)
	if err != nil {
		return err
	}

	file, err := os.Create(fname)
	if err != nil {
		return err
	}

	fmt.Fprintf(file, "%s", buf.Bytes())
	return nil
}

func parseFn(fn *ast.FuncDecl) {
	varMap := make(map[string]bool)
	for _, stmt := range fn.Body.List {
		switch stmt.(type) {
		case *ast.DeclStmt:
			vars := parseDecl(stmt.(*ast.DeclStmt))
			appendAll(varMap, vars)

		case *ast.AssignStmt:
			assignStmt := stmt.(*ast.AssignStmt)
			idents := []*ast.Ident{}
			for _, expr := range assignStmt.Lhs {
				idents = append(idents, expr.(*ast.Ident))
			}

			vars := getVars(idents)
			if isAllRedeclared(varMap, vars) {
				changeTok(assignStmt)
			} else {
				appendAll(varMap, vars)
			}
		}
	}
}

func changeTok(stmt *ast.AssignStmt) {
	stmt.Tok = token.ASSIGN
}

func isAllRedeclared(varMap map[string]bool, vars []string) bool {
	for _, v := range vars {
		_, exist := varMap[v]
		if !exist {
			return false
		}
	}

	return true
}

func appendAll(m map[string]bool, keys []string) {
	for _, k := range keys {
		m[k] = true
	}
}

func parseDecl(decl *ast.DeclStmt) (vars []string) {
	switch decl.Decl.(type) {
	case *ast.GenDecl:
		for _, spec := range decl.Decl.(*ast.GenDecl).Specs {
			valSpec := spec.(*ast.ValueSpec)
			vars = append(vars, getVars(valSpec.Names)...)
		}
	}
	return
}

func getVars(idents []*ast.Ident) (vars []string) {
	for _, ident := range idents {
		vars = append(vars, ident.Name)
	}
	return
}
