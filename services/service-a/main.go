package main

import (
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

	// Inicializa o Tracer
	shutdown := helpers.InitTracer(serviceName)
	defer shutdown()

	// Configura o handler para a rota POST /
	http.HandleFunc("/", handlers.ForwardRequest)

	// Inicia o servidor HTTP na porta 8080
	port := ":8080"
	fmt.Printf("Serviço A rodando na porta %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
