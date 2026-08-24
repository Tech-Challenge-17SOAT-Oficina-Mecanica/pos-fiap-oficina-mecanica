package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	segurancaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	segurancaPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/seguranca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/shared/database"
)

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
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("POST /autenticacao/login", segurancaPresentation.NewLoginHandler(login))

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
