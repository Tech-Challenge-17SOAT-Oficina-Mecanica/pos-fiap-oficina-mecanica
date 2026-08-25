package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	pecaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	segurancaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	pecaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/peca"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	pecaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/peca"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

const escopoEstoqueLer = "estoque:ler"

func main() {
	ctx := context.Background()
	db, err := database.Open(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	if err := db.Ping(ctx); err != nil {
		log.Fatal(err)
	}
	jwt, err := segurancaInfrastructure.NewJWT(os.Getenv("JWT_SECRET"))
	if err != nil {
		log.Fatal(err)
	}
	login := segurancaApplication.NewAutenticar(segurancaInfrastructure.NewPostgresRepository(db), jwt)
	consultarPecas := pecaApplication.NewConsultarPecas(pecaInfrastructure.NewPostgresRepository(db))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("POST /autenticacao/login", segurancaPresentation.NewLoginHandler(login))
	mux.Handle("GET /estoque/pecas", segurancaPresentation.ComEscopo(jwt, escopoEstoqueLer,
		pecaPresentation.NewConsultarPecasHandler(consultarPecas)))
	mux.Handle("GET /estoque/pecas/{pecaId}", segurancaPresentation.ComEscopo(jwt, escopoEstoqueLer,
		pecaPresentation.NewConsultarPecaPorIDHandler(consultarPecas)))

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	log.Println("API iniciada na porta 8080")
	log.Fatal(server.ListenAndServe())
}

func healthHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(map[string]string{"status": "ok"})
}
