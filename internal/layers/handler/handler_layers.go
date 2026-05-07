package handler

import (
	"net/http"

	"geo-project/internal/layers/service"
	apperrors "geo-project/pkg/errors"
	"geo-project/pkg/midleware"

	"github.com/labstack/echo/v5"
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

	layers, totalPages, err := h.service.GetLayers(ctx, req.GeometryType, req.Status, req.SortBy, req.Page, req.PageSize)
	if err != nil {
		return err
	}

	var response okResponse[LayerResponse]
	response.Data = make([]LayerResponse, 0, len(layers))
	for _, l := range layers {
		response.Data = append(response.Data, LayerResponse{
			ID:           l.ID,
			Name:         l.Name,
			Status:       l.Status,
			GeometryType: l.GeometryType,
			Description:  &l.Description,
			OwnerID:      l.OwnerID,
			CreatedAt:    l.CreatedAt,
			UpdatedAt:    l.UpdatedAt,
		})
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	response.Meta = PaginationResponse{
		TotalPages: totalPages,
		Page:       req.Page,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *layerHandler) HandleGetLayerByID(c *echo.Context) error {
	var req IDRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()

	layer, err := h.service.GetLayerByID(ctx, req.ID)
	if err != nil {
		return err
	}

	response := LayerResponse{
		ID:           layer.ID,
		Name:         layer.Name,
		Description:  &layer.Description,
		GeometryType: layer.GeometryType,
		Status:       layer.Status,
		OwnerID:      layer.OwnerID,
		CreatedAt:    layer.CreatedAt,
		UpdatedAt:    layer.UpdatedAt,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *layerHandler) HandleCreateLayer(c *echo.Context) error {
	var req CreateLayerRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	params := service.CreateLayerParams{
		Name:         req.Name,
		Description:  req.Description,
		GeometryType: req.GeometryType,
		SRID:         req.SRID,
		OwnerID:      req.OwnerID,
	}

	ctx := c.Request().Context()
	layer, err := h.service.CreateLayer(ctx, params)
	if err != nil {
		return err
	}

	response := LayerResponse{
		ID:           layer.ID,
		Name:         layer.Name,
		Description:  &layer.Description,
		GeometryType: layer.GeometryType,
		Status:       layer.Status,
		OwnerID:      layer.OwnerID,
		CreatedAt:    layer.CreatedAt,
		UpdatedAt:    layer.UpdatedAt,
	}

	return c.JSON(http.StatusCreated, response)
}

func (h *layerHandler) HandleUpdateLayer(c *echo.Context) error {
	var req UpdateLayerRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	if req.Name == nil && req.Description == nil {
		return apperrors.NewAppError("VALIDATION_ERROR", "Body empty")
	}

	cl, ok := midleware.GetClaims(c)
	if !ok || cl.UserID <= 0 {
		return apperrors.NewAppError("UNAUTHORIZED", "missing or invalid user claims")
	}
	userID := cl.UserID

	ctx := c.Request().Context()
	layer, err := h.service.UpdateLayer(ctx, int32(req.ID), userID, service.UpdateLayerParams{
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
	})
	if err != nil {
		return err
	}

	response := LayerResponse{
		ID:           layer.ID,
		Name:         layer.Name,
		Description:  &layer.Description,
		GeometryType: layer.GeometryType,
		Status:       layer.Status,
		OwnerID:      layer.OwnerID,
		CreatedAt:    layer.CreatedAt,
		UpdatedAt:    layer.UpdatedAt,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *layerHandler) HandleDeleteLayer(c *echo.Context) error {
	var req IDRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	cl, ok := midleware.GetClaims(c)
	if !ok || cl.UserID <= 0 {
		return apperrors.NewAppError("UNAUTHORIZED", "missing or invalid user claims")
	}
	userID := cl.UserID

	ctx := c.Request().Context()
	layer, err := h.service.DeleteLayer(ctx, int32(req.ID), userID)
	if err != nil {
		return err
	}

	response := LayerResponse{
		ID:           layer.ID,
		Name:         layer.Name,
		Description:  &layer.Description,
		GeometryType: layer.GeometryType,
		Status:       layer.Status,
		OwnerID:      layer.OwnerID,
		CreatedAt:    layer.CreatedAt,
		UpdatedAt:    layer.UpdatedAt,
	}

	return c.JSON(http.StatusOK, response)
}
