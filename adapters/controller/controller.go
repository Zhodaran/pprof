package controller

import (
	"net/http"

	"studentgit.kata.academy/Zhodaran/go-kata/core/entity"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type Responder interface {
	OutputJSON(w http.ResponseWriter, responseData interface{})
	ErrorUnauthorized(w http.ResponseWriter, err error)
	ErrorBadRequest(w http.ResponseWriter, err error)
	ErrorForbidden(w http.ResponseWriter, err error)
	ErrorInternal(w http.ResponseWriter, err error)
}

type GeoProvider interface {
	AddressSearch(input string) ([]*entity.Address, error)
	GeoCode(lat, lng string) ([]*entity.Address, error)
	GetGeoCoordinatesAddress(query string) (entity.ResponseAddresses, error)
	GetGeoCoordinatesGeocode(lat float64, lng float64) (entity.ResponseAddresses, error)
}
