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
		CreatedAt: created.CreatedAt,
		UpdatedAt: created.UpdatedAt,
	})
}

func (h *revisionHandler) ListByFeatureID(c *echo.Context) error {
	var req ListRevisionsRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	revisions, err := h.service.ListByFeatureID(c.Request().Context(), req.FeatureID)
	if err != nil {
		return err
	}

	if revisions == nil {
		revisions = []*model.Revision{}
	}

	responses := make([]ListRevisionResponse, len(revisions))
	for i, rev := range revisions {
		responses[i] = ListRevisionResponse{
			ID:        rev.ID.Hex(),
			FeatureID: rev.FeatureID,
			AuthorID:  rev.AuthorID,
			ChangeLog: rev.ChangeLog,
			CreatedAt: rev.CreatedAt,
			UpdatedAt: rev.UpdatedAt,
		}
	}

	return c.JSON(http.StatusOK, responses)
}

func (h *revisionHandler) AddComment(c *echo.Context) error {
	var req AddCommentRequest
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

	err := h.service.AddComment(c.Request().Context(), req.ID, cl.UserID, req.Text)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, AddCommentResponse{
		Message: "comment added successfully",
	})
}

func (h *revisionHandler) GetRevisionByID(c *echo.Context) error {
	var req GetRevisionRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	revision, featureName, err := h.service.GetByID(c.Request().Context(), req.ID)
	if err != nil {
		return err
	}

	comments := make([]CommentResponse, len(revision.Comments))
	for i, comment := range revision.Comments {
		comments[i] = CommentResponse{
			AuthorID:  comment.AuthorID,
			Text:      comment.Text,
			CreatedAt: comment.CreatedAt,
		}
	}

	return c.JSON(http.StatusOK, GetRevisionResponse{
		ID:          revision.ID.Hex(),
		FeatureID:   revision.FeatureID,
		AuthorID:    revision.AuthorID,
		ChangeLog:   revision.ChangeLog,
		Comments:    comments,
		CreatedAt:   revision.CreatedAt,
		UpdatedAt:   revision.UpdatedAt,
		FeatureName: featureName,
	})
}
