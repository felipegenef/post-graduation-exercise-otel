package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"service-a/models"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer trace.Tracer

// ForwardRequest lida com a requisição POST do Serviço A
func ForwardRequest(w http.ResponseWriter, r *http.Request) {
	// Recupera o nome do serviço da variável de ambiente OTEL_SERVICE_NAME
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "service-a" // Se não houver, usa o nome "service-a"
	}
	// Define o tracer para ser usado nas requisições
	tracer = otel.Tracer(serviceName)
	// Inicia o span para a requisição
	ctx, span := tracer.Start(r.Context(), "forward-request-to-service-b")
	defer span.End() // Finaliza o span quando a função terminar

	// Decodifica o corpo da requisição
	var requestBody models.RequestBody
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		span.SetStatus(codes.Error, "Invalid request body") // Marca erro no span
		return
	}

	// Envia o CEP para o Serviço B
	// Defina a URL do Serviço B no seu docker-compose ou em variável de ambiente
	serviceBURL := os.Getenv("SERVICE_B_URL") // "http://service-b:8081" por exemplo

	// Envia o CEP para o Serviço B via POST
	serviceBResp, err := sendToServiceB(ctx, serviceBURL, requestBody.Cep)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		span.SetStatus(codes.Error, "Error calling Service B") // Marca erro no span
		return
	}

	// Retorna o status e o corpo de resposta do Serviço B
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(serviceBResp.StatusCode)       // Status code de Serviço B
	json.NewEncoder(w).Encode(serviceBResp.Body) // Corpo da resposta de Serviço B

	span.SetStatus(codes.Ok, "Successfully Forwarded request")
}

// sendToServiceB envia a requisição para o Serviço B
func sendToServiceB(ctx context.Context, serviceBURL, cep string) (*http.Response, error) {
	// Cria o corpo da requisição para o Serviço B
	requestBody := models.RequestBody{Cep: cep}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Cria a requisição POST para o Serviço B
	req, err := http.NewRequestWithContext(ctx, "POST", serviceBURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Envia a requisição para o Serviço B
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Service B: %v", err)
	}
	defer resp.Body.Close()

	// Decodifica o corpo da resposta do Serviço B
	var responseBody map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("failed to decode response body: %v", err)
	}

	// Retorna a resposta do Serviço B com status e corpo
	// Usando io.NopCloser para garantir que o corpo seja um io.ReadCloser
	return &http.Response{
		StatusCode: resp.StatusCode,
		Body:       io.NopCloser(bytes.NewBuffer([]byte(fmt.Sprintf("%v", responseBody)))),
	}, nil
}
