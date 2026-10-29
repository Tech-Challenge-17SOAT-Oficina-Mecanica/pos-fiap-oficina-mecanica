package architecture_test

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

// Impede que novos handlers voltem a receber structs concretas de casos de uso.
func TestHandlersDependemDeInterfaces(t *testing.T) {
	root := filepath.Join("..", "..", "internal", "presentation")
	entries, err := filepath.Glob(filepath.Join(root, "*"))
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, dir := range entries {
		packages, err := parser.ParseDir(token.NewFileSet(), dir, func(info fs.FileInfo) bool {
			return !strings.HasSuffix(info.Name(), "_test.go")
		}, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, pkg := range packages {
			interfaces := map[string]bool{}
			for _, file := range pkg.Files {
				for _, declaration := range file.Decls {
					group, ok := declaration.(*ast.GenDecl)
					if !ok {
						continue
					}
					for _, spec := range group.Specs {
						if named, ok := spec.(*ast.TypeSpec); ok {
							_, interfaces[named.Name.Name] = named.Type.(*ast.InterfaceType)
						}
					}
				}
			}
			for path, file := range pkg.Files {
				for _, declaration := range file.Decls {
					constructor, ok := declaration.(*ast.FuncDecl)
					if !ok || !strings.HasPrefix(constructor.Name.Name, "New") || !strings.HasSuffix(constructor.Name.Name, "Handler") {
						continue
					}
					for _, param := range constructor.Type.Params.List {
						checked++
						contract, ok := param.Type.(*ast.Ident)
						if !ok || !interfaces[contract.Name] {
							t.Errorf("%s: %s deve receber uma interface local de caso de uso", path, constructor.Name.Name)
						}
					}
				}
			}
		}
	}
	if checked == 0 {
		t.Fatal("nenhum contrato de handler foi verificado")
	}
	t.Logf("%d construtores de handlers verificados", checked)
}
