package tests

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"service-b/models"
	"service-b/services"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRaceFetch(t *testing.T) {
	mockLocationService := new(MockLocationService)

	// Canais
	chBrasilAPI := make(chan models.Location)
	chViaCEP := make(chan models.Location)

	cep := "12345678"
	localidadeBrasilAPI := "São Paulo"
	ufBrasilAPI := "SP"
	locationBrasilAPI := models.Location{
		Cep:        &cep,
		Localidade: &localidadeBrasilAPI,
		Uf:         &ufBrasilAPI,
	}

	localidadeViaCEP := "Rio de Janeiro"
	ufViaCEP := "RJ"
	locationViaCEP := models.Location{
		Cep:        &cep,
		Localidade: &localidadeViaCEP,
		Uf:         &ufViaCEP,
	}

	// Expect behavior on MockLocationService
	mockLocationService.On("GetLocationFromCEP", cep, chBrasilAPI, chViaCEP).Return(locationBrasilAPI, nil)

	// Testando a resposta com chBrasilAPI primeiro
	go func() {
		chBrasilAPI <- locationBrasilAPI
	}()

	// Testando a resposta com chViaCEP primeiro
	go func() {
		time.Sleep(1 * time.Second)
		chViaCEP <- locationViaCEP
	}()

	// Esperado o resultado
	resultado, err := mockLocationService.GetLocationFromCEP(cep, chBrasilAPI, chViaCEP)
	assert.NoError(t, err)
	assert.Equal(t, locationBrasilAPI, resultado)

	// Verificando chamadas
	mockLocationService.AssertExpectations(t)
}

func TestHttpFetchSuccess(t *testing.T) {
	mockApiClient := new(MockApiClient)
	apiKey := os.Getenv("WEATHER_API_KEY")
	weatherService := services.NewWeatherService(mockApiClient)

	// Mock do retorno do método Get
	mockApiClient.On("Get", fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%v&q=SP", apiKey)).
		Return(&http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"current": {"temp_c":13.14}}`))),
		}, nil).Once()

	mockApiClient.On("Get", mock.Anything).Return(&http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"current": {"temp_c":13.0}}`))),
	}, nil).Once()

	// Teste para a URL "SP"
	response, err := weatherService.GetTemperature("SP")
	assert.NoError(t, err)
	assert.Equal(t, 13.14, response)

	// Teste para a URL "other-city" (outro valor)
	response, err = weatherService.GetTemperature("other-city")
	assert.NoError(t, err)
	assert.Equal(t, 13.0, response)
}

func TestHttpFetchNotFound(t *testing.T) {
	cep := "11111111"
	chBrasilAPI := make(chan models.Location)
	chViaCEP := make(chan models.Location)
	mockApiClient := new(MockApiClient)
	weatherService := services.NewWeatherService(mockApiClient)
	locationService := services.NewLocationService(weatherService)

	// Mock do retorno do método Get
	mockApiClient.On("Get", mock.Anything).
		Return(&http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(bytes.NewReader([]byte(`{"message":"Todos os serviços de CEP retornaram erro.","type":"service_error","name":"CepPromiseError","errors":[{"name":"ServiceError","message":"A autenticacao de null falhou!","service":"correios"},{"name":"ServiceError","message":"Cannot read properties of undefined (reading 'replace')","service":"viacep"},{"name":"ServiceError","message":"Erro ao se conectar com o serviço WideNet.","service":"widenet"},{"name":"ServiceError","message":"Erro ao se conectar com o serviço dos Correios Alt.","service":"correios-alt"}]}`))),
		}, nil)

		// Teste para a URL "SP"
	response, err := locationService.GetLocationFromCEP(cep, chBrasilAPI, chViaCEP)

	assert.Equal(t, models.Location{}, response)
	assert.NotEqual(t, nil, err)
}
