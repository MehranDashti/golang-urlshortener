package handler

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"urlshortener/internal/apperror"
	"urlshortener/internal/model"
	"urlshortener/internal/service"
)

type AdminService interface {
	GetAllLinks(ctx context.Context) ([]*model.URL, *apperror.AppError)
	GetAllLinksPaginated(ctx context.Context,
		params model.PaginationParams) (*model.PaginatedResult[*model.URL], *apperror.AppError)
	DeleteLink(ctx context.Context, id string) *apperror.AppError
	GetAllUsers(ctx context.Context) ([]*model.User, *apperror.AppError)
	GetAllUsersPaginated(ctx context.Context,
		params model.PaginationParams) (*model.PaginatedResult[*model.User], *apperror.AppError)
	DeleteUser(ctx context.Context, id string) *apperror.AppError
	GetDashboard(ctx context.Context) (*service.DashboardData, *apperror.AppError)
	WriteLinksCSV(ctx context.Context, w io.Writer) error
}
type AdminHandler struct {
	service AdminService
}

func NewAdminHandler(service AdminService) *AdminHandler {
	return &AdminHandler{service: service}
}

func (h *AdminHandler) ListLinks(c *gin.Context) {
	urls, appErr := h.service.GetAllLinks(c.Request.Context())
	if appErr != nil {
		respondError(c, appErr)
		return
	}
	respondData(c, http.StatusOK, "عملیات با موفقیت انجام شد", urls)
}

func (h *AdminHandler) DeleteLink(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondError(c, apperror.BadRequest("id is required"))
		return
	}

	if appErr := h.service.DeleteLink(
		c.Request.Context(), id); appErr != nil {
		return
	}

	respondData[*struct{}](c, http.StatusOK, "لینک با موفقیت حذف شد", nil)
}

func (h *AdminHandler) ListUsers(c *gin.Context) {
	users, appErr := h.service.GetAllUsers(c.Request.Context())
	if appErr != nil {
		respondError(c, appErr)
		return
	}

	type safeUser struct {
		ID        string     `json:"id"`
		Email     string     `json:"email"`
		Role      model.Role `json:"role"`
		CreatedAt time.Time  `json:"created_at"`
	}

	result := make([]safeUser, len(users))
	for i, u := range users {
		result[i] = safeUser{
			ID:        u.ID,
			Email:     u.Email,
			Role:      u.Role,
			CreatedAt: u.CreatedAt,
		}
	}

	respondData(c, http.StatusOK, "عملیات با موفقیت انجام شد", result)
}

func (h *AdminHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		respondError(c, apperror.BadRequest("id is required"))
		return
	}

	if appErr := h.service.DeleteUser(
		c.Request.Context(), id); appErr != nil {
		respondError(c, appErr)
		return
	}

	respondData[*struct{}](c, http.StatusOK, "کاربر با موفقیت حذف شد", nil)
}

func (h *AdminHandler) Dashboard(c *gin.Context) {
	data, appErr := h.service.GetDashboard(
		c.Request.Context())
	if appErr != nil {
		respondError(c, appErr)
		return
	}

	respondSuccess(c, http.StatusOK,
		"عملیات با موفقیت انجام شد",
		gin.H{
			"links_count": len(data.Links),
			"users_count": len(data.Users),
			"links":       data.Links,
			"users":       data.Users,
		})
}

func (h *AdminHandler) ExportLinksCSV(c *gin.Context) {
	// Tell the browser this is a CSV file download
	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition",
		"attachment; filename=links.csv")

	// c.Writer implements io.Writer
	// We stream directly to the HTTP response — no buffer needed
	// First byte goes to client immediately
	if err := h.service.WriteLinksCSV(
		c.Request.Context(), c.Writer); err != nil {
		// Can't change status code here — headers already sent
		// Log the error, client gets partial CSV
		slog.Error("ExportLinksCSV failed", "error", err)
		return
	}
}

func (h *AdminHandler) ListLinksPaginated(c *gin.Context) {
	params, appErr := parsePagination(c)
	if appErr != nil {
		respondError(c, appErr)
		return
	}

	result, appErr := h.service.GetAllLinksPaginated(
		c.Request.Context(), params)
	if appErr != nil {
		respondError(c, appErr)
		return
	}

	respondSuccess(c, http.StatusOK,
		"عملیات با موفقیت انجام شد", result)
}

func (h *AdminHandler) ListUsersPaginated(c *gin.Context) {
	params, appErr := parsePagination(c)
	if appErr != nil {
		respondError(c, appErr)
		return
	}

	// Safe user response — no passwords
	type safeUser struct {
		ID        string     `json:"id"`
		Email     string     `json:"email"`
		Role      model.Role `json:"role"`
		CreatedAt time.Time  `json:"created_at"`
	}

	result, appErr := h.service.GetAllUsersPaginated(
		c.Request.Context(), params)
	if appErr != nil {
		respondError(c, appErr)
		return
	}

	// Map to safe users
	safeUsers := make([]safeUser, len(result.Data))
	for i, u := range result.Data {
		safeUsers[i] = safeUser{
			ID:        u.ID,
			Email:     u.Email,
			Role:      u.Role,
			CreatedAt: u.CreatedAt,
		}
	}

	// Rebuild result with safe users
	safeResult := model.PaginatedResult[safeUser]{
		Data:       safeUsers,
		Total:      result.Total,
		Page:       result.Page,
		Limit:      result.Limit,
		TotalPages: result.TotalPages,
	}

	respondSuccess(c, http.StatusOK,
		"عملیات با موفقیت انجام شد", safeResult)
}
