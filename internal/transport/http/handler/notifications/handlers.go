// Package notifications provides HTTP handlers for notification-related operations.
package notifications

import (
	"delay/internal/domain"
	svcnotifications "delay/internal/service/notifications"
	"delay/internal/transport/http/helpers"
	"delay/internal/transport/http/response"
	"errors"
	"net/http"
	"time"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/logger"
)

// Handler implements HTTP transport logic for notifications.
type Handler struct {
	svc Service
	log *logger.ZerologAdapter
}

// New creates a new Handler instance.
func New(svc Service, log *logger.ZerologAdapter) *Handler {
	return &Handler{
		svc: svc,
		log: log,
	}
}

// GetAll returns all notifications.
func (h *Handler) GetAll(c *ginext.Context) {
	notes, err := h.svc.GetAll(c.Request.Context())
	if err != nil {
		h.log.Error("failed to fetch notifications", "err", err)
		_ = response.Fail(c.Writer, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	if notes == nil {
		_ = response.OK(c.Writer, []notificationResponse{})
		return
	}

	// notes is a pointer to a slice (*[]domain.Notification), so we dereference it with *notes
	resp := make([]notificationResponse, len(*notes))
	for i, n := range *notes {
		resp[i] = notificationResponse{
			ID:          n.ID.String(),
			Destination: n.Destination,
			// Using explicit type conversion to string for custom types
			Channel:    string(n.Channel),
			Status:     string(n.Status),
			Message:    n.Message,
			DataSentAt: n.DataToSent.Format(time.RFC3339),
			CreatedAt:  n.CreatedAt.Format(time.RFC3339),
		}
	}

	h.log.Info("notifications fetched successfully")

	// Crucial: return 'resp' (DTO with lowercase JSON tags), not 'notes'
	_ = response.OK(c.Writer, resp)
}

// Create creates a new notification.
func (h *Handler) Create(c *ginext.Context) {
	var req createRequest

	if err := helpers.DecodeJSON(c.Request, &req); err != nil {
		h.log.Warn("request decoding failed",
			"err", err,
			"component", "notifications_handler",
			"method", "Create",
		)
		_ = response.Fail(c.Writer, http.StatusBadRequest, ErrBadRequest)
		return
	}

	notification := domain.Notification{
		Message:     req.Message,
		DataToSent:  req.DataToSent,
		Channel:     req.Channel,
		Destination: req.Destination,
	}

	id, err := h.svc.Create(c.Request.Context(), &notification)
	if err != nil {
		h.log.Error("failed to create notification", "err", err)
		_ = response.Fail(c.Writer, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	h.log.Info("notification created successfully",
		"id", id,
		"channel", notification.Channel,
		"destination", notification.Destination,
	)

	_ = response.OK(c.Writer, id)
}

// GetStatus returns notification status by ID.
func (h *Handler) GetStatus(c *ginext.Context) {
	id, err := helpers.ParseUUIDParam(c, "id")
	if err != nil {
		h.log.Warn("invalid id parameter", "err", err)
		_ = response.Fail(c.Writer, http.StatusBadRequest, ErrBadRequest)
		return
	}

	status, err := h.svc.Status(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, svcnotifications.ErrNotFound) {
			_ = response.Fail(c.Writer, http.StatusNotFound, ErrNotFound)
			return
		}

		h.log.Error("failed to get notification status", "id", id, "err", err)
		_ = response.Fail(c.Writer, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	h.log.Info("notification status fetched successfully", "id", id)
	_ = response.OK(c.Writer, status)
}

// Cancel cancels a notification by ID.
func (h *Handler) Cancel(c *ginext.Context) {
	id, err := helpers.ParseUUIDParam(c, "id")
	if err != nil {
		h.log.Warn("invalid id parameter", "err", err)
		_ = response.Fail(c.Writer, http.StatusBadRequest, ErrBadRequest)
		return
	}

	if err := h.svc.Cancel(c.Request.Context(), id); err != nil {
		if errors.Is(err, svcnotifications.ErrNotFound) {
			_ = response.Fail(c.Writer, http.StatusNotFound, ErrNotFound)
			return
		}

		h.log.Error("failed to cancel notification", "id", id, "err", err)
		_ = response.Fail(c.Writer, http.StatusInternalServerError, ErrInternalServerError)
		return
	}

	h.log.Info("notification canceled successfully", "id", id)
	c.Writer.WriteHeader(http.StatusNoContent)
}
