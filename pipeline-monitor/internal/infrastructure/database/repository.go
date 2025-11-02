package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"pipeline-monitor/internal/domain/service"

	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

// ServiceRepository implements the service.Repository interface using PostgreSQL
type ServiceRepository struct {
	db *sql.DB
}

// NewServiceRepository creates a new service repository
func NewServiceRepository(db *sql.DB) *ServiceRepository {
	return &ServiceRepository{db: db}
}

// Connect establishes a database connection
func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// CreateSchema creates the necessary database tables
func CreateSchema(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS services (
		id VARCHAR(36) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		url VARCHAR(512) NOT NULL,
		status VARCHAR(50) DEFAULT 'unknown',
		last_check TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		response_time INTEGER DEFAULT 0,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		description TEXT,
		tags TEXT[]
	);

	CREATE INDEX IF NOT EXISTS idx_services_status ON services(status);
	CREATE INDEX IF NOT EXISTS idx_services_last_check ON services(last_check);
	`

	_, err := db.Exec(query)
	return err
}

// GetAll retrieves all services from the database
func (r *ServiceRepository) GetAll(ctx context.Context) ([]service.Service, error) {
	query := `
		SELECT id, name, url, status, last_check, response_time,
		       created_at, updated_at, description, tags
		FROM services
		ORDER BY name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	defer rows.Close()

	var services []service.Service
	for rows.Next() {
		var svc service.Service
		err := rows.Scan(
			&svc.ID, &svc.Name, &svc.URL, &svc.Status, &svc.LastCheck,
			&svc.ResponseTime, &svc.CreatedAt, &svc.UpdatedAt,
			&svc.Description, pq.Array(&svc.Tags),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		services = append(services, svc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return services, nil
}

// GetByID retrieves a single service by ID
func (r *ServiceRepository) GetByID(ctx context.Context, id string) (*service.Service, error) {
	query := `
		SELECT id, name, url, status, last_check, response_time,
		       created_at, updated_at, description, tags
		FROM services
		WHERE id = $1
	`

	var svc service.Service
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&svc.ID, &svc.Name, &svc.URL, &svc.Status, &svc.LastCheck,
		&svc.ResponseTime, &svc.CreatedAt, &svc.UpdatedAt,
		&svc.Description, pq.Array(&svc.Tags),
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("service with ID %s not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	return &svc, nil
}

// Create inserts a new service into the database
func (r *ServiceRepository) Create(ctx context.Context, svc *service.Service) error {
	// Generate UUID if not provided
	if svc.ID == "" {
		svc.ID = uuid.New().String()
	}

	query := `
		INSERT INTO services (id, name, url, status, description, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`

	_, err := r.db.ExecContext(ctx, query,
		svc.ID, svc.Name, svc.URL, svc.Status,
		svc.Description, pq.Array(svc.Tags),
	)

	if err != nil {
		return fmt.Errorf("failed to create service: %w", err)
	}

	return nil
}

// Update modifies an existing service
func (r *ServiceRepository) Update(ctx context.Context, svc *service.Service) error {
	query := `
		UPDATE services
		SET name = $2, url = $3, description = $4, tags = $5, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query,
		svc.ID, svc.Name, svc.URL, svc.Description, pq.Array(svc.Tags),
	)
	if err != nil {
		return fmt.Errorf("failed to update service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("service with ID %s not found", svc.ID)
	}

	return nil
}

// Delete removes a service from the database
func (r *ServiceRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM services WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("service with ID %s not found", id)
	}

	return nil
}

// UpdateStatus updates only the status and response time of a service
func (r *ServiceRepository) UpdateStatus(ctx context.Context, id string, status service.Status, responseTime int) error {
	query := `
		UPDATE services
		SET status = $2, response_time = $3, last_check = NOW(), updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id, status, responseTime)
	if err != nil {
		return fmt.Errorf("failed to update service status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("service with ID %s not found", id)
	}

	return nil
}

// GetHealthyCount returns the count of healthy services
func (r *ServiceRepository) GetHealthyCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM services WHERE status = 'healthy'`

	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get healthy count: %w", err)
	}

	return count, nil
}

// GetStatusCounts returns counts for each status
func (r *ServiceRepository) GetStatusCounts(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT status, COUNT(*)
		FROM services
		GROUP BY status
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query status counts: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("failed to scan status count: %w", err)
		}
		counts[status] = count
	}

	return counts, nil
}
