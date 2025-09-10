package http

import (
	"encoding/json"
	"net/http"

	"studentgit.kata.academy/Zhodaran/go-kata/adapters/adapter"
	"studentgit.kata.academy/Zhodaran/go-kata/adapters/controllers/controller/repository"
	"studentgit.kata.academy/Zhodaran/go-kata/core/entity"
	"studentgit.kata.academy/Zhodaran/go-kata/core/usecase"
)

type GeoService interface {
	GetGeoCoordinatesAddress(query string) (entity.ResponseAddresses, error)
	GetGeoCoordinatesGeocode(lat float64, lng float64) (entity.ResponseAddresses, error)
}

type GeoSvc struct {
	repo repository.GeoRepository // Используем интерфейс репозитория
}

type RequestAddressSearch struct {
	Query string `json:"query"`
}

func (s *GeoSvc) GetGeoCoordinatesGeocode(lat float64, lng float64) (entity.ResponseAddresses, error) {
	return s.repo.GetGeoCoordinatesGeocode(lat, lng)
}

func (s *GeoSvc) GetGeoCoordinatesAddress(query string) (entity.ResponseAddresses, error) {
	return s.repo.GetGeoCoordinatesAddress(query)
}

func geocodeHandler(resp entity.Responder, geoService entity.GeoProvider, cache *adapter.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req entity.GeocodeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			resp.ErrorBadRequest(w, err)
			return
		}

		geo, err := usecase.HandleGeocodeRequest(req, geoService, cache)
		if err != nil {
			resp.ErrorInternal(w, err)
			return
		}
		resp.OutputJSON(w, geo)
	}
}

func searchHandler(resp entity.Responder, geoService entity.GeoProvider, cache *adapter.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req entity.RequestAddressSearch
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			resp.ErrorBadRequest(w, err)
			return
		}

		geo, err := usecase.HandleGeocodeAddressReq(req, geoService, cache)
		if err != nil {
			resp.ErrorInternal(w, err)
			return
		}
		resp.OutputJSON(w, geo)
	}
}
