package controller

import (
	"encoding/json"
	"net/http"

	"studentgit.kata.academy/Zhodaran/go-kata/internal/entity"
	"studentgit.kata.academy/Zhodaran/go-kata/internal/usecase"
)

type Controller struct {
	geoService usecase.GeoServicer
}

func NewController(geoService usecase.GeoServicer) *Controller {
	return &Controller{geoService: geoService}
}

// @Summary Get Geo Coordinates by Address
// @Description This endpoint allows you to get geo coordinates by address.
// @Tags geo
// @Accept json
// @Produce json
// @Param address body service.RequestAddressSearch true "Address search query"
// @Success 200 {object} service.ResponseAddress "Успешное выполнение"
// @Failure 400 {object} string "Ошибка запроса"
// @Failure 500 {object} string "Ошибка подключения к серверу"
// @Security BearerAuth
// @Router /api/address/search [post]
func (c *Controller) GetGeoCoordinatesAddress(w http.ResponseWriter, r *http.Request) {
	var req entity.RequestAddressSearch
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	geo, err := c.geoService.GetGeoCoordinatesAddress(req.Query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(geo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// @Summary Get Geo Coordinates by Latitude and Longitude
// @Description This endpoint allows you to get geo coordinates by latitude and longitude.
// @Tags geo
// @Accept json
// @Produce json
// @Param body body service.GeocodeRequest true "Geographic coordinates"
// @Success 200 {object} service.ResponseAddress "Успешное выполнение"
// @Failure 400 {object} string "Ошибка запроса"
// @Failure 500 {object} string "Ошибка подключения к серверу"
// @Security BearerAuth
// @Router /api/address/geocode [post]
func (c *Controller) GetGeoCoordinatesGeocode(w http.ResponseWriter, r *http.Request) {
	var req entity.GeocodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	geo, err := c.geoService.GetGeoCoordinatesGeocode(req.Lat, req.Lng)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.Marshal(geo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
