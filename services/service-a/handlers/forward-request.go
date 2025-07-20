package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"service-a/models"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
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

	carrier := propagation.HeaderCarrier(r.Header)
	context := r.Context()
	ctx := otel.GetTextMapPropagator().Extract(context, carrier)
	// Inicia o span para a requisição
	ctx, span := tracer.Start(ctx, "service-a-request")
	defer span.End() // Finaliza o span quando a função terminar

	// Decodifica o corpo da requisição

	// Envia o CEP para o Serviço B
	// Defina a URL do Serviço B no seu docker-compose ou em variável de ambiente
	serviceBURL := os.Getenv("SERVICE_B_URL") // "http://service-b:8081" por exemplo
	ctx, validateZipCodeSpan := tracer.Start(ctx, "validate-zip-code")

	var requestBody models.RequestBody
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil || !isValidCep(requestBody.Cep) {
		response := models.ErrorResponse{
			Error: "invalid zipcode", // Error message in English
		}
		// Set the HTTP status code to 422 (Unprocessable Entity)
		// Define o código de status HTTP como 422 (Entidade não processável)
		w.WriteHeader(http.StatusUnprocessableEntity)

		// Encode the response into JSON and send it to the client
		// Codifica a resposta em JSON e envia para o cliente
		json.NewEncoder(w).Encode(response)
		validateZipCodeSpan.SetStatus(codes.Error, "Invalid Zip Code Sent")
		validateZipCodeSpan.End()
		return
	}

	validateZipCodeSpan.SetStatus(codes.Ok, "Valid Zip Code Sent")
	validateZipCodeSpan.End()

	// Envia o CEP para o Serviço B via POST
	responseBody, statusCode, err := sendToServiceB(ctx, serviceBURL, requestBody.Cep, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		span.SetStatus(codes.Error, "Error calling Service B") // Marca erro no span
		return
	}

	if statusCode == http.StatusNotFound {
		w.Header().Set("Content-Type", "application/json")
		response := models.ErrorResponse{
			Error: "can not find zipcode", // Error message in English
		}

		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Retorna o status e o corpo de resposta do Serviço B
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode) // Status code de Serviço B
	json.NewEncoder(w).Encode(responseBody)

	span.SetStatus(codes.Ok, "Successfully Forwarded request")
}
func sendToServiceB(ctx context.Context, serviceBURL, cep string, r *http.Request) (body models.ResponseBody, status int, err error) {
	// Decodifica o corpo da resposta do Serviço B
	var responseBody models.ResponseBody
	// Cria o corpo da requisição para o Serviço B
	requestBody := models.RequestBody{Cep: cep}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return responseBody, 0, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Cria uma cópia dos headers
	reqHeaders := r.Header.Clone()

	// Propaga os headers, incluindo o RequestId e o trace context
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(reqHeaders))

	// Cria a requisição POST para o Serviço B com os cabeçalhos modificados
	req, err := http.NewRequestWithContext(ctx, "POST", serviceBURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return responseBody, 0, fmt.Errorf("failed to create request: %v", err)
	}

	// Adiciona os cabeçalhos copiados e modificados na requisição
	req.Header = reqHeaders

	// Envia a requisição para o Serviço B
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return responseBody, 0, fmt.Errorf("failed to call Service B: %v", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
		return responseBody, 0, fmt.Errorf("failed to decode response body: %v", err)
	}

	// Retorna a resposta do Serviço B com status e corpo
	return responseBody, resp.StatusCode, nil
}

func isValidCep(cep string) bool {
	match, _ := regexp.MatchString(`^\d{8}$`, cep) // Check if the CEP matches the regex pattern
	return match                                   // Return true if it matches, fals
}
