package api

import (
	"net/url"
)

type BusinessPlaceService struct {
	client *Client
}

func NewBusinessPlaceService(client *Client) *BusinessPlaceService {
	return &BusinessPlaceService{client: client}
}

func (s *BusinessPlaceService) ListBusinessPlaces(cursor string, count int) (*Response, error) {
	return s.client.Get("/business-support/business-places" + BuildPaginationQuery(cursor, count))
}

func (s *BusinessPlaceService) GetBusinessPlace(businessPlaceID string) (*Response, error) {
	return s.client.Get("/business-support/business-places/" + url.PathEscape(businessPlaceID))
}

func (s *BusinessPlaceService) CreateBusinessPlace(body map[string]interface{}) (*Response, error) {
	data, err := marshalBody(body)
	if err != nil {
		return nil, err
	}
	return s.client.Post("/business-support/business-places", data)
}

func (s *BusinessPlaceService) UpdateBusinessPlace(businessPlaceID string, body map[string]interface{}) (*Response, error) {
	data, err := marshalBody(body)
	if err != nil {
		return nil, err
	}
	return s.client.Patch("/business-support/business-places/"+url.PathEscape(businessPlaceID), data)
}

func (s *BusinessPlaceService) DeleteBusinessPlace(businessPlaceID string) (*Response, error) {
	return s.client.Delete("/business-support/business-places/" + url.PathEscape(businessPlaceID))
}
