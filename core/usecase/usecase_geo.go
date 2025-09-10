package usecase

import (
	"fmt"

	"studentgit.kata.academy/Zhodaran/go-kata/adapters/adapter"

	"studentgit.kata.academy/Zhodaran/go-kata/core/entity"
)

func HandleGeocodeRequest(req entity.GeocodeRequest, geoService entity.GeoProvider, cache *adapter.Cache) (entity.ResponseAddresses, error) {
	cacheKey := fmt.Sprintf("geocode:%f:%f", req.Lat, req.Lng)

	// Проверка кэша
	if cachedGeo, found := cache.Get(cacheKey); found {
		return cachedGeo.(entity.ResponseAddresses), nil // Приведение типа
	}

	// Вызов сервиса
	geo, err := geoService.GetGeoCoordinatesGeocode(req.Lat, req.Lng)
	if err != nil {
		return entity.ResponseAddresses{}, err
	}
	cache.Set(cacheKey, geo)
	return geo, nil
}

func HandleGeocodeAddressReq(req entity.RequestAddressSearch, geoService entity.GeoProvider, cache *adapter.Cache) (entity.ResponseAddresses, error) {
	cacheKey := fmt.Sprintf("search:%s", req.Query)
	if cachedGeo, found := cache.Get(cacheKey); found {
		return cachedGeo.(entity.ResponseAddresses), nil
	}
	geo, err := geoService.GetGeoCoordinatesAddress(req.Query)
	if err != nil {
		return entity.ResponseAddresses{}, err
	}
	cache.Set(cacheKey, geo)
	return geo, nil
}
