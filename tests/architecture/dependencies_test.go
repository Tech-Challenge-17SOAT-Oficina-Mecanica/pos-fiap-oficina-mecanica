package architecture_test

import (
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// Verifica o codigo de producao sem importar nem inicializar os adaptadores.
func TestDependenciasEntreCamadas(t *testing.T) {
	root := filepath.Join("..", "..")
	mod, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	fields := strings.Fields(string(mod))
	if len(fields) < 2 || fields[0] != "module" {
		t.Fatal("modulo Go nao identificado")
	}
	prefix := fields[1] + "/internal/"
	scanned := map[string]int{}
	err = filepath.WalkDir(filepath.Join(root, "internal"), func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		rel, err := filepath.Rel(filepath.Join(root, "internal"), path)
		if err != nil {
			return err
		}
		layer := strings.Split(filepath.ToSlash(rel), "/")[0]
		if layer == "shared" && strings.HasPrefix(filepath.ToSlash(rel), "shared/validation/") {
			layer = "validation"
		}
		if layer == "shared" {
			return nil
		}
		scanned[layer]++
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, spec := range file.Imports {
			imported, err := strconv.Unquote(spec.Path.Value)
			if err != nil {
				return err
			}
			if !permitido(layer, imported, prefix) {
				t.Errorf("%s: camada %s nao pode depender de %s", filepath.ToSlash(rel), layer, imported)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, layer := range []string{"domain", "application", "presentation", "infrastructure", "validation"} {
		if scanned[layer] == 0 {
			t.Errorf("nenhum arquivo verificado na camada %s", layer)
		}
	}
	t.Logf("arquivos de producao verificados por camada: %v", scanned)
}

func permitido(layer, imported, prefix string) bool {
	if strings.HasPrefix(imported, prefix) {
		target := strings.TrimPrefix(imported, prefix)
		allowed := map[string][]string{
			"domain":         {"domain", "shared/validation"},
			"application":    {"application", "domain", "shared/validation"},
			"presentation":   {"presentation", "application", "domain", "shared/http", "shared/validation"},
			"infrastructure": {"infrastructure", "application", "domain", "shared/database", "shared/validation"},
			"validation":     {},
		}
		for _, path := range allowed[layer] {
			if target == path || strings.HasPrefix(target, path+"/") {
				return true
			}
		}
		return false
	}
	// net/mail apenas interpreta sintaxe; nao realiza conexoes nem envio.
	if layer == "validation" && imported == "net/mail" {
		return true
	}
	core := layer == "domain" || layer == "application" || layer == "validation"
	if core {
		// Drivers e bibliotecas externas pertencem aos adaptadores.
		if strings.Contains(strings.Split(imported, "/")[0], ".") {
			return false
		}
		for _, externalDetail := range []string{"net", "database", "os", "io", "syscall", "unsafe", "plugin"} {
			if imported == externalDetail || strings.HasPrefix(imported, externalDetail+"/") {
				return false
			}
		}
	}
	// A apresentacao nao acessa banco ou drivers; usa os casos de uso.
	if layer == "presentation" {
		if strings.Contains(strings.Split(imported, "/")[0], ".") || imported == "database/sql" || strings.HasPrefix(imported, "database/") {
			return false
		}
	}
	return true
}
