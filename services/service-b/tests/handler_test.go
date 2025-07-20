package tests

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"service-b/handlers"
	"service-b/models"
	"service-b/services"
	"service-b/shared"
)

func TestWeatherHandlerSuccess(t *testing.T) {
	// Mock API client and services
	apiKey := os.Getenv("WEATHER_API_KEY")
	cep := "12345678"
	chBrasilAPI := make(chan models.Location)
	chViaCEP := make(chan models.Location)
	mockApiClient := new(MockApiClient)
	weatherService := services.NewWeatherService(mockApiClient)
	locationService := services.NewLocationService(weatherService)
	handler := handlers.NewWeatherHandler(locationService, weatherService, &shared.TemperatureConverter{}, chBrasilAPI, chViaCEP)

	// mock CEP Responses
	mockApiClient.On("Get", fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)).
		Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"cep": "12345678","state": "SP","city": "São Paulo","neighborhood": "SP","street": "Rua XV de Novembro","service": "ViaCEP"}`))),
		}, nil)
	mockApiClient.On("Get", fmt.Sprintf("http://viacep.com.br/ws/%s/json", cep)).
		Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"cep": "12345678","logradouro": "Rua XV de Novembro","complemento": "Apto 101","unidade": "Unidade 2","bairro": "Centro","localidade": "SP","uf": "SP","estado": "São Paulo","regiao": "Sudeste","ibge": "3550308","gia": "1004","ddd": "11","siafi": "1234"}`))),
		}, nil)
	// Mock Weather API response
	mockApiClient.On("Get", fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, "SP")).
		Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"current": {"temp_c":22.0}}`))),
		}, nil)

	// Create a mock HTTP request
	req, err := http.NewRequest("GET", fmt.Sprintf("/weather?cep=%s", cep), nil)
	assert.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Create a handler function
	handlerFunc := handler.WeatherHandlerFunc()

	// Call the handler
	handlerFunc.ServeHTTP(rr, req)

	// Assert status code
	assert.Equal(t, http.StatusOK, rr.Code)

	// Assert response body
	expectedResponse := `{"temp_C":22,"temp_F":71.6,"temp_K":295}`
	assert.JSONEq(t, expectedResponse, rr.Body.String())
}

func TestWeatherHandlerInvalidCepValidator(t *testing.T) {
	cep := "123456"
	chBrasilAPI := make(chan models.Location)
	chViaCEP := make(chan models.Location)
	mockApiClient := new(MockApiClient)
	weatherService := services.NewWeatherService(mockApiClient)
	locationService := services.NewLocationService(weatherService)
	handler := handlers.NewWeatherHandler(locationService, weatherService, &shared.TemperatureConverter{}, chBrasilAPI, chViaCEP)

	// Create a mock HTTP request
	req, err := http.NewRequest("GET", fmt.Sprintf("/weather?cep=%s", cep), nil)
	assert.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Create a handler function
	handlerFunc := handler.WeatherHandlerFunc()

	// Call the handler
	handlerFunc.ServeHTTP(rr, req)

	// Assert status code
	assert.Equal(t, http.StatusUnprocessableEntity, rr.Code)

	// Assert response body
	expectedResponse := `{"error":"invalid zipcode"}`
	assert.JSONEq(t, expectedResponse, rr.Body.String())
}

func TestWeatherHandlerCepNotFound(t *testing.T) {
	cep := "12345678"
	chBrasilAPI := make(chan models.Location)
	chViaCEP := make(chan models.Location)
	mockApiClient := new(MockApiClient)
	weatherService := services.NewWeatherService(mockApiClient)
	locationService := services.NewLocationService(weatherService)
	handler := handlers.NewWeatherHandler(locationService, weatherService, &shared.TemperatureConverter{}, chBrasilAPI, chViaCEP)

	// mock CEP Responses
	mockApiClient.On("Get", fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)).
		Return(&http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(bytes.NewReader([]byte(``))),
		}, nil).Once()
	mockApiClient.On("Get", fmt.Sprintf("http://viacep.com.br/ws/%s/json", cep)).
		Return(&http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(bytes.NewReader([]byte(``))),
		}, nil).Once()

	// Create a mock HTTP request
	req, err := http.NewRequest("GET", fmt.Sprintf("/weather?cep=%s", cep), nil)
	assert.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Create a handler function
	handlerFunc := handler.WeatherHandlerFunc()

	// Call the handler
	handlerFunc.ServeHTTP(rr, req)

	// Assert status code
	assert.Equal(t, http.StatusNotFound, rr.Code)

	// Assert response body
	expectedResponse := `{"error": "can not find zipcode"}`
	assert.JSONEq(t, expectedResponse, rr.Body.String())
}

func TestWeatherHandlerInternalServerError(t *testing.T) {
	cep := "12345678"
	chBrasilAPI := make(chan models.Location)
	chViaCEP := make(chan models.Location)
	mockApiClient := new(MockApiClient)
	weatherService := new(MockWeatherService)
	locationService := services.NewLocationService(weatherService)
	handler := handlers.NewWeatherHandler(locationService, weatherService, &shared.TemperatureConverter{}, chBrasilAPI, chViaCEP)

	// mock CEP Responses
	mockApiClient.On("Get", fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)).
		Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"cep": "12345678","state": "SP","city": "São Paulo","neighborhood": "SP","street": "Rua XV de Novembro","service": "ViaCEP"}`))),
		}, nil)
	mockApiClient.On("Get", fmt.Sprintf("http://viacep.com.br/ws/%s/json", cep)).
		Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"cep": "12345678","logradouro": "Rua XV de Novembro","complemento": "Apto 101","unidade": "Unidade 2","bairro": "Centro","localidade": "SP","uf": "SP","estado": "São Paulo","regiao": "Sudeste","ibge": "3550308","gia": "1004","ddd": "11","siafi": "1234"}`))),
		}, nil)

	weatherService.On("GetTemperature", mock.Anything).Return(0.0, fmt.Errorf("Error Getting Temperature")).Once()
	weatherService.On("GetClient", mock.Anything).Return(mockApiClient)

	// Create a mock HTTP request
	req, err := http.NewRequest("GET", fmt.Sprintf("/weather?cep=%s", cep), nil)
	assert.NoError(t, err)

	// Create a response recorder to capture the response
	rr := httptest.NewRecorder()

	// Create a handler function
	handlerFunc := handler.WeatherHandlerFunc()

	// Call the handler
	handlerFunc.ServeHTTP(rr, req)

	// Assert status code
	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	// Assert response body
	expectedResponse := `{"error": "failed to get temperature"}`
	assert.JSONEq(t, expectedResponse, rr.Body.String())
}
