package http

import (
	"errors"
	"testing"
)

func TestLerIfMatch(t *testing.T) {
	casos := []struct {
		header   string
		esperado int
		erro     error
	}{
		{"3", 3, nil},
		{`"3"`, 3, nil}, // forma correta de ETag por HTTP
		{"  7  ", 7, nil},
		{`  "12"  `, 12, nil},
		{"", 0, ErrIfMatchAusente},
		{"   ", 0, ErrIfMatchAusente},
		{"abc", 0, ErrIfMatchInvalido},
		{`"abc"`, 0, ErrIfMatchInvalido},
		{"0", 0, ErrIfMatchInvalido},
		{"-1", 0, ErrIfMatchInvalido},
		{"*", 0, ErrIfMatchInvalido}, // curinga anularia o controle otimista
		{"3.5", 0, ErrIfMatchInvalido},
	}

	for _, caso := range casos {
		t.Run("header="+caso.header, func(t *testing.T) {
			version, err := LerIfMatch(caso.header)
			if !errors.Is(err, caso.erro) {
				t.Fatalf("erro = %v, esperado %v", err, caso.erro)
			}
			if version != caso.esperado {
				t.Fatalf("version = %d, esperado %d", version, caso.esperado)
			}
			if caso.erro != nil && CampoDoErro(err) != "If-Match" {
				t.Fatalf("erro deveria apontar o campo If-Match, veio %q", CampoDoErro(err))
			}
		})
	}
}
