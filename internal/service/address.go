package service

import (
	"context"
	"encoding/json"
	"fmt"
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
