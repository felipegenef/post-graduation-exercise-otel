package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	handlers "service-b/handlers"
	"service-b/helpers"
	"service-b/models"
	"service-b/services"
	"service-b/shared"
)

// getHandler initializes and returns a new instance of WeatherHandler.
// Inicializa e retorna uma nova instância de WeatherHandler.
func getHandler() *handlers.WeatherHandler {
	// Create channels for receiving location data from APIs
	// Cria canais para receber dados de localização das APIs
	chBrasilAPI := make(chan models.Location)
	chViaCEP := make(chan models.Location)

	// Create an HTTP client
	// Cria um cliente HTTP
	client := &http.Client{}

	// Initialize temperature converter
	// Inicializa o conversor de temperatura
	temperatureConverter := &shared.TemperatureConverter{}

	// Initialize API client with the HTTP client
	// Inicializa o cliente da API com o cliente HTTP
	apiClient := &services.APIClientImpl{Client: client}

	// Create a new instance of WeatherService with the API client
	// Cria uma nova instância do WeatherService com o cliente da API
	weatherService := services.NewWeatherService(apiClient)

	// Initialize LocationService which depends on WeatherService
	// Inicializa o LocationService, que depende do WeatherService
	locationService := services.NewLocationService(weatherService)

	// Initialize and return WeatherHandler with the necessary services and channels
	// Inicializa e retorna o WeatherHandler com os serviços e canais necessários
	handler := handlers.NewWeatherHandler(
		locationService,
		weatherService,
		temperatureConverter,
		chBrasilAPI,
		chViaCEP,
	)
	return handler
}

// main function that starts the HTTP server
// Função main que inicia o servidor HTTP
func main() {

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

	// Get the weather handler to handle incoming weather-related requests
	// Obtém o handler de clima para lidar com requisições relacionadas ao clima
	weatherHandler := getHandler()

	// Define the route for weather data, and associate it with the WeatherHandler
	// Define a rota para os dados do clima e associa com o WeatherHandler
	http.HandleFunc("/", weatherHandler.WeatherHandlerFunc())

	// Get the port number from environment variable, default to "8080" if not set
	// Obtém o número da porta da variável de ambiente, padrão para "8080" se não estiver definida
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Default port if not provided
	}

	// Log the port the server is running on
	// Registra o número da porta em que o servidor está rodando
	log.Printf("Server running on port %s", port)

	// Start the HTTP server and log any fatal errors
	// Inicia o servidor HTTP e registra qualquer erro fatal
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
