package service

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// @Summary Get Geo Coordinates by Address
// @Description This endpoint allows you to get geo coordinates by address.
// @Tags geo
// @Accept json
// @Produce json
// @Param address body RequestAddressSearch true "Address search query"
// @Param Authorization header string true "Bearer {token}"
// @Success 200 {object} ResponseAddress "Успешное выполнение"
// @Failure 400 {object} string "Ошибка запроса"
// @Failure 500 {object} string "Ошибка подключения к серверу"
// @Security BearerAuth
// @Router /api/address/search [post]
func (g *GeoService) GetGeoCoordinatesAddress(query string) (ResponseAddresses, error) {
	url := "http://suggestions.dadata.ru/suggestions/api/4_1/rs/suggest/address"
	reqData := map[string]string{"query": query}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return ResponseAddresses{}, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return ResponseAddresses{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token d9e0649452a137b73d941aa4fb4fcac859372c8c")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ResponseAddresses{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResponseAddresses{}, err
	}

	var response ResponseAddress
	err = json.Unmarshal(body, &response)
	if err != nil {
		return ResponseAddresses{}, err
	}

	var addresses ResponseAddresses
	for _, suggestion := range response.Suggestions {
		address := &Address{
			City:   suggestion.Address.City,
			Street: suggestion.Address.Street,
			Lat:    suggestion.Address.Lat,
			Lon:    suggestion.Address.Lon,
		}
		addresses.Addresses = append(addresses.Addresses, address)
	}

	return addresses, nil
}

// @Summary Get Geo Coordinates by Latitude and Longitude
// @Description This endpoint allows you to get geo coordinates by latitude and longitude.
// @Tags geo
// @Accept json
// @Produce json
// @Param body body GeocodeRequest true "Geographic coordinates"
// @Param Authorization header string true "Bearer {token}"
// @Success 200 {object} ResponseAddress "Успешное выполнение"
// @Failure 400 {object} string "Ошибка запроса"
// @Failure 500 {object} string "Ошибка подключения к серверу"
// @Security BearerAuth
// @Router /api/address/geocode [post]
func (g *GeoService) GetGeoCoordinatesGeocode(lat float64, lng float64) (ResponseAddresses, error) {
	url := "http://suggestions.dadata.ru/suggestions/api/4_1/rs/geolocate/address"
	data := map[string]float64{"lat": lat, "lon": lng}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return ResponseAddresses{}, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return ResponseAddresses{}, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "token d9e0649452a137b73d941aa4fb4fcac859372c8c")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ResponseAddresses{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ResponseAddresses{}, err
	}

	var response ResponseAddress
	err = json.Unmarshal(body, &response)
	if err != nil {
		return ResponseAddresses{}, err
	}

	var addresses ResponseAddresses
	for _, suggestion := range response.Suggestions {
		address := &Address{
			City:   suggestion.Address.City,
			Street: suggestion.Address.Street,
			House:  suggestion.Address.House,
			Lat:    suggestion.Address.Lat,
			Lon:    suggestion.Address.Lon,
		}
		addresses.Addresses = append(addresses.Addresses, address)
	}

	return addresses, nil
}
