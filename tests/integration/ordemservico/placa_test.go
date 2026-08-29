package integration_test

import "fmt"

func placaMercosul(prefixo string, suffix int64) string {
	return fmt.Sprintf("%s%d%c%02d", prefixo, (suffix/2600)%10, 'A'+rune(suffix%26), (suffix/26)%100)
}
