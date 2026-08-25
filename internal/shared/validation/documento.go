package validation

import "strings"

func OnlyDigits(value string) string {
	return strings.Map(func(r rune) rune {
		if r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, value)
}

func IsDocumento(documento, tipo string) bool {
	switch strings.ToUpper(strings.TrimSpace(tipo)) {
	case "CPF":
		return IsCPF(documento)
	case "CNPJ":
		return IsCNPJ(documento)
	default:
		return false
	}
}

func IsCPF(cpf string) bool {
	if len(cpf) != 11 || todosIguais(cpf) {
		return false
	}
	return digitoCPF(cpf, 9) == int(cpf[9]-'0') && digitoCPF(cpf, 10) == int(cpf[10]-'0')
}

func IsCNPJ(cnpj string) bool {
	if len(cnpj) != 14 || todosIguais(cnpj) {
		return false
	}
	return digitoCNPJ(cnpj, []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}) == int(cnpj[12]-'0') &&
		digitoCNPJ(cnpj, []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2}) == int(cnpj[13]-'0')
}

func digitoCPF(cpf string, pos int) int {
	sum := 0
	for i := 0; i < pos; i++ {
		sum += int(cpf[i]-'0') * (pos + 1 - i)
	}
	digito := (sum * 10) % 11
	if digito == 10 {
		return 0
	}
	return digito
}

func digitoCNPJ(cnpj string, weights []int) int {
	sum := 0
	for i, weight := range weights {
		sum += int(cnpj[i]-'0') * weight
	}
	resto := sum % 11
	if resto < 2 {
		return 0
	}
	return 11 - resto
}

func todosIguais(value string) bool {
	for i := 1; i < len(value); i++ {
		if value[i] != value[0] {
			return false
		}
	}
	return true
}
