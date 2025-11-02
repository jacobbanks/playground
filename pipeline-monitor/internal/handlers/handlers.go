package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"pipeline-monitor/internal/domain/service"
	"pipeline-monitor/internal/infrastructure/monitor"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handlers contains all HTTP handlers for the application
type Handlers struct {
	serviceRepo service.Repository
	monitor     *monitor.ServiceMonitor
}

// New creates a new handlers instance
func New(repo service.Repository, monitor *monitor.ServiceMonitor) *Handlers {
	return &Handlers{
		serviceRepo: repo,
		monitor:     monitor,
	}
}

// Dashboard renders the main dashboard page
func (h *Handlers) Dashboard(c *gin.Context) {
	services, err := h.serviceRepo.GetAll(c.Request.Context())
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load services",
		})
		return
	}

	// Get status counts for dashboard stats
	statusCounts, err := h.getStatusCounts(c.Request.Context())
	if err != nil {
		statusCounts = map[string]int{}
	}

	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title":         "Pipeline Monitor Dashboard",
		"services":      services,
		"statusCounts":  statusCounts,
		"totalServices": len(services),
	})
}

// ListServices returns the services list page
func (h *Handlers) ListServices(c *gin.Context) {
	services, err := h.serviceRepo.GetAll(c.Request.Context())
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to load services",
		})
		return
	}

	c.HTML(http.StatusOK, "services/list.html", gin.H{
		"title":    "Services",
		"services": services,
	})
}

// NewServiceForm shows the form to create a new service
func (h *Handlers) NewServiceForm(c *gin.Context) {
	c.HTML(http.StatusOK, "services/form.html", gin.H{
		"title":   "Add New Service",
		"service": &service.Service{}, // Empty service for new form
		"isEdit":  false,
	})
}

// CreateService handles service creation
func (h *Handlers) CreateService(c *gin.Context) {
	var req struct {
		Name        string   `form:"name" binding:"required"`
		URL         string   `form:"url" binding:"required,url"`
		Description string   `form:"description"`
		Tags        []string `form:"tags"`
	}

	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusBadRequest, "services/form.html", gin.H{
			"title": "Add New Service",
			"error": "Invalid form data: " + err.Error(),
			"service": &service.Service{
				Name:        req.Name,
				URL:         req.URL,
				Description: req.Description,
				Tags:        req.Tags,
			},
			"isEdit": false,
		})
		return
	}

	newService := &service.Service{
		ID:          uuid.New().String(),
		Name:        req.Name,
		URL:         req.URL,
		Description: req.Description,
		Tags:        req.Tags,
		Status:      service.StatusUnknown,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := h.serviceRepo.Create(c.Request.Context(), newService); err != nil {
		c.HTML(http.StatusInternalServerError, "services/form.html", gin.H{
			"title":   "Add New Service",
			"error":   "Failed to create service: " + err.Error(),
			"service": newService,
			"isEdit":  false,
		})
		return
	}

	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") == "true" {
		// Return updated services table
		services, _ := h.serviceRepo.GetAll(c.Request.Context())
		c.HTML(http.StatusOK, "partials/services-table.html", gin.H{
			"services": services,
		})
		return
	}

	// Regular redirect for non-HTMX requests
	c.Redirect(http.StatusSeeOther, "/services")
}

// GetService shows a single service
func (h *Handlers) GetService(c *gin.Context) {
	id := c.Param("id")
	svc, err := h.serviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Service not found",
		})
		return
	}

	c.HTML(http.StatusOK, "services/detail.html", gin.H{
		"title":   "Service: " + svc.Name,
		"service": svc,
	})
}

// EditServiceForm shows the edit form for a service
func (h *Handlers) EditServiceForm(c *gin.Context) {
	id := c.Param("id")
	svc, err := h.serviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Service not found",
		})
		return
	}

	c.HTML(http.StatusOK, "services/form.html", gin.H{
		"title":   "Edit Service: " + svc.Name,
		"service": svc,
		"isEdit":  true,
	})
}

// UpdateService handles service updates
func (h *Handlers) UpdateService(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Name        string   `form:"name" binding:"required"`
		URL         string   `form:"url" binding:"required,url"`
		Description string   `form:"description"`
		Tags        []string `form:"tags"`
	}

	if err := c.ShouldBind(&req); err != nil {
		svc, _ := h.serviceRepo.GetByID(c.Request.Context(), id)
		c.HTML(http.StatusBadRequest, "services/form.html", gin.H{
			"title":   "Edit Service",
			"error":   "Invalid form data: " + err.Error(),
			"service": svc,
			"isEdit":  true,
		})
		return
	}

	// Get existing service
	svc, err := h.serviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Service not found",
		})
		return
	}

	// Update fields
	svc.Name = req.Name
	svc.URL = req.URL
	svc.Description = req.Description
	svc.Tags = req.Tags
	svc.UpdatedAt = time.Now()

	if err := h.serviceRepo.Update(c.Request.Context(), svc); err != nil {
		c.HTML(http.StatusInternalServerError, "services/form.html", gin.H{
			"title":   "Edit Service",
			"error":   "Failed to update service: " + err.Error(),
			"service": svc,
			"isEdit":  true,
		})
		return
	}

	// Check if this is an HTMX request
	if c.GetHeader("HX-Request") == "true" {
		// Return updated service row
		c.HTML(http.StatusOK, "partials/service-row.html", gin.H{
			"service": svc,
		})
		return
	}

	c.Redirect(http.StatusSeeOther, "/services")
}

// DeleteService handles service deletion
func (h *Handlers) DeleteService(c *gin.Context) {
	id := c.Param("id")

	err := h.serviceRepo.Delete(c.Request.Context(), id)
	if err != nil {
		if c.GetHeader("HX-Request") == "true" {
			c.Header("HX-Trigger", "error")
			c.Status(http.StatusInternalServerError)
			return
		}
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to delete service: " + err.Error(),
		})
		return
	}

	// For HTMX requests, return empty content (the row will be removed)
	if c.GetHeader("HX-Request") == "true" {
		c.Status(http.StatusOK)
		return
	}

	c.Redirect(http.StatusSeeOther, "/services")
}

// HTMX Partial Handlers
// These return HTML fragments for dynamic updates

// ServiceStatusPartial returns just the status component for a service
func (h *Handlers) ServiceStatusPartial(c *gin.Context) {
	id := c.Param("id")
	svc, err := h.serviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.HTML(http.StatusNotFound, "partials/service-status.html", gin.H{
			"error": "Service not found",
		})
		return
	}

	c.HTML(http.StatusOK, "partials/service-status.html", gin.H{
		"service": svc,
	})
}

// ServicesTablePartial returns the complete services table
func (h *Handlers) ServicesTablePartial(c *gin.Context) {
	services, err := h.serviceRepo.GetAll(c.Request.Context())
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/services-table.html", gin.H{
			"error": "Failed to load services",
		})
		return
	}

	c.HTML(http.StatusOK, "partials/services-table.html", gin.H{
		"services": services,
	})
}

// DashboardStatsPartial returns the dashboard statistics
func (h *Handlers) DashboardStatsPartial(c *gin.Context) {
	statusCounts, err := h.getStatusCounts(c.Request.Context())
	if err != nil {
		c.HTML(http.StatusInternalServerError, "partials/dashboard-stats.html", gin.H{
			"error": "Failed to load statistics",
		})
		return
	}

	services, _ := h.serviceRepo.GetAll(c.Request.Context())

	c.HTML(http.StatusOK, "partials/dashboard-stats.html", gin.H{
		"statusCounts":  statusCounts,
		"totalServices": len(services),
	})
}

// API Handlers (JSON responses for external consumption)

// APIListServices returns services as JSON
func (h *Handlers) APIListServices(c *gin.Context) {
	services, err := h.serviceRepo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch services",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"services": services,
		"count":    len(services),
	})
}

// APIGetService returns a single service as JSON
func (h *Handlers) APIGetService(c *gin.Context) {
	id := c.Param("id")
	svc, err := h.serviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Service not found",
		})
		return
	}

	c.JSON(http.StatusOK, svc)
}

// APICreateService creates a service via JSON API
func (h *Handlers) APICreateService(c *gin.Context) {
	var req service.Service
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON: " + err.Error(),
		})
		return
	}

	req.ID = uuid.New().String()
	req.Status = service.StatusUnknown
	req.CreatedAt = time.Now()
	req.UpdatedAt = time.Now()

	if err := h.serviceRepo.Create(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create service: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, req)
}

// APIUpdateService updates a service via JSON API
func (h *Handlers) APIUpdateService(c *gin.Context) {
	id := c.Param("id")

	svc, err := h.serviceRepo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Service not found",
		})
		return
	}

	if err := c.ShouldBindJSON(svc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid JSON: " + err.Error(),
		})
		return
	}

	svc.UpdatedAt = time.Now()

	if err := h.serviceRepo.Update(c.Request.Context(), svc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update service: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, svc)
}

// APIDeleteService deletes a service via JSON API
func (h *Handlers) APIDeleteService(c *gin.Context) {
	id := c.Param("id")

	if err := h.serviceRepo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete service: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Service deleted successfully",
	})
}

// APIHealthCheck returns API health status
func (h *Handlers) APIHealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	})
}

// ServiceUpdatesSSE provides Server-Sent Events for real-time updates
func (h *Handlers) ServiceUpdatesSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	// Get the update channel from monitor
	updates := h.monitor.GetUpdates()

	// Send initial ping
	c.Writer.Write([]byte("data: {\"type\":\"ping\"}\n\n"))
	c.Writer.Flush()

	// Listen for updates
	for {
		select {
		case update, ok := <-updates:
			if !ok {
				return
			}

			// Send update as SSE
			data := fmt.Sprintf("data: {\"type\":\"service_update\",\"service_id\":\"%s\",\"status\":\"%s\",\"response_time\":%d}\n\n",
				update.ServiceID, update.Status, update.ResponseTime)
			c.Writer.Write([]byte(data))
			c.Writer.Flush()

		case <-c.Request.Context().Done():
			return
		}
	}
}

// getStatusCounts returns counts for each service status
func (h *Handlers) getStatusCounts(ctx context.Context) (map[string]int, error) {
	services, err := h.serviceRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	counts := map[string]int{
		"healthy":   0,
		"unhealthy": 0,
		"timeout":   0,
		"unknown":   0,
	}

	for _, svc := range services {
		counts[string(svc.Status)]++
	}

	return counts, nil
}
