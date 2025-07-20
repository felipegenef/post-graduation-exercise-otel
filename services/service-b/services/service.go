package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"service-b/models"
	"time"
)

// APIClient defines the behavior of an external API client.
// APIClient define o comportamento de um cliente para consumir APIs externas.
type APIClient interface {
	Get(url string) (*http.Response, error)
}

// LocationService is an interface that defines the methods to interact with location services.
// LocationService é uma interface que define os métodos para interagir com serviços de localização.
type LocationService interface {
	GetLocationFromCEP(cep string, chBrasilAPI, chViaCEP chan models.Location) (models.Location, error)
}

// WeatherService is an interface that defines the methods for interacting with weather services.
// WeatherService é uma interface que define os métodos para interagir com serviços de clima.
type WeatherService interface {
	GetTemperature(city string) (float64, error) // Get the temperature for a given city.
	GetClient() APIClient                        // Return the API client used by the service.
}

// WeatherServiceImpl is the concrete implementation of the WeatherService interface.
// WeatherServiceImpl é a implementação concreta da interface WeatherService.
type WeatherServiceImpl struct {
	Client APIClient // The API client used for making requests.
}

// APIClientImpl is the concrete implementation of the APIClient interface.
// APIClientImpl é a implementação concreta da interface APIClient.
type APIClientImpl struct {
	Client *http.Client // The HTTP client used for making GET requests.
}

// LocationServiceImpl is the concrete implementation of the LocationService interface.
// LocationServiceImpl é a implementação concreta da interface LocationService.
type LocationServiceImpl struct {
	WeatherService WeatherService // Weather service instance to interact with weather data
}

// NewWeatherService creates and returns a new instance of WeatherServiceImpl.
// Cria e retorna uma nova instância do WeatherServiceImpl.
func NewWeatherService(client APIClient) WeatherService {
	return &WeatherServiceImpl{
		Client: client, // Assign the provided API client
	}
}

// NewLocationService creates and returns a new LocationServiceImpl instance.
// Cria e retorna uma nova instância do LocationServiceImpl.
func NewLocationService(weatherService WeatherService) LocationService {
	return &LocationServiceImpl{
		WeatherService: weatherService, // Assign the provided weather service
	}
}

// NewAPIClient creates and returns a new instance of APIClientImpl.
// Cria e retorna uma nova instância do APIClientImpl.
func NewAPIClient(client *http.Client) *APIClientImpl {
	return &APIClientImpl{
		Client: client, // Initialize the HTTP client
	}
}

// GetTemperature retrieves the current temperature for a given city.
// Recupera a temperatura atual para uma cidade específica.
func (ws *WeatherServiceImpl) GetTemperature(city string) (float64, error) {
	apiKey := os.Getenv("WEATHER_API_KEY") // Retrieve API key from environment variable
	// Fix spaces on names
	encodedCity := url.QueryEscape(city) // Encode the city name to ensure it works in a URL
	url := fmt.Sprintf("https://api.weatherapi.com/v1/current.json?key=%s&q=%s", apiKey, encodedCity)

	resp, err := ws.Client.Get(url) // Send GET request to the weather API
	if err != nil {
		return 0, err // Return error if the request fails
	}
	defer resp.Body.Close() // Close response body when done

	var weather models.WeatherResponse
	if err := json.NewDecoder(resp.Body).Decode(&weather); err != nil {
		return 0, err // Return error if the response cannot be decoded
	}

	return weather.Current.TempC, nil // Return the temperature in Celsius
}

// GetClient returns the APIClient used in WeatherServiceImpl.
// Retorna o APIClient usado no WeatherServiceImpl.
func (ws *WeatherServiceImpl) GetClient() APIClient {
	return ws.Client // Return the client used for HTTP requests
}

// GetLocationFromCEP retrieves location data based on a given CEP.
// Recupera dados de localização com base em um CEP fornecido.
func (ls *LocationServiceImpl) GetLocationFromCEP(cep string, chBrasilAPI, chViaCEP chan models.Location) (models.Location, error) {
	timeout := time.After(10 * time.Second) // Set a timeout for the operation
	// Asynchronously fetch data from the APIs
	// Busca os dados de forma assíncrona das APIs
	go ls.fetchFromBrasilAPI(cep, chBrasilAPI)
	go ls.fetchFromViaCEP(cep, chViaCEP)

	select {
	case res := <-chBrasilAPI: // Handle response from BrasilAPI
		if res.Localidade != nil {
			return res, nil // Return location data if valid
		}
		return models.Location{}, errors.New("error searching for CEP data")
	case res := <-chViaCEP: // Handle response from ViaCEP
		if res.Localidade != nil {
			return res, nil // Return location data if valid
		}
		return models.Location{}, errors.New("error searching for CEP data")
	case <-timeout: // Timeout after 10 seconds
		return models.Location{}, errors.New("timeout after 10 seconds") // Return timeout error
	}
}

// fetchFromBrasilAPI fetches location data from the BrasilAPI.
// Busca dados de localização da API BrasilAPI.
func (ls *LocationServiceImpl) fetchFromBrasilAPI(cep string, ch chan models.Location) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep) // BrasilAPI URL
	resp, err := ls.WeatherService.GetClient().Get(url)               // Use GetClient to avoid casting
	if err != nil || resp.StatusCode != http.StatusOK {
		ch <- models.Location{} // Send empty location if error occurs
		return
	}
	defer resp.Body.Close() // Close response body when done

	var address models.BrasilAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&address); err != nil {
		ch <- models.Location{} // Send empty location if decoding fails
		return
	}

	// Send location data to the channel
	ch <- models.Location{
		Cep:        &cep,
		Localidade: &address.Neighborhood,
		Uf:         &address.State,
		City:       &address.City,
	}
}

// fetchFromViaCEP fetches location data from the ViaCEP API.
// Busca dados de localização da API ViaCEP.
func (ls *LocationServiceImpl) fetchFromViaCEP(cep string, ch chan models.Location) {
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json", cep) // ViaCEP URL
	resp, err := ls.WeatherService.GetClient().Get(url)        // Use GetClient to avoid casting
	if err != nil || resp.StatusCode != http.StatusOK {
		ch <- models.Location{} // Send empty location if error occurs
		return
	}
	defer resp.Body.Close() // Close response body when done

	var address models.ViaCEPResponse
	if err := json.NewDecoder(resp.Body).Decode(&address); err != nil {
		ch <- models.Location{} // Send empty location if decoding fails
		return
	}

	if address.ErrorMessage == "true" {
		ch <- models.Location{} // Send empty location if decoding fails
		return
	}

	// Send location data to the channel
	ch <- models.Location{
		Cep:        &cep,
		Localidade: &address.Localidade,
		Uf:         &address.UF,
		City:       &address.Localidade,
	}
}

// Get performs an HTTP GET request.
// Realiza uma requisição HTTP GET.
func (api *APIClientImpl) Get(url string) (*http.Response, error) {
	return api.Client.Get(url) // Perform the GET request using the HTTP client
}
