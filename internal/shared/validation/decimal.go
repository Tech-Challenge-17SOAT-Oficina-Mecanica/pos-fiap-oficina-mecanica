package validation

import (
	"regexp"
	"strconv"
	"strings"
)

func DecimalPositivo(valor string, casasDecimais int) bool {
	return decimal(valor, casasDecimais) && !zero(valor)
}

func DecimalNaoNegativo(valor string, casasDecimais int) bool {
	return decimal(valor, casasDecimais)
}

func decimal(valor string, casasDecimais int) bool {
	if casasDecimais < 0 {
		return false
	}
	padrao := `^(0|[1-9][0-9]*)$`
	if casasDecimais > 0 {
		padrao = `^(0|[1-9][0-9]*)(\.[0-9]{1,` + strconv.Itoa(casasDecimais) + `})?$`
	}
	return regexp.MustCompile(padrao).MatchString(strings.TrimSpace(valor))
}

func zero(valor string) bool {
	numero, err := strconv.ParseFloat(strings.TrimSpace(valor), 64)
	return err == nil && numero == 0
}
