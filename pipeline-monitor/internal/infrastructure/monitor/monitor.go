package monitor

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"pipeline-monitor/internal/domain/service"
)

// ServiceMonitor handles concurrent monitoring of multiple services
type ServiceMonitor struct {
	repo         service.Repository
	interval     time.Duration
	updates      chan ServiceUpdate
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	client       *http.Client
	activeChecks map[string]context.CancelFunc
	checksMutex  sync.RWMutex
}

// ServiceUpdate represents a status update from a health check
type ServiceUpdate struct {
	ServiceID    string
	Status       service.Status
	ResponseTime int
	Timestamp    time.Time
	Error        error
}

// New creates a new ServiceMonitor instance
func New(repo service.Repository, intervalSeconds int) *ServiceMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &ServiceMonitor{
		repo:         repo,
		interval:     time.Duration(intervalSeconds) * time.Second,
		updates:      make(chan ServiceUpdate, 100), // Buffered channel for non-blocking updates
		ctx:          ctx,
		cancel:       cancel,
		client:       &http.Client{Timeout: 10 * time.Second},
		activeChecks: make(map[string]context.CancelFunc),
	}
}

// Start begins the monitoring process
func (m *ServiceMonitor) Start() error {
	log.Println("Starting service monitor...")

	// Start the main monitoring loop
	m.wg.Add(1)
	go m.monitorLoop()

	// Start the update processor
	m.wg.Add(1)
	go m.processUpdates()

	return nil
}

// Stop gracefully stops the monitoring
func (m *ServiceMonitor) Stop() error {
	log.Println("Stopping service monitor...")

	// Cancel the main context
	m.cancel()

	// Cancel all active health checks
	m.checksMutex.Lock()
	for _, cancelFunc := range m.activeChecks {
		cancelFunc()
	}
	m.checksMutex.Unlock()

	// Wait for all goroutines to finish
	m.wg.Wait()

	// Close the updates channel
	close(m.updates)

	log.Println("Service monitor stopped")
	return nil
}

// GetUpdates returns the channel for receiving service updates
func (m *ServiceMonitor) GetUpdates() <-chan ServiceUpdate {
	return m.updates
}

// monitorLoop runs the main monitoring cycle
func (m *ServiceMonitor) monitorLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// Perform initial check
	m.checkAllServices()

	for {
		select {
		case <-ticker.C:
			m.checkAllServices()
		case <-m.ctx.Done():
			log.Println("Monitor loop stopping...")
			return
		}
	}
}

// checkAllServices fetches all services and starts concurrent health checks
func (m *ServiceMonitor) checkAllServices() {
	services, err := m.repo.GetAll(m.ctx)
	if err != nil {
		log.Printf("Error fetching services: %v", err)
		return
	}

	log.Printf("Checking %d services...", len(services))

	// Fan-out pattern: check all services concurrently
	for _, svc := range services {
		m.wg.Add(1)
		go m.checkService(svc)
	}
}

// checkService performs a health check on a single service
func (m *ServiceMonitor) checkService(svc service.Service) {
	defer m.wg.Done()

	// Create a context for this specific check with timeout
	checkCtx, cancel := context.WithTimeout(m.ctx, 8*time.Second)
	defer cancel()

	// Store the cancel function for potential early termination
	m.checksMutex.Lock()
	m.activeChecks[svc.ID] = cancel
	m.checksMutex.Unlock()

	// Clean up when done
	defer func() {
		m.checksMutex.Lock()
		delete(m.activeChecks, svc.ID)
		m.checksMutex.Unlock()
	}()

	start := time.Now()
	status, err := m.performHealthCheck(checkCtx, svc.URL)
	responseTime := int(time.Since(start).Milliseconds())

	// Send update through channel (non-blocking due to buffer)
	select {
	case m.updates <- ServiceUpdate{
		ServiceID:    svc.ID,
		Status:       status,
		ResponseTime: responseTime,
		Timestamp:    time.Now(),
		Error:        err,
	}:
	case <-m.ctx.Done():
		return
	default:
		// Channel is full, log and continue
		log.Printf("Update channel full, dropping update for service %s", svc.ID)
	}
}

// performHealthCheck makes an HTTP request to check service health
func (m *ServiceMonitor) performHealthCheck(ctx context.Context, url string) (service.Status, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return service.StatusUnknown, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		// Check if it's a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return service.StatusTimeout, fmt.Errorf("request timeout: %w", err)
		}
		return service.StatusUnhealthy, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Consider 200-399 as healthy
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return service.StatusHealthy, nil
	}

	return service.StatusUnhealthy, fmt.Errorf("unhealthy status code: %d", resp.StatusCode)
}

// processUpdates handles incoming service updates
func (m *ServiceMonitor) processUpdates() {
	defer m.wg.Done()

	for {
		select {
		case update, ok := <-m.updates:
			if !ok {
				// Channel closed, exit
				return
			}
			m.handleUpdate(update)
		case <-m.ctx.Done():
			return
		}
	}
}

// handleUpdate processes a single service update
func (m *ServiceMonitor) handleUpdate(update ServiceUpdate) {
	// Update the database
	err := m.repo.UpdateStatus(m.ctx, update.ServiceID, update.Status, update.ResponseTime)
	if err != nil {
		log.Printf("Failed to update service %s status: %v", update.ServiceID, err)
		return
	}

	// Log the update
	if update.Error != nil {
		log.Printf("Service %s: %s (error: %v)", update.ServiceID, update.Status, update.Error)
	} else {
		log.Printf("Service %s: %s (%dms)", update.ServiceID, update.Status, update.ResponseTime)
	}
}
