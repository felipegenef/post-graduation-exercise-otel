package tests

import (
	"service-b/shared"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemperatureConverter_CelsiusToFahrenheit(t *testing.T) {
	tc := &shared.TemperatureConverter{}

	result := tc.CelsiusToFahrenheit(25)
	assert.Equal(t, 77.0, result)
}

func TestTemperatureConverter_CelsiusToKelvin(t *testing.T) {
	tc := &shared.TemperatureConverter{}

	result := tc.CelsiusToKelvin(25)
	assert.Equal(t, 298.0, result)
}
