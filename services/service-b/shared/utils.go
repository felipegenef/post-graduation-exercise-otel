package shared

import (
	"regexp"
)

// TemperatureConverter provides conversion methods for temperature.
// TemperatureConverter fornece métodos para conversão de temperatura.
type TemperatureConverter struct{}

// CelsiusToFahrenheit converts Celsius to Fahrenheit.
// Converte Celsius para Fahrenheit.
func (tc *TemperatureConverter) CelsiusToFahrenheit(c float64) float64 {
	return c*1.8 + 32 // Formula to convert Celsius to Fahrenheit
}

// CelsiusToKelvin converts Celsius to Kelvin.
// Converte Celsius para Kelvin.
func (tc *TemperatureConverter) CelsiusToKelvin(c float64) float64 {
	return c + 273 // Formula to convert Celsius to Kelvin
}

// CepValidator validates if a CEP is in the correct format.
// CepValidator valida se um CEP está no formato correto.
type CepValidator struct {
	RegexPattern string // The regular expression pattern used for validation.
	// Padrão de expressão regular usado para validação.
}

// NewCepValidator creates a new CepValidator with a given regex pattern.
// Cria um novo CepValidator com o padrão de regex fornecido.
func NewCepValidator(pattern string) *CepValidator {
	return &CepValidator{RegexPattern: pattern} // Initialize CepValidator with the provided pattern
}

// IsValidCep checks if the provided CEP matches the regex pattern.
// Verifica se o CEP fornecido corresponde ao padrão da expressão regular.
func (cv *CepValidator) IsValidCep(cep string) bool {
	match, _ := regexp.MatchString(cv.RegexPattern, cep) // Check if the CEP matches the regex pattern
	return match                                         // Return true if it matches, fals
}
