# Pipeline Monitor 🚀

A real-time service monitoring dashboard built with **Go** and **HTMX** to demonstrate modern web development patterns without heavy JavaScript frameworks.

## 🎯 Project Overview

This project showcases the power of combining Go's concurrency features with HTMX's hypermedia-driven approach to create a responsive, real-time web application. It's designed as a learning project to explore:

- **Go**: Goroutines, channels, clean architecture, and concurrent programming
- **HTMX**: Server-side rendering, real-time updates, and progressive enhancement
- **Modern Web Patterns**: Moving from SPA complexity to hypermedia simplicity

## ✨ Features

### Real-time Monitoring
- **Concurrent Health Checks**: Go routines monitor multiple services simultaneously
- **Live Status Updates**: HTMX auto-refreshes service status every 10 seconds
- **Server-Sent Events**: Real-time notifications without WebSocket complexity
- **Progressive Enhancement**: Works without JavaScript, enhanced with HTMX

### Dashboard
- **Live Statistics**: Real-time counts of healthy/unhealthy/timeout services
- **Service Overview**: Visual status indicators with response times
- **Auto-refresh**: Dashboard updates every 30 seconds automatically
- **Responsive Design**: Tailwind CSS for modern, mobile-friendly UI

### Service Management
- **CRUD Operations**: Create, read, update, delete services
- **Form Validation**: Server-side validation with client-side feedback
- **Inline Editing**: Edit services without page refreshes
- **Bulk Operations**: Manage multiple services efficiently

## 🏗️ Architecture

### Go Backend Architecture

```
pipeline-monitor/
├── main.go                         # Application entry point
├── internal/
│   ├── app/                        # Application layer
│   │   └── app.go                  # Dependency injection & routing
│   ├── config/                     # Configuration management
│   │   └── config.go               # Environment-based config
│   ├── domain/                     # Business logic layer
│   │   └── service/
│   │       └── service.go          # Domain models & interfaces
│   ├── handlers/                   # HTTP handlers
│   │   └── handlers.go             # HTMX-aware request handlers
│   └── infrastructure/             # External concerns
│       ├── database/               # Database operations
│       │   └── repository.go       # PostgreSQL repository
│       └── monitor/                # Concurrent monitoring
│           └── monitor.go          # Goroutine-based health checks
└── templates/                      # HTML templates
    ├── base.html                   # Base layout with HTMX
    ├── dashboard.html              # Main dashboard
    └── partials/                   # HTMX partial templates
        ├── dashboard-stats.html    # Statistics cards
        ├── services-table.html     # Services list
        └── service-status.html     # Individual service status
```

### Key Architectural Patterns

#### 1. Clean Architecture
- **Domain Layer**: Business logic independent of frameworks
- **Application Layer**: Orchestrates use cases
- **Infrastructure Layer**: Database, monitoring, external services
- **Interface Layer**: HTTP handlers and templates

#### 2. Concurrent Monitoring
```go
// Fan-out pattern for concurrent health checks
for _, service := range services {
    m.wg.Add(1)
    go m.checkService(service)  // Each service checked concurrently
}
```

#### 3. HTMX Integration
```html
<!-- Auto-refreshing service status -->
<div id="services-list"
     hx-get="/partials/services-table"
     hx-trigger="load, every 30s"
     hx-swap="innerHTML">
</div>
```

## 🚦 Go Concurrency Patterns

### Goroutines for Parallel Health Checks
```go
func (m *ServiceMonitor) checkAllServices() {
    services, err := m.repo.GetAll(m.ctx)
    if err != nil {
        log.Printf("Error fetching services: %v", err)
        return
    }

    // Fan-out: check all services concurrently
    for _, svc := range services {
        m.wg.Add(1)
        go m.checkService(svc)
    }
}
```

### Channels for Safe Communication
```go
type ServiceMonitor struct {
    updates chan ServiceUpdate  // Buffered channel for non-blocking updates
    ctx     context.Context     // Cancellation context
    cancel  context.CancelFunc  // Cancel function for graceful shutdown
}
```

### Graceful Shutdown
```go
func gracefulShutdown(server *http.Server, app *app.Application) {
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    app.Shutdown(ctx)           // Stop monitors, close DB connections
    server.Shutdown(ctx)        // Graceful HTTP server shutdown
}
```

## 🌐 HTMX Patterns

### Progressive Enhancement
```html
<!-- Works without JavaScript, enhanced with HTMX -->
<form action="/services" method="post"
      hx-post="/services"
      hx-target="#services-list"
      hx-swap="innerHTML">
    <input name="name" required>
    <input name="url" required>
    <button type="submit">Add Service</button>
</form>
```

### Real-time Updates
```html
<!-- Automatic polling for live updates -->
<div hx-get="/partials/dashboard-stats"
     hx-trigger="load, every 30s"
     hx-target="this"
     hx-swap="innerHTML">
</div>
```

### Optimistic UI
```html
<!-- Immediate feedback with rollback capability -->
<button hx-delete="/services/123"
        hx-target="#service-123"
        hx-swap="outerHTML"
        hx-confirm="Delete this service?">
    Delete
</button>
```

## 🛠️ Technology Stack

### Backend
- **Go 1.21**: Modern Go with generics and improved performance
- **Gin Framework**: Fast HTTP router with middleware support
- **PostgreSQL**: Reliable database with connection pooling
- **Clean Architecture**: Domain-driven design patterns

### Frontend
- **HTMX 2.0**: Hypermedia-driven interactions
- **Tailwind CSS**: Utility-first CSS framework
- **Server-Side Templates**: Go's `html/template` package
- **Progressive Enhancement**: JavaScript-optional design

### Infrastructure
- **Docker**: Containerized deployment
- **PostgreSQL**: Primary database
- **Server-Sent Events**: Real-time updates without WebSockets

## 🚀 Getting Started

### Prerequisites
- Go 1.21 or later
- PostgreSQL 12+ (or Docker)
- Modern web browser

### Installation

1. **Clone and setup**
```bash
cd pipeline-monitor
go mod tidy
```

2. **Database setup**
```bash
# Using Docker
docker run --name postgres -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres:15

# Or use existing PostgreSQL instance
export DATABASE_URL="postgres://user:password@localhost/pipeline_monitor?sslmode=disable"
```

3. **Run the application**
```bash
go run main.go
```

4. **Open your browser**
```
http://localhost:8080
```

### Environment Variables
```bash
PORT=:7777                          # Server port
DATABASE_URL=postgres://...         # PostgreSQL connection string
ENVIRONMENT=development             # Environment (development/production)
LOG_LEVEL=info                      # Logging level
CHECK_INTERVAL=30                   # Health check interval (seconds)
```

## 📊 Key Learning Outcomes

### Go Concepts Demonstrated
- **Goroutines**: Lightweight concurrency for parallel health checks
- **Channels**: Type-safe communication between goroutines
- **Interfaces**: Dependency inversion and testable code
- **Context**: Cancellation and timeout handling
- **Clean Architecture**: Separation of concerns and testability

### HTMX Concepts Demonstrated
- **Hypermedia Controls**: Server-driven UI state management
- **Progressive Enhancement**: Graceful degradation without JavaScript
- **Partial Updates**: Efficient DOM updates without full page reloads
- **Real-time Features**: Server-Sent Events integration
- **Form Enhancement**: Better UX without complex form libraries

### Web Development Patterns
- **Server-Side Rendering**: Templates over JSON APIs
- **Real-time Updates**: SSE over WebSockets for simplicity
- **Progressive Enhancement**: HTML-first, JavaScript-optional
- **Hypermedia-Driven Architecture**: REST as intended

## 🔄 Key Differences from React/Node.js

### State Management
```go
// Go: Server-side state, templates drive UI
type DashboardData struct {
    Services     []Service
    StatusCounts map[string]int
    TotalServices int
}

func (h *Handlers) Dashboard(c *gin.Context) {
    data := h.buildDashboardData(c.Request.Context())
    c.HTML(200, "dashboard.html", data)
}
```

vs React:
```javascript
// React: Client-side state, hooks manage complexity
const [services, setServices] = useState([]);
const [loading, setLoading] = useState(true);
const [error, setError] = useState(null);

useEffect(() => {
    fetchServices().then(setServices).catch(setError);
}, []);
```

### Real-time Updates
```html
<!-- HTMX: Declarative, server-driven -->
<div hx-get="/partials/services"
     hx-trigger="every 30s"
     hx-swap="innerHTML">
</div>
```

vs React:
```javascript
// React: Imperative, client-driven
useEffect(() => {
    const interval = setInterval(() => {
        fetchServices().then(setServices);
    }, 30000);
    return () => clearInterval(interval);
}, []);
```

## 🧪 Testing Strategy

### Unit Tests
```bash
go test ./internal/... -v
```

### Integration Tests
```bash
go test ./internal/handlers/... -v -tags=integration
```

### Load Testing
```bash
# Test concurrent health checks
go test ./internal/infrastructure/monitor/... -bench=.
```

## 🚀 Deployment

### Docker Deployment
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o pipeline-monitor .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/pipeline-monitor .
COPY --from=builder /app/templates ./templates
CMD ["./pipeline-monitor"]
```

### Production Considerations
- **Graceful Shutdown**: Proper cleanup of goroutines and connections
- **Health Checks**: Kubernetes-ready health endpoints
- **Observability**: Structured logging and metrics
- **Security**: Input validation and CSRF protection

## 📈 Performance Benefits

### Go vs Node.js
- **Memory Usage**: Lower memory footprint with efficient garbage collection
- **Concurrency**: True parallelism vs single-threaded event loop
- **CPU Usage**: Better utilization of multi-core systems
- **Startup Time**: Faster application startup

### HTMX vs React
- **Bundle Size**: No JavaScript framework overhead
- **Initial Load**: Faster first contentful paint
- **Real-time**: Simpler server-sent events vs complex WebSocket management
- **SEO**: Full server-side rendering by default

## 🤝 Contributing

This is a learning project, but contributions are welcome!

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## 📚 Further Learning

### Go Resources
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Concurrency Patterns](https://talks.golang.org/2012/concurrency.slide)
- [Clean Architecture in Go](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)

### HTMX Resources
- [HTMX Documentation](https://htmx.org/docs/)
- [Hypermedia Systems](https://hypermedia.systems/)
- [HTMX Essays](https://htmx.org/essays/)

## 📄 License

MIT License - feel free to use this for learning and experimentation!

---

**Happy Learning!** 🎉

This project demonstrates that you can build modern, responsive web applications without the complexity of heavy JavaScript frameworks. Go's simplicity and HTMX's pragmatic approach create a powerful combination for web development.
