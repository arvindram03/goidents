package goidents

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

var fset *token.FileSet

func Process(fname string) ([]byte, error) {

	fset = token.NewFileSet()

	f, err := parser.ParseFile(fset, fname, nil, 0)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return buf.Bytes(), nil
}

func parseFn(fn *ast.FuncDecl) {
	varMap := make(map[string]bool)
	if fn.Type.Results != nil {
		for _, field := range fn.Type.Results.List {
			vars := getVars(field.Names)
			appendAll(varMap, vars)
		}
	}

	for _, stmt := range fn.Body.List {
		switch stmt.(type) {
		case *ast.DeclStmt:
			vars := parseDecl(stmt.(*ast.DeclStmt))
			appendAll(varMap, vars)

		case *ast.AssignStmt:
			assignStmt := stmt.(*ast.AssignStmt)
			idents := []*ast.Ident{}
			for _, expr := range assignStmt.Lhs {
				if ident, ok := expr.(*ast.Ident); ok {
					idents = append(idents, ident)
				}
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
