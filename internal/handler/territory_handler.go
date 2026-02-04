package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/GTDGit/PPOB_BE/internal/service"
)

// TerritoryHandler handles territory-related HTTP requests
type TerritoryHandler struct {
	service *service.TerritoryService
}

// NewTerritoryHandler creates a new territory handler
func NewTerritoryHandler(service *service.TerritoryService) *TerritoryHandler {
	return &TerritoryHandler{service: service}
}

// GetProvinces handles GET /v1/territory/provinces
func (h *TerritoryHandler) GetProvinces(c *gin.Context) {
	provinces, err := h.service.GetProvinces(c.Request.Context())
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Berhasil mengambil data provinsi",
		"data":    provinces,
		"meta": gin.H{
			"total": len(provinces),
		},
	})
}

// GetCities handles GET /v1/territory/cities/:provinceCode
func (h *TerritoryHandler) GetCities(c *gin.Context) {
	provinceCode := c.Param("provinceCode")

	cities, err := h.service.GetCitiesByProvince(c.Request.Context(), provinceCode)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Berhasil mengambil data kota/kabupaten",
		"data":    cities,
		"meta": gin.H{
			"total":        len(cities),
			"provinceCode": provinceCode,
		},
	})
}

// GetDistricts handles GET /v1/territory/districts/:cityCode
func (h *TerritoryHandler) GetDistricts(c *gin.Context) {
	cityCode := c.Param("cityCode")

	districts, err := h.service.GetDistrictsByCity(c.Request.Context(), cityCode)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Berhasil mengambil data kecamatan",
		"data":    districts,
		"meta": gin.H{
			"total":    len(districts),
			"cityCode": cityCode,
		},
	})
}

// GetSubDistricts handles GET /v1/territory/sub-districts/:districtCode
func (h *TerritoryHandler) GetSubDistricts(c *gin.Context) {
	districtCode := c.Param("districtCode")

	subDistricts, err := h.service.GetSubDistrictsByDistrict(c.Request.Context(), districtCode)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Berhasil mengambil data kelurahan",
		"data":    subDistricts,
		"meta": gin.H{
			"total":        len(subDistricts),
			"districtCode": districtCode,
		},
	})
}

// GetPostalCodes handles GET /v1/territory/postal-codes/:subDistrictCode
func (h *TerritoryHandler) GetPostalCodes(c *gin.Context) {
	subDistrictCode := c.Param("subDistrictCode")

	postalCodes, err := h.service.GetPostalCodesBySubDistrict(c.Request.Context(), subDistrictCode)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Berhasil mengambil data kode pos",
		"data":    postalCodes,
		"meta": gin.H{
			"total":           len(postalCodes),
			"subDistrictCode": subDistrictCode,
		},
	})
}

// SearchByPostalCode handles GET /v1/territory/search/postal-code/:postalCode
func (h *TerritoryHandler) SearchByPostalCode(c *gin.Context) {
	postalCode := c.Param("postalCode")

	results, err := h.service.SearchByPostalCode(c.Request.Context(), postalCode)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	respondWithSuccess(c, http.StatusOK, gin.H{
		"message": "Berhasil mencari data kode pos",
		"data":    results,
		"meta": gin.H{
			"total":      len(results),
			"postalCode": postalCode,
		},
	})
}
