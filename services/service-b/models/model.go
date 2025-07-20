package models

type Location struct {
	Cep        *string `json:"cep"`
	Localidade *string `json:"localidade"` // City
	Uf         *string `json:"uf"`
	City       *string // alias for Localidade
}

type WeatherResponse struct {
	Location struct {
		Name           string  `json:"name"`
		Region         string  `json:"region"`
		Country        string  `json:"country"`
		Lat            float64 `json:"lat"`
		Lon            float64 `json:"lon"`
		TzID           string  `json:"tz_id"`
		LocalTimeEpoch int64   `json:"localtime_epoch"`
		LocalTime      string  `json:"localtime"`
	} `json:"location"`
	Current struct {
		LastUpdatedEpoch int64   `json:"last_updated_epoch"`
		LastUpdated      string  `json:"last_updated"`
		TempC            float64 `json:"temp_c"`
		TempF            float64 `json:"temp_f"`
		IsDay            int     `json:"is_day"`
		Condition        struct {
			Text string `json:"text"`
			Icon string `json:"icon"`
			Code int    `json:"code"`
		} `json:"condition"`
		WindMph    float64 `json:"wind_mph"`
		WindKph    float64 `json:"wind_kph"`
		WindDegree int     `json:"wind_degree"`
		WindDir    string  `json:"wind_dir"`
		PressureMb float64 `json:"pressure_mb"`
		PressureIn float64 `json:"pressure_in"`
		PrecipMm   float64 `json:"precip_mm"`
		PrecipIn   float64 `json:"precip_in"`
		Humidity   int     `json:"humidity"`
		Cloud      int     `json:"cloud"`
		FeelsLikeC float64 `json:"feelslike_c"`
		FeelsLikeF float64 `json:"feelslike_f"`
		WindchillC float64 `json:"windchill_c"`
		WindchillF float64 `json:"windchill_f"`
		HeatIndexC float64 `json:"heatindex_c"`
		HeatIndexF float64 `json:"heatindex_f"`
		DewPointC  float64 `json:"dewpoint_c"`
		DewPointF  float64 `json:"dewpoint_f"`
		VisKm      float64 `json:"vis_km"`
		VisMiles   float64 `json:"vis_miles"`
		UV         float64 `json:"uv"`
		GustMph    float64 `json:"gust_mph"`
		GustKph    float64 `json:"gust_kph"`
	} `json:"current"`
}

type TemperatureResponse struct {
	Celsius    float64 `json:"temp_C"`
	Fahrenheit float64 `json:"temp_F"`
	Kelvin     float64 `json:"temp_K"`
	City       string  `json:"city"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// Structs para as respostas das APIs
// Struct to hold the response from ViaCEP API
type ViaCEPResponse struct {
	CEP          string `json:"cep"`
	Logradouro   string `json:"logradouro"`
	Complemento  string `json:"complemento"`
	Unidade      string `json:"unidade"`
	Bairro       string `json:"bairro"`
	Localidade   string `json:"localidade"`
	UF           string `json:"uf"`
	Estado       string `json:"estado"`
	Regiao       string `json:"regiao"`
	IBGE         string `json:"ibge"`
	GIA          string `json:"gia"`
	DDD          string `json:"ddd"`
	SIAFI        string `json:"siafi"`
	ErrorMessage string `json:"erro"`
}

// Struct para a resposta da BrasilAPI
// Struct to hold the response from BrasilAPI
type BrasilAPIResponse struct {
	CEP          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}
