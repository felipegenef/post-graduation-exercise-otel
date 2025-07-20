package models

// RequestBody define a estrutura do corpo da requisição com o CEP
type RequestBody struct {
	Cep string `json:"cep"`
}
