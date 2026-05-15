package handler

import (
	"net/http"

	"geo-project/internal/revisions/model"
	"geo-project/internal/revisions/service"
	apperrors "geo-project/pkg/errors"
	"geo-project/pkg/midleware"

	"github.com/labstack/echo/v5"
)

type revisionHandler struct {
	service service.RevisionService
}

func NewRevisionHandler(service service.RevisionService) *revisionHandler {
	return &revisionHandler{service: service}
}

func (h *revisionHandler) CreateRevision(c *echo.Context) error {
	var req CreateRevisionRequest
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

	revision := &model.Revision{
		FeatureID: req.FeatureID,
		AuthorID:  cl.UserID,
		ChangeLog: req.ChangeLog,
	}

	created, err := h.service.CreateRevision(c.Request().Context(), revision)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, CreateRevisionResponse{
		ID:        created.ID.Hex(),
		FeatureID: created.FeatureID,
		AuthorID:  created.AuthorID,
		ChangeLog: created.ChangeLog,
		Status:    created.Status,
		CreatedAt: created.CreatedAt,
		UpdatedAt: created.UpdatedAt,
	})
}
