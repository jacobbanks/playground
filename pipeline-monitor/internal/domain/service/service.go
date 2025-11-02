package service

import (
	"context"
	"time"
)

// Service represents a monitored service in our pipeline
type Service struct {
	ID           string    `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	URL          string    `json:"url" db:"url"`
	Status       Status    `json:"status" db:"status"`
	LastCheck    time.Time `json:"last_check" db:"last_check"`
	ResponseTime int       `json:"response_time" db:"response_time"` // milliseconds
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	Description  string    `json:"description" db:"description"`
	Tags         []string  `json:"tags" db:"tags"`
}

// Status represents the health status of a service
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusUnknown   Status = "unknown"
	StatusTimeout   Status = "timeout"
)

// String returns the string representation of status
func (s Status) String() string {
	return string(s)
}

// IsHealthy returns true if the service is in a healthy state
func (s Status) IsHealthy() bool {
	return s == StatusHealthy
}

// Repository defines what our service layer needs from the data layer
// This is Go's way of dependency inversion - interfaces are defined by consumers
type Repository interface {
	GetAll(ctx context.Context) ([]Service, error)
	GetByID(ctx context.Context, id string) (*Service, error)
	Create(ctx context.Context, service *Service) error
	Update(ctx context.Context, service *Service) error
	Delete(ctx context.Context, id string) error
	UpdateStatus(ctx context.Context, id string, status Status, responseTime int) error
}

// HealthCheck represents a single health check result
type HealthCheck struct {
	ServiceID    string    `json:"service_id"`
	Status       Status    `json:"status"`
	ResponseTime int       `json:"response_time"`
	Timestamp    time.Time `json:"timestamp"`
	Error        string    `json:"error,omitempty"`
}

// ServiceManager defines the interface for managing services
type ServiceManager interface {
	StartMonitoring(ctx context.Context) error
	StopMonitoring() error
	AddService(ctx context.Context, service *Service) error
	RemoveService(ctx context.Context, id string) error
	GetHealthStatus(ctx context.Context, id string) (*HealthCheck, error)
}
