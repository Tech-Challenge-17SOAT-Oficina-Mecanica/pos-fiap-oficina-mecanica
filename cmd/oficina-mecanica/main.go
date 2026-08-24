package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	fornecedorApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	segurancaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/seguranca"
	fornecedorInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/fornecedor"
	segurancaInfrastructure "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/infrastructure/seguranca"
	fornecedorPresentation "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/presentation/fornecedor"
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
	fornecedorRepository := fornecedorInfrastructure.NewPostgresRepository(db)
	login := segurancaApplication.NewAutenticar(segurancaInfrastructure.NewPostgresRepository(db), jwt)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.Handle("POST /fornecedores", fornecedorPresentation.NewCadastrarHandler(
		fornecedorApplication.NewCadastrar(fornecedorRepository),
	))
	mux.Handle("GET /fornecedores", segurancaPresentation.RequireScope(jwt, "compras:ler", fornecedorPresentation.NewListarHandler(
		fornecedorApplication.NewConsultarFornecedores(fornecedorRepository),
	)))
	mux.Handle("GET /fornecedores/{fornecedorId}", segurancaPresentation.RequireScope(jwt, "compras:ler", fornecedorPresentation.NewBuscarPorIDHandler(
		fornecedorApplication.NewConsultarFornecedorPorID(fornecedorRepository),
	)))
	mux.Handle("PUT /fornecedores/{fornecedorId}", segurancaPresentation.RequireScope(jwt, "compras:escrever", fornecedorPresentation.NewAtualizarHandler(
		fornecedorApplication.NewAtualizarFornecedor(fornecedorRepository),
	)))
	mux.Handle("DELETE /fornecedores/{fornecedorId}", segurancaPresentation.RequireScope(jwt, "compras:escrever", fornecedorPresentation.NewDesativarHandler(
		fornecedorApplication.NewDesativarFornecedor(fornecedorRepository),
	)))
	mux.Handle("POST /fornecedores/{fornecedorId}/reativacao", segurancaPresentation.RequireScope(jwt, "compras:escrever", fornecedorPresentation.NewReativarHandler(
		fornecedorApplication.NewReativarFornecedor(fornecedorRepository),
	)))
	mux.Handle("POST /autenticacao/login", segurancaPresentation.NewLoginHandler(login))

	server := &http.Server{
		Addr:    ":8080",
		Handler: corsMiddleware(mux),
	}

	log.Println("API iniciada na porta 8080")
	log.Fatal(server.ListenAndServe())
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if origin == "http://localhost:8081" || origin == "http://127.0.0.1:8081" {
			writer.Header().Set("Access-Control-Allow-Origin", origin)
			writer.Header().Set("Vary", "Origin")
			writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization, If-Match")
		}
		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(writer, request)
	})
}

func healthHandler(writer http.ResponseWriter, _ *http.Request) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(writer).Encode(map[string]string{"status": "ok"})
}
