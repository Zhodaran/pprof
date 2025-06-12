package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"studentgit.kata.academy/Zhodaran/go-kata/internal/controller"
	"studentgit.kata.academy/Zhodaran/go-kata/internal/entity"
	"studentgit.kata.academy/Zhodaran/go-kata/internal/service"
)

type GeocodeRequest struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type RequestAddressSearch struct {
	Query string `json:"query"`
}

func geocodeHandler(resp controller.Responder, geoService service.GeoProvider, cache *entity.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req GeocodeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			resp.ErrorBadRequest(w, err)
			return
		}

		cacheKey := fmt.Sprintf("geocode:%f:%f", req.Lat, req.Lng)

		// Проверяем кэш
		if cachedGeo, found := cache.Get(cacheKey); found {
			resp.OutputJSON(w, cachedGeo)
			return
		}

		geo, err := geoService.GetGeoCoordinatesGeocode(req.Lat, req.Lng)
		if err != nil {
			resp.ErrorInternal(w, err)
			return
		}
		cache.Set(cacheKey, geo)
		resp.OutputJSON(w, geo)
	}
}

func searchHandler(resp controller.Responder, geoService service.GeoProvider, cache *entity.Cache) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req RequestAddressSearch
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			resp.ErrorBadRequest(w, err)
			return
		}

		cacheKey := fmt.Sprintf("search:%s", req.Query)

		if cachedGeo, found := cache.Get(cacheKey); found {
			resp.OutputJSON(w, cachedGeo)
			return
		}

		geo, err := geoService.GetGeoCoordinatesAddress(req.Query)
		if err != nil {
			resp.ErrorInternal(w, err)
			return
		}
		cache.Set(cacheKey, geo)
		resp.OutputJSON(w, geo)
	}
}
