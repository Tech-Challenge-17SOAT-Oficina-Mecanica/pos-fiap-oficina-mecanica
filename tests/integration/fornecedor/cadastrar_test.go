package integration_test

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	infrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/fornecedor"
	presentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/fornecedor"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func TestCadastrarFornecedor(t *testing.T) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL nao configurada")
	}

	ctx := context.Background()
	db, err := database.OpenPool()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	documento := cnpjValido(fmt.Sprintf("%012d", time.Now().UnixNano()%1000000000000))
	mux := http.NewServeMux()
	mux.Handle("POST /fornecedores", presentation.NewCadastrarHandler(
		application.NewCadastrar(infrastructure.NewPostgresRepository(db)),
	))
	request := func() *httptest.ResponseRecorder {
		body := fmt.Sprintf(`{"razaoSocial":"Fornecedor Teste Ltda","documento":"%s","tipoDocumento":"CNPJ","email":"teste@example.com"}`, documento)
		recorder := httptest.NewRecorder()
		mux.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/fornecedores", bytes.NewBufferString(body)))
		return recorder
	}

	firstResponse := request()
	if firstResponse.Code != http.StatusCreated {
		t.Fatalf("status=%d body=%s", firstResponse.Code, firstResponse.Body.String())
	}
	t.Cleanup(func() {
		_, _ = db.Exec(ctx, "DELETE FROM fornecedor WHERE documento = $1", documento)
	})
	if duplicateResponse := request(); duplicateResponse.Code != http.StatusConflict {
		t.Fatalf("status duplicado=%d body=%s", duplicateResponse.Code, duplicateResponse.Body.String())
	}
}

func cnpjValido(base string) string {
	firstDigit := digitoCNPJ(base, []int{5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	secondDigit := digitoCNPJ(base+string(rune('0'+firstDigit)), []int{6, 5, 4, 3, 2, 9, 8, 7, 6, 5, 4, 3, 2})
	return base + string(rune('0'+firstDigit)) + string(rune('0'+secondDigit))
}

func digitoCNPJ(value string, weights []int) int {
	sum := 0
	for index, digit := range value {
		sum += int(digit-'0') * weights[index]
	}
	remainder := sum % 11
	if remainder < 2 {
		return 0
	}
	return 11 - remainder
}
