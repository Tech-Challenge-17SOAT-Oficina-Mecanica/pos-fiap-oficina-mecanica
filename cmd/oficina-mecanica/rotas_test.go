package main

import (
	"os"
	"regexp"
	"testing"
)

// O net/http entra em panico ao registrar a mesma rota duas vezes, e isso so acontece
// quando a aplicacao sobe — build e vet passam normalmente. Este teste pega a duplicata
// antes, lendo o proprio main.go.
func TestNenhumaRotaRegistradaDuasVezes(t *testing.T) {
	fonte, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatal(err)
	}

	padrao := regexp.MustCompile(`mux\.Handle(?:Func)?\("([A-Z]+ [^"]+)"`)
	vistas := map[string]int{}
	for _, achado := range padrao.FindAllStringSubmatch(string(fonte), -1) {
		vistas[achado[1]]++
	}

	if len(vistas) == 0 {
		t.Fatal("nenhuma rota encontrada; o teste precisa ser ajustado ao formato do main.go")
	}
	for rota, quantidade := range vistas {
		if quantidade > 1 {
			t.Errorf("rota %q registrada %d vezes; o mux entra em pânico ao subir", rota, quantidade)
		}
	}
}
