package gerbang

import (
	"context"

	"github.com/GTDGit/PPOB_BE/internal/domain"
)

// ========== Territory API Methods ==========

// GetProvinces fetches all provinces from Gerbang API
func (c *Client) GetProvinces(ctx context.Context) ([]*domain.Province, error) {
	resp, err := c.doRequestWithRetry(ctx, "GET", "/v1/territory/province", nil)
	if err != nil {
		return nil, err
	}

	var provinces []*domain.Province
	if err := c.parseData(resp, &provinces); err != nil {
		return nil, err
	}

	return provinces, nil
}

// GetCities fetches cities by province code from Gerbang API
func (c *Client) GetCities(ctx context.Context, provinceCode string) ([]*domain.City, error) {
	resp, err := c.doRequestWithRetry(ctx, "GET", "/v1/territory/city/"+provinceCode, nil)
	if err != nil {
		return nil, err
	}

	var cities []*domain.City
	if err := c.parseData(resp, &cities); err != nil {
		return nil, err
	}

	return cities, nil
}

// GetDistricts fetches districts by city code from Gerbang API
func (c *Client) GetDistricts(ctx context.Context, cityCode string) ([]*domain.District, error) {
	resp, err := c.doRequestWithRetry(ctx, "GET", "/v1/territory/district/"+cityCode, nil)
	if err != nil {
		return nil, err
	}

	var districts []*domain.District
	if err := c.parseData(resp, &districts); err != nil {
		return nil, err
	}

	return districts, nil
}

// GetSubDistricts fetches sub-districts by district code from Gerbang API
func (c *Client) GetSubDistricts(ctx context.Context, districtCode string) ([]*domain.SubDistrict, error) {
	resp, err := c.doRequestWithRetry(ctx, "GET", "/v1/territory/sub-district/"+districtCode, nil)
	if err != nil {
		return nil, err
	}

	var subDistricts []*domain.SubDistrict
	if err := c.parseData(resp, &subDistricts); err != nil {
		return nil, err
	}

	return subDistricts, nil
}

// GetPostalCodes fetches postal codes by sub-district code from Gerbang API
func (c *Client) GetPostalCodes(ctx context.Context, subDistrictCode string) ([]*domain.PostalCode, error) {
	resp, err := c.doRequestWithRetry(ctx, "GET", "/v1/territory/postal-code/sub-district/"+subDistrictCode, nil)
	if err != nil {
		return nil, err
	}

	var postalCodes []*domain.PostalCode
	if err := c.parseData(resp, &postalCodes); err != nil {
		return nil, err
	}

	return postalCodes, nil
}
