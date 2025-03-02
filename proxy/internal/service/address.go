package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/ekomobile/dadata/v2/api/suggest"
	"github.com/ekomobile/dadata/v2/client"
)

type GeoService struct {
	api       *suggest.Api
	apiKey    string
	secretKey string
}

type GeoProvider interface {
	AddressSearch(input string) ([]*Address, error)
	GeoCode(lat, lng string) ([]*Address, error)
	GetGeoCoordinatesAddress(query string) (ResponseAddresses, error)
	GetGeoCoordinatesGeocode(lat float64, lng float64) (ResponseAddresses, error)
}

func NewGeoService(apiKey, secretKey string) *GeoService {
	var err error
	endpointUrl, err := url.Parse("https://suggestions.dadata.ru/suggestions/api/4_1/rs/")
	if err != nil {
		return nil
	}

	creds := client.Credentials{
		ApiKeyValue:    apiKey,
		SecretKeyValue: secretKey,
	}

	api := suggest.Api{
		Client: client.NewClient(endpointUrl, client.WithCredentialProvider(&creds)),
	}

	return &GeoService{
		api:       &api,
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

type Address struct {
	City   string `json:"city"`
	Street string `json:"street"`
	House  string `json:"house"`
	Lat    string `json:"geo_lat"`
	Lon    string `json:"geo_lon"`
}

type RequestAddressSearch struct {
	Query string `json:"query"`
}

type GeocodeRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type ResponseAddresses struct {
	Addresses []*Address `json:"addresses"`
}

type ResponseAddress struct {
	Suggestions []struct {
		Address Address `json:"data"`
	} `json:"suggestions"`
}

type GeoServicer interface {
	GetGeoCoordinatesAddress(query string) (ResponseAddresses, error)
	GetGeoCoordinatesGeocode(lat float64, lng float64) (ResponseAddresses, error)
}

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
			House:  suggestion.Address.House,
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

func (g *GeoService) AddressSearch(input string) ([]*Address, error) {
	var res []*Address
	rawRes, err := g.api.Address(context.Background(), &suggest.RequestParams{Query: input})
	if err != nil {
		return nil, err
	}

	for _, r := range rawRes {
		if r.Data.City == "" || r.Data.Street == "" {
			continue
		}
		res = append(res, &Address{City: r.Data.City, Street: r.Data.Street, House: r.Data.House, Lat: r.Data.GeoLat, Lon: r.Data.GeoLon})
	}

	return res, nil
}

func (g *GeoService) GeoCode(lat, lng string) ([]*Address, error) {
	httpClient := &http.Client{}
	var data = strings.NewReader(fmt.Sprintf(`{"lat": %s, "lon": %s}`, lat, lng))
	req, err := http.NewRequest("POST", "https://suggestions.dadata.ru/suggestions/api/4_1/rs/geolocate/address", data)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", g.apiKey))
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	var geoCode GeoCode

	err = json.NewDecoder(resp.Body).Decode(&geoCode)
	if err != nil {
		return nil, err
	}
	var res []*Address
	for _, r := range geoCode.Suggestions {
		var address Address
		address.City = string(r.Data.City)
		address.Street = string(r.Data.Street)
		address.House = r.Data.House
		address.Lat = r.Data.GeoLat
		address.Lon = r.Data.GeoLon

		res = append(res, &address)
	}

	return res, nil
}
