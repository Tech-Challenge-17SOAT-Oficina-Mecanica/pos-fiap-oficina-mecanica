package veiculo

import (
	"errors"
	"regexp"
	"strings"
	"time"
)

var placa = regexp.MustCompile(`^(?:[A-Z]{3}[0-9]{4}|[A-Z]{3}[0-9][A-Z][0-9]{2})$`)

type Cadastro struct {
	Placa  string `json:"placa"`
	Marca  string `json:"marca"`
	Modelo string `json:"modelo"`
	Ano    int    `json:"ano"`
}

type Veiculo struct {
	ID        string `json:"id"`
	ClienteID string `json:"-"`
	Cadastro
	Ativo   bool    `json:"ativo"`
	Version int     `json:"version"`
	Cliente Cliente `json:"cliente"`
}

type Cliente struct {
	ID        string `json:"id"`
	Nome      string `json:"nome"`
	Documento string `json:"documento"`
}

func NormalizarPlaca(valor string) (string, error) {
	normalizada := strings.ToUpper(strings.NewReplacer("-", "", " ", "").Replace(strings.TrimSpace(valor)))
	if !placa.MatchString(normalizada) {
		return "", errors.New("placa inválida")
	}
	return normalizada, nil
}

func NovoCadastro(placaVeiculo, marca, modelo string, ano int) (Cadastro, error) {
	pl, err := NormalizarPlaca(placaVeiculo)
	if err != nil {
		return Cadastro{}, err
	}
	cadastro := Cadastro{pl, strings.TrimSpace(marca), strings.TrimSpace(modelo), ano}
	if cadastro.Marca == "" || cadastro.Modelo == "" {
		return Cadastro{}, errors.New("marca e modelo são obrigatórios")
	}
	if cadastro.Ano < 1900 || cadastro.Ano > time.Now().Year()+1 {
		return Cadastro{}, errors.New("ano inválido")
	}
	return cadastro, nil
}
