package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"service-a/handlers"
	helpers "service-a/helpers" // Importando o InitTracer de Helpers/otel.go
)

func main() {
	// Recupera o nome do serviço da variável de ambiente OTEL_SERVICE_NAME
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "service-a" // Se não houver, usa o nome "service-a"
	}

	otelEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if otelEndpoint == "" {
		otelEndpoint = "otel-collector:4317"
	}

	// Inicializa o Tracer
	shutdown, err := helpers.InitTracer(serviceName, otelEndpoint)
	if err != nil {
		fmt.Errorf("error initializing tracer %w", err)
		// panic("error initializing tracer")
	}
	defer shutdown(context.Background())

	// Cria o roteador Chi
	r := chi.NewRouter()

	// Adiciona os middlewares do Chi
	r.Use(middleware.RequestID) // Middleware para RequestID
	r.Use(middleware.RealIP)    // Middleware para pegar o IP real
	r.Use(middleware.Recoverer) // Middleware para recuperação de panics
	r.Use(middleware.Logger)    // Middleware para logging das requisições

	// Configura o handler para a rota POST /
	r.Post("/", handlers.ForwardRequest)

	// Inicia o servidor HTTP na porta 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not provided
	}

	fmt.Printf("Serviço A rodando na porta %s...\n", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
