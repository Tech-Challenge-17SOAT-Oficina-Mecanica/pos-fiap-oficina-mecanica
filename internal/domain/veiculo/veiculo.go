package veiculo

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var placa = regexp.MustCompile(`^(?:[A-Z]{3}[0-9]{4}|[A-Z]{3}[0-9][A-Z][0-9]{2})$`)

type Cadastro struct {
	Placa, Marca, Modelo string
	Ano                  int
}
type Veiculo struct {
	ID, ClienteID string
	Cadastro
	Ativo bool
}

func NovoCadastro(placaVeiculo, marca, modelo string, ano int) (Cadastro, error) {
	cadastro := Cadastro{strings.ToUpper(strings.NewReplacer("-", "", " ", "").Replace(strings.TrimSpace(placaVeiculo))), strings.TrimSpace(marca), strings.TrimSpace(modelo), ano}
	if !placa.MatchString(cadastro.Placa) {
		return Cadastro{}, errors.New("placa inválida")
	}
	if cadastro.Marca == "" || cadastro.Modelo == "" {
		return Cadastro{}, errors.New("marca e modelo são obrigatórios")
	}
	if cadastro.Ano < 1900 || cadastro.Ano > time.Now().Year()+1 {
		return Cadastro{}, errors.New("ano inválido")
	}
	return cadastro, nil
}
