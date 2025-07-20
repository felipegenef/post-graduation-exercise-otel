package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"service-b/models"
	"service-b/services"
	"service-b/shared"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Define interfaces for services that can be injected

// WeatherHandler is responsible for handling weather-related requests and managing dependencies.
type WeatherHandler struct {
	LocationService      services.LocationService     // Service to retrieve location data
	WeatherService       services.WeatherService      // Service to retrieve weather data
	CepValidator         *shared.CepValidator         // Validator for validating CEP (Brazilian ZIP code)
	TemperatureConverter *shared.TemperatureConverter // Utility to convert temperatures between Celsius, Fahrenheit, and Kelvin
}

// NewWeatherHandler creates and returns a new WeatherHandler with everything initialized
// Nova instância do WeatherHandler é criada e retornada com todos os serviços e utilitários inicializados
func NewWeatherHandler(
	locationService services.LocationService,
	weatherService services.WeatherService,
	temperatureConverter *shared.TemperatureConverter,
	chBrasilAPI, chViaCEP chan models.Location,
) *WeatherHandler {
	// Initialize channels for fetching location data
	// Inicializa os canais para buscar dados de localização

	return &WeatherHandler{
		LocationService:      locationService,                   // Assign location service
		WeatherService:       weatherService,                    // Assign weather service
		CepValidator:         shared.NewCepValidator(`^\d{8}$`), // Assign CEP validator with a regex pattern
		TemperatureConverter: temperatureConverter,              // Assign temperature converter utility
	}
}

type RequestBody struct {
	Cep string `json:"cep"`
}

var tracer trace.Tracer

// WeatherHandlerFunc handles the HTTP requests for weather data
// Função que lida com as requisições HTTP para obter dados meteorológicos
func (h *WeatherHandler) WeatherHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		serviceName := os.Getenv("OTEL_SERVICE_NAME")
		if serviceName == "" {
			serviceName = "service-b" // Se não houver, usa o nome "service-a"
		}
		tracer = otel.Tracer(serviceName)
		carrier := propagation.HeaderCarrier(r.Header)
		context := r.Context()
		ctx := otel.GetTextMapPropagator().Extract(context, carrier)
		_, serviceBRequestSpan := tracer.Start(ctx, "service-b-request")
		defer serviceBRequestSpan.End()
		// Decodificando o corpo da requisição para obter o CEP
		var requestBody RequestBody
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			// Caso não consiga decodificar o JSON, retorna erro 400
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		tracer = otel.Tracer(serviceName)
		if tracer == nil {
			log.Println("Tracer is nil! There is a problem with initialization.")
			http.Error(w, "Tracer initialization failed", http.StatusInternalServerError)
			return
		} else {
			log.Println("Tracer initialized successfully")
		}
		_, validateZipCodeSpan := tracer.Start(r.Context(), "validating-zip-code")
		defer validateZipCodeSpan.End()
		// Create channels for receiving location data from APIs
		// Cria canais para receber dados de localização das APIs
		chBrasilAPI := make(chan models.Location)
		chViaCEP := make(chan models.Location)

		// Validate the CEP input
		// Valida o CEP fornecido
		if !h.CepValidator.IsValidCep(requestBody.Cep) {
			// Respond with an error message if CEP is invalid
			// Retorna uma resposta de erro caso o CEP seja inválido
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
			serviceBRequestSpan.SetStatus(codes.Error, "Invalid Zip Code Sent")

			return
		}
		validateZipCodeSpan.SetStatus(codes.Ok, "Valid Zip Code Sent")
		serviceBRequestSpan.SetStatus(codes.Ok, "Valid Zip Code Sent")

		_, getLocationFromZipCodeSpan := tracer.Start(r.Context(), "getting-zip-code-information")
		defer getLocationFromZipCodeSpan.End()
		// Fetch location data based on CEP, using channels to simulate multiple API responses
		// Busca dados de localização com base no CEP, utilizando canais para simular múltiplas respostas de APIs
		location, err := h.LocationService.GetLocationFromCEP(requestBody.Cep, chBrasilAPI, chViaCEP)
		if err != nil || location.City == nil {
			// Respond with an error message if the location cannot be found
			// Retorna uma resposta de erro caso não seja possível encontrar a localização
			response := models.ErrorResponse{
				Error: "can not find zipcode", // Error message in English
			}
			// Set the HTTP status code to 422 (Unprocessable Entity)
			// Define o código de status HTTP como 422 (Entidade não processável)
			w.WriteHeader(http.StatusNotFound)

			// Encode the response into JSON and send it to the client
			// Codifica a resposta em JSON e envia para o cliente
			json.NewEncoder(w).Encode(response)
			getLocationFromZipCodeSpan.SetStatus(codes.Error, "Can not find zipcode")
			serviceBRequestSpan.SetStatus(codes.Error, "Can not find zipcode")

			return
		}
		getLocationFromZipCodeSpan.SetStatus(codes.Ok, "Found Zip Code")
		serviceBRequestSpan.SetStatus(codes.Ok, "Found Zip Code")

		_, getTemperatureSpan := tracer.Start(r.Context(), "getting-temerature-information")
		defer getTemperatureSpan.End()
		// Fetch temperature for the city
		// Busca a temperatura para a cidade
		tempC, err := h.WeatherService.GetTemperature(*location.City)
		if err != nil {
			// Respond with an error message if fetching the temperature fails
			// Retorna uma resposta de erro caso a busca pela temperatura falhe
			response := models.ErrorResponse{
				Error: "failed to get temperature", // Error message in English
			}
			// Set the HTTP status code to 422 (Unprocessable Entity)
			// Define o código de status HTTP como 500 (Erro interno do servidor)
			w.WriteHeader(http.StatusInternalServerError)

			// Encode the response into JSON and send it to the client
			// Codifica a resposta em JSON e envia para o cliente
			json.NewEncoder(w).Encode(response)
			getTemperatureSpan.SetStatus(codes.Error, "failed to get temperature")
			serviceBRequestSpan.SetStatus(codes.Error, "failed to get temperature")

			return
		}

		getTemperatureSpan.SetStatus(codes.Ok, "Found Temperature")
		serviceBRequestSpan.SetStatus(codes.Ok, "Found Temperature")

		// Convert temperature using the shared utility
		// Converte a temperatura utilizando a ferramenta compartilhada
		tempF := h.TemperatureConverter.CelsiusToFahrenheit(tempC)
		tempK := h.TemperatureConverter.CelsiusToKelvin(tempC)

		// Prepare the response with temperature data in Celsius, Fahrenheit, and Kelvin
		// Prepara a resposta com os dados de temperatura em Celsius, Fahrenheit e Kelvin
		response := models.TemperatureResponse{
			Celsius:    tempC,          // Temperature in Celsius
			Fahrenheit: tempF,          // Temperature in Fahrenheit
			Kelvin:     tempK,          // Temperature in Kelvin
			City:       *location.City, // City
		}

		// Send the response as JSON
		// Envia a resposta como JSON
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
		serviceBRequestSpan.SetStatus(codes.Ok, "Finished Request Successfully")
	}
}
