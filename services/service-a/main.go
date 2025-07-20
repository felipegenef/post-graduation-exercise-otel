package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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

	// Configura o handler para a rota POST /
	http.HandleFunc("/", handlers.ForwardRequest)

	// Inicia o servidor HTTP na porta 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not provided
	}

	fmt.Printf("Serviço A rodando na porta %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
