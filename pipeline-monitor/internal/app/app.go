package app

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"pipeline-monitor/internal/config"
	"pipeline-monitor/internal/domain/service"
	"pipeline-monitor/internal/handlers"
	"pipeline-monitor/internal/infrastructure/database"
	"pipeline-monitor/internal/infrastructure/monitor"

	"github.com/gin-gonic/gin"
)

// Application holds all the application dependencies
type Application struct {
	config      *config.Config
	serviceRepo service.Repository
	monitor     *monitor.ServiceMonitor
	handlers    *handlers.Handlers
	router      *gin.Engine
}

// New creates a new application instance with all dependencies wired up
func New(cfg *config.Config) *Application {
	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Database connection
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Create database schema
	if err := database.CreateSchema(db); err != nil {
		log.Fatal("Failed to create database schema:", err)
	}

	// Repository layer
	serviceRepo := database.NewServiceRepository(db)

	// Service monitor (this is where Go concurrency shines)
	serviceMonitor := monitor.New(serviceRepo, cfg.CheckInterval)

	// Handlers
	handlers := handlers.New(serviceRepo, serviceMonitor)

	// Create application instance
	app := &Application{
		config:      cfg,
		serviceRepo: serviceRepo,
		monitor:     serviceMonitor,
		handlers:    handlers,
	}

	// Setup router
	app.setupRouter()

	// Start monitoring
	if err := serviceMonitor.Start(); err != nil {
		log.Fatal("Failed to start service monitor:", err)
	}

	return app
}

// Router returns the configured HTTP router
func (a *Application) Router() http.Handler {
	return a.router
}

// Shutdown gracefully shuts down the application
func (a *Application) Shutdown(ctx context.Context) error {
	log.Println("Shutting down application...")

	// Stop the service monitor
	if err := a.monitor.Stop(); err != nil {
		log.Printf("Error stopping monitor: %v", err)
	}

	return nil
}

// setupRouter configures all routes and middleware
func (a *Application) setupRouter() {
	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(a.corsMiddleware())

	// Load HTML templates
	router.SetHTMLTemplate(a.loadTemplates())

	// Static files
	router.Static("/static", "./templates/static")

	// Routes
	a.setupRoutes(router)

	a.router = router
}

// setupRoutes defines all application routes
func (a *Application) setupRoutes(router *gin.Engine) {
	// Dashboard routes
	router.GET("/", a.handlers.Dashboard)
	router.GET("/dashboard", a.handlers.Dashboard)

	// Service management routes
	router.GET("/services", a.handlers.ListServices)
	router.GET("/services/new", a.handlers.NewServiceForm)
	router.POST("/services", a.handlers.CreateService)
	router.GET("/services/:id", a.handlers.GetService)
	router.GET("/services/:id/edit", a.handlers.EditServiceForm)
	router.PUT("/services/:id", a.handlers.UpdateService)
	router.DELETE("/services/:id", a.handlers.DeleteService)

	// HTMX partial routes for real-time updates
	router.GET("/partials/service-status/:id", a.handlers.ServiceStatusPartial)
	router.GET("/partials/services-table", a.handlers.ServicesTablePartial)
	router.GET("/partials/dashboard-stats", a.handlers.DashboardStatsPartial)

	// API routes for external access
	api := router.Group("/api/v1")
	{
		api.GET("/services", a.handlers.APIListServices)
		api.GET("/services/:id", a.handlers.APIGetService)
		api.POST("/services", a.handlers.APICreateService)
		api.PUT("/services/:id", a.handlers.APIUpdateService)
		api.DELETE("/services/:id", a.handlers.APIDeleteService)
		api.GET("/health", a.handlers.APIHealthCheck)
	}

	// Server-Sent Events for real-time updates
	router.GET("/events/service-updates", a.handlers.ServiceUpdatesSSE)
}

// loadTemplates loads and parses HTML templates
func (a *Application) loadTemplates() *template.Template {
	tmpl := template.New("")

	// Template functions for HTMX integration
	tmpl.Funcs(template.FuncMap{
		"statusClass": func(status string) string {
			switch status {
			case "healthy":
				return "status-healthy"
			case "unhealthy":
				return "status-unhealthy"
			case "timeout":
				return "status-timeout"
			default:
				return "status-unknown"
			}
		},
		"formatResponseTime": func(responseTime int) string {
			if responseTime < 1000 {
				return fmt.Sprintf("%dms", responseTime)
			}
			return fmt.Sprintf("%.1fs", float64(responseTime)/1000)
		},
	})

	// Load template files
	templateFiles := []string{
		"templates/base.html",
		"templates/dashboard.html",
		"templates/error.html",
		"templates/services/list.html",
		"templates/services/form.html",
		"templates/services/detail.html",
		"templates/partials/service-status.html",
		"templates/partials/services-table.html",
		"templates/partials/service-row.html",
		"templates/partials/dashboard-stats.html",
	}

	for _, file := range templateFiles {
		if _, err := tmpl.ParseFiles(file); err != nil {
			log.Printf("Warning: Could not load template %s: %v", file, err)
		}
	}

	return tmpl
}

// corsMiddleware adds CORS headers
func (a *Application) corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
