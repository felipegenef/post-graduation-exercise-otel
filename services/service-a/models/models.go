package models

// RequestBody define a estrutura do corpo da requisição com o CEP
type RequestBody struct {
	Cep string `json:"cep"`
}

type ResponseBody struct {
	Celsius    float64 `json:"temp_C"`
	Fahrenheit float64 `json:"temp_F"`
	Kelvin     float64 `json:"temp_K"`
	City       string  `json:"city"`
}
