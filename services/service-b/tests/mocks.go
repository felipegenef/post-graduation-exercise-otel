package tests

import (
	"net/http"
	"service-b/models"
	"service-b/services"

	"github.com/stretchr/testify/mock"
)

// MockWeatherService com retorno fixo
type MockWeatherService struct {
	mock.Mock
}

func (m *MockWeatherService) GetTemperature(city string) (float64, error) {
	args := m.Called(city)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockWeatherService) GetClient() services.APIClient {
	args := m.Called()
	return args.Get(0).(services.APIClient)
}

// MockApiClient implementando o método Get
type MockApiClient struct {
	mock.Mock
}

// Mock do método Get com switch para URL
func (m *MockApiClient) Get(url string) (*http.Response, error) {
	args := m.Called(url)
	return args.Get(0).(*http.Response), args.Error(1)
}

// MockLocationService simula a obtenção de localização
type MockLocationService struct {
	mock.Mock
}

// Simula a obtenção de localização a partir do CEP
func (m *MockLocationService) GetLocationFromCEP(cep string, chBrasilAPI, chViaCEP chan models.Location) (models.Location, error) {
	args := m.Called(cep, chBrasilAPI, chViaCEP)
	return args.Get(0).(models.Location), args.Error(1)
}
