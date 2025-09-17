# USSD Framework

A powerful and flexible Go framework for building USSD (Unstructured Supplementary Service Data) applications with support for multiple gateways, session management, and menu navigation.

## Features

- 🚀 **High Performance**: Built on top of Fiber web framework for fast HTTP handling
- 🔄 **Session Management**: Multiple session storage backends (Redis, Hazelcast, In-Memory)
- 🌐 **Gateway Support**: Pluggable gateway system with built-in Econet support
- 📱 **Menu Navigation**: Intuitive menu system with pagination support
- 🔧 **Middleware Support**: Extensible middleware system for request/response processing
- 📊 **Monitoring**: Built-in Prometheus metrics support
- 🔍 **Logging**: Structured logging with Zap
- ⚙️ **Configuration**: Flexible configuration with YAML and environment variables

## Installation

```bash
go get github.com/jamesdube/ussd
```

## Quick Start

### Basic USSD Application

```go
package main

import (
    "log/slog"
    "github.com/jamesdube/ussd/pkg/ussd"
    "github.com/jamesdube/ussd/pkg/menu"
)

// Define a simple menu
type MainMenu struct{}

func (m *MainMenu) OnRequest(ctx *menu.Context, msg string) menu.Response {
    return menu.Response{
        Prompt: "Welcome to USSD Service\n1. Check Balance\n2. Buy Airtime\n3. Exit",
        Options: []string{"1", "2", "3"},
    }
}

func (m *MainMenu) Process(ctx *menu.Context, msg string) menu.NavigationType {
    switch msg {
    case "1":
        return menu.Forward
    case "2":
        return menu.Forward
    case "3":
        return menu.End
    default:
        return menu.Replay
    }
}

func main() {
    // Create logger
    logger := slog.Default()
    
    // Initialize USSD application
    app := ussd.New(ussd.Config{
        AppName: "My USSD Service",
        Port:    8080,
        Logger:  logger,
    })
    
    // Register menus
    app.AddMenu("main", &MainMenu{})
    
    // Start the application
    app.Start()
}
```

### Menu with Pagination

```go
type PaginatedMenu struct{}

func (m *PaginatedMenu) OnRequest(ctx *menu.Context, msg string) menu.Response {
    items := []string{
        "Item 1", "Item 2", "Item 3", "Item 4", "Item 5",
        "Item 6", "Item 7", "Item 8", "Item 9", "Item 10",
    }
    
    return menu.Response{
        Prompt:    "Select an item:",
        Options:   items,
        Paginated: true,
        PerPage:   3,
    }
}

func (m *PaginatedMenu) Process(ctx *menu.Context, msg string) menu.NavigationType {
    // Handle pagination logic
    return menu.Forward
}
```

## Configuration

### YAML Configuration (config.yaml)

```yaml
app:
  port: 8080

cluster:
  provider: "redis"  # Options: redis, hazelcast, memory

menu:
  navigation:
    main: "MainMenu"
    balance: "BalanceMenu"
    airtime: "AirtimeMenu"
```

### Environment Variables

Create a `.env` file in your project root:

```env
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
HAZELCAST_HOST=localhost
HAZELCAST_PORT=5701
```

## Architecture

### Core Components

#### 1. Framework
The main framework orchestrates all components:
- Router for menu navigation
- Gateway registry for handling different USSD providers
- Session management
- Menu registry
- Middleware pipeline

#### 2. Gateways
Gateways handle communication with USSD service providers:

```go
type Gateway interface {
    ToRequest(b *fiber.Ctx) (Request, error)
    Request() Request
    ToResponse(response Response) interface{}
    Name() string
}
```

#### 3. Sessions
Session management with multiple storage backends:

```go
// Session structure
type Session struct {
    Id               string
    Attributes       map[string]string
    Selections       []string
    Active           bool
    Paginated        bool
    Pages            [][]string
    CurrentPage      int
}
```

#### 4. Menus
Menu interface for defining USSD screens:

```go
type Menu interface {
    OnRequest(c *Context, msg string) Response
    Process(ctx *Context, msg string) NavigationType
}
```

#### 5. Middleware
Extensible middleware system:

```go
type Middleware interface {
    Process(ctx *Context, next func())
}
```

## Session Storage Backends

### Redis
```go
// Redis configuration in config.yaml
cluster:
  provider: "redis"
```

### Hazelcast
```go
// Hazelcast configuration in config.yaml
cluster:
  provider: "hazelcast"
```

### In-Memory
```go
// In-memory configuration in config.yaml
cluster:
  provider: "memory"
```

## Gateway Integration

### Econet Gateway
Built-in support for Econet USSD gateway:

```go
// The framework automatically handles Econet gateway requests
// Gateway converts HTTP requests to internal Request format
type Request struct {
    SessionId         string
    Message          string
    Msisdn           string
    Stage            string
    DestinationNumber string
}
```

### Custom Gateway
Implement your own gateway:

```go
type CustomGateway struct{}

func (g *CustomGateway) ToRequest(ctx *fiber.Ctx) (gateway.Request, error) {
    // Parse incoming request from your USSD provider
    return gateway.Request{
        SessionId: ctx.Get("session-id"),
        Message:   ctx.FormValue("message"),
        Msisdn:    ctx.FormValue("msisdn"),
    }, nil
}

func (g *CustomGateway) ToResponse(response gateway.Response) interface{} {
    // Format response for your USSD provider
    return map[string]interface{}{
        "message": response.Message,
        "session": response.Session,
        "active":  response.SessionActive,
    }
}

func (g *CustomGateway) Name() string {
    return "custom"
}

func (g *CustomGateway) Request() gateway.Request {
    return gateway.Request{}
}
```

## Menu Navigation

### Navigation Types
- `Forward`: Move to next menu
- `Backward`: Go back to previous menu
- `Replay`: Stay on current menu
- `End`: Terminate session

### Menu Context
Access session data and user information:

```go
func (m *MyMenu) OnRequest(ctx *menu.Context, msg string) menu.Response {
    // Access user phone number
    msisdn := ctx.Msisdn
    
    // Store data in session
    ctx.Add("user_selection", msg)
    
    // Retrieve stored data
    previousSelection := ctx.Get("previous_selection")
    
    return menu.Response{
        Prompt: fmt.Sprintf("Hello %s, you selected: %s", msisdn, msg),
    }
}
```

## Middleware

### Custom Middleware
```go
type LoggingMiddleware struct{}

func (m *LoggingMiddleware) Process(ctx *menu.Context, next func()) {
    log.Printf("Processing request for MSISDN: %s", ctx.Msisdn)
    next()
    log.Printf("Request processed")
}

// Register middleware
app.AddMiddleware(&LoggingMiddleware{})
```

## Monitoring

The framework includes built-in Prometheus metrics support for monitoring:
- Request count
- Response time
- Active sessions
- Error rates

## API Reference

### USSD Config
```go
type Config struct {
    AppName    string        // Application name
    Port       int          // HTTP server port
    HideBanner bool         // Hide Fiber banner
    Logger     *slog.Logger // Structured logger
}
```

### Menu Response
```go
type Response struct {
    Prompt         string           // Text to display to user
    Options        []string         // Available options
    Paginated      bool            // Enable pagination
    PerPage        int             // Items per page
    NavigationType NavigationType   // Navigation behavior
}
```

### Session Methods
```go
// Session management
session.AddSelection(message string)
session.RemoveLastSelection()
session.GetSelections() []string
session.GetID() string
```

## Examples

### Banking USSD Service
```go
type BankingMenu struct{}

func (m *BankingMenu) OnRequest(ctx *menu.Context, msg string) menu.Response {
    return menu.Response{
        Prompt: "Banking Services\n1. Check Balance\n2. Transfer Money\n3. Mini Statement\n0. Exit",
        Options: []string{"1", "2", "3", "0"},
    }
}

type BalanceMenu struct{}

func (m *BalanceMenu) OnRequest(ctx *menu.Context, msg string) menu.Response {
    // Simulate balance check
    balance := "USD 1,250.00"
    return menu.Response{
        Prompt: fmt.Sprintf("Your balance is: %s\n0. Back to Main Menu", balance),
        Options: []string{"0"},
    }
}
```

### E-commerce USSD Service
```go
type ShoppingMenu struct{}

func (m *ShoppingMenu) OnRequest(ctx *menu.Context, msg string) menu.Response {
    products := []string{
        "Laptop - $999", "Phone - $599", "Tablet - $399",
        "Headphones - $199", "Watch - $299", "Camera - $799",
    }
    
    return menu.Response{
        Prompt:    "Available Products:",
        Options:   products,
        Paginated: true,
        PerPage:   2,
    }
}
```

## Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./pkg/menu
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Dependencies

- [Fiber](https://github.com/gofiber/fiber) - Web framework
- [Redis](https://github.com/go-redis/redis) - Redis client
- [Hazelcast](https://github.com/hazelcast/hazelcast-go-client) - Hazelcast client
- [Viper](https://github.com/spf13/viper) - Configuration management
- [Zap](https://github.com/uber-go/zap) - Structured logging
- [Prometheus](https://github.com/prometheus/client_golang) - Metrics

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

For support and questions:
- Create an issue on GitHub
- Check the documentation
- Review the examples

## Roadmap

- [ ] WebSocket support for real-time updates
- [ ] GraphQL API integration
- [ ] Additional gateway providers
- [ ] Enhanced monitoring dashboard
- [ ] Performance optimizations
- [ ] Docker containerization examples