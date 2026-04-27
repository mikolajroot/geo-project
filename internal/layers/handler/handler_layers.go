package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"geo-project/internal/layers/service"
)


type layerHandler struct {
	service service.LayerService
}


func NewLayerHandler(service service.LayerService) *layerHandler {
	return &layerHandler{
		service: service,
	}
}


func (h *layerHandler) HandleGetLayers(c *echo.Context) error {
	var req GetLayersRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	layers,totalPages, err := h.service.GetLayers(ctx,req.GeometryType,req.Status,req.SortBy,req.Page,req.PageSize)
	if err != nil {
		return err
	}

	var response okResponse[LayerResponse];
	response.Data = make([]LayerResponse, 0, len(layers))
	for _, l := range layers {
		response.Data = append(response.Data, LayerResponse{
			ID:   l.ID,
			Name: l.Name,
			Status: l.Status,
			GeometryType: l.GeometryType,
			Description: &l.Description,
			OwnerID: l.OwnerID,
			CreatedAt: l.CreatedAt,
			UpdatedAt: l.UpdatedAt,
		})
	}

	response.Meta = PaginationResponse{
		TotalPages: totalPages,
		PageSize: req.Page,
	}

	return c.JSON(http.StatusOK, response)
}