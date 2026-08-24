package main

import (
	"log"
	"net/http"
	"os"

	segurancaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	veiculoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/veiculo"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	veiculoPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/veiculo"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

func main() {
	db, err := database.Open()
	if err != nil { log.Fatal(err) }
	defer db.Close()
	jwt, err := segurancaInfrastructure.NewJWT(os.Getenv("JWT_SECRET"))
	if err != nil { log.Fatal(err) }
	login := segurancaApplication.NewAutenticar(segurancaInfrastructure.NewPostgresRepository(db), jwt)
	cadastrar := veiculoApplication.NewCadastrar(veiculo.NewPostgresRepository(db))
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("POST /autenticacao/login", segurancaPresentation.NewLoginHandler(login))
	mux.Handle("POST /clientes/{clienteId}/veiculos", veiculoPresentation.NewHandler(cadastrar))
	log.Println("API iniciada na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func healthHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_, _ = writer.Write([]byte(`{"status":"ok"}`))
}
