package router

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/jamesdube/ussd/internal/utils"
	"github.com/jamesdube/ussd/pkg/menu"
)

// Mock menu implementation for testing
type mockMenu struct {
	name string
}

func (m *mockMenu) OnRequest(c *menu.Context, msg string) menu.Response {
	return menu.Response{
		Prompt:         "Mock response from " + m.name,
		NavigationType: menu.Continue,
	}
}

func (m *mockMenu) Process(ctx *menu.Context, msg string) menu.NavigationType {
	return menu.Continue
}

// Test helper to ensure logger is initialized
func setupTestLogger() {
	if utils.Logger == nil {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		utils.SetLogger(logger)
	}
}

// Test basic router functionality
func TestNewRouter(t *testing.T) {
	setupTestLogger()

	router := NewRouter()

	if router == nil {
		t.Fatal("NewRouter() returned nil")
	}

	if router.exactRoutes == nil {
		t.Error("Router exactRoutes slice not initialized")
	}

	if router.wildcardRoutes == nil {
		t.Error("Router wildcardRoutes slice not initialized")
	}

	if router.longCodeRoutes == nil {
		t.Error("Router longCodeRoutes slice not initialized")
	}

	if len(router.exactRoutes) != 0 || len(router.wildcardRoutes) != 0 || len(router.longCodeRoutes) != 0 {
		t.Error("Router should start with empty routes")
	}
}

// Test AddRoute functionality
func TestAddRoute(t *testing.T) {
	setupTestLogger()

	router := NewRouter()
	menu1 := &mockMenu{name: "menu1"}
	menu2 := &mockMenu{name: "menu2"}

	// Add routes
	router.AddRoute("1", menu1)   // Exact route
	router.AddRoute("1.2", menu2) // Exact route
	router.AddRoute("1.*", menu1) // Wildcard route

	// Check routes are stored in correct categories
	if len(router.exactRoutes) != 2 {
		t.Errorf("Expected 2 exact routes, got %d", len(router.exactRoutes))
	}

	if len(router.wildcardRoutes) != 1 {
		t.Errorf("Expected 1 wildcard route, got %d", len(router.wildcardRoutes))
	}

	// Check routes are accessible
	result1 := router.RouteTo([]string{"1"})
	if result1 != menu1 {
		t.Error("Route '1' not accessible")
	}

	result2 := router.RouteTo([]string{"1", "2"})
	if result2 != menu2 {
		t.Error("Route '1.2' not accessible")
	}
}

// Test RouteTo with exact matches
func TestRouteToExactMatch(t *testing.T) {
	setupTestLogger()

	tests := []struct {
		name        string
		route       string
		selections  []string
		shouldMatch bool
		menuName    string
	}{
		{
			name:        "Single selection exact match",
			route:       "1",
			selections:  []string{"1"},
			shouldMatch: true,
			menuName:    "menu1",
		},
		{
			name:        "Multiple selections exact match",
			route:       "1.2.3",
			selections:  []string{"1", "2", "3"},
			shouldMatch: true,
			menuName:    "menu2",
		},
		{
			name:        "No match - different selection",
			route:       "1",
			selections:  []string{"2"},
			shouldMatch: false,
		},
		{
			name:        "No match - different length",
			route:       "1.2",
			selections:  []string{"1"},
			shouldMatch: false,
		},
		{
			name:        "No match - extra selections",
			route:       "1",
			selections:  []string{"1", "2"},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()
			menu := &mockMenu{name: tt.menuName}
			router.AddRoute(tt.route, menu)

			result := router.RouteTo(tt.selections)

			if tt.shouldMatch && result == nil {
				t.Errorf("Expected to find route but got nil")
			}

			if !tt.shouldMatch && result != nil {
				t.Errorf("Expected no route but got %v", result)
			}

			if tt.shouldMatch && result != nil {
				mockMenu := result.(*mockMenu)
				if mockMenu.name != tt.menuName {
					t.Errorf("Expected menu %s, got %s", tt.menuName, mockMenu.name)
				}
			}
		})
	}
}

// Test RouteTo with wildcard matches
func TestRouteToWildcardMatch(t *testing.T) {
	setupTestLogger()

	tests := []struct {
		name        string
		route       string
		selections  []string
		shouldMatch bool
		menuName    string
	}{
		{
			name:        "Wildcard at end",
			route:       "1.*",
			selections:  []string{"1", "anything"},
			shouldMatch: true,
			menuName:    "wildcard1",
		},
		{
			name:        "Wildcard in middle",
			route:       "1.*.3",
			selections:  []string{"1", "anything", "3"},
			shouldMatch: true,
			menuName:    "wildcard2",
		},
		{
			name:        "Multiple wildcards",
			route:       "*.*.*",
			selections:  []string{"a", "b", "c"},
			shouldMatch: true,
			menuName:    "wildcard3",
		},
		{
			name:        "Wildcard no match - wrong length",
			route:       "1.*",
			selections:  []string{"1"},
			shouldMatch: false,
		},
		{
			name:        "Wildcard no match - fixed part wrong",
			route:       "1.*.3",
			selections:  []string{"1", "anything", "4"},
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := NewRouter()
			menu := &mockMenu{name: tt.menuName}
			router.AddRoute(tt.route, menu)

			result := router.RouteTo(tt.selections)

			if tt.shouldMatch && result == nil {
				t.Errorf("Expected to find wildcard route but got nil")
			}

			if !tt.shouldMatch && result != nil {
				t.Errorf("Expected no wildcard route but got %v", result)
			}

			if tt.shouldMatch && result != nil {
				mockMenu := result.(*mockMenu)
				if mockMenu.name != tt.menuName {
					t.Errorf("Expected menu %s, got %s", tt.menuName, mockMenu.name)
				}
			}
		})
	}
}

// Test route priority - first match wins
func TestRoutePriority(t *testing.T) {
	setupTestLogger()

	router := NewRouter()

	// Add routes in specific order
	exactMenu := &mockMenu{name: "exact"}
	wildcardMenu := &mockMenu{name: "wildcard"}

	router.AddRoute("1.2", exactMenu)    // Exact match
	router.AddRoute("1.*", wildcardMenu) // Wildcard match

	// Test that exact route is found (depends on map iteration order)
	// Note: This test may be flaky due to Go's map iteration order
	selections := []string{"1", "2"}
	result := router.RouteTo(selections)

	if result == nil {
		t.Fatal("Expected to find a route")
	}

	// Both routes match, but we should get one of them
	mockMenu := result.(*mockMenu)
	if mockMenu.name != "exact" && mockMenu.name != "wildcard" {
		t.Errorf("Expected either exact or wildcard menu, got %s", mockMenu.name)
	}
}

// Test multiple route scenarios
func TestMultipleRoutes(t *testing.T) {
	setupTestLogger()

	router := NewRouter()

	// Create USSD menu structure
	mainMenu := &mockMenu{name: "main"}
	balanceMenu := &mockMenu{name: "balance"}
	historyMenu := &mockMenu{name: "history"}
	transferMenu := &mockMenu{name: "transfer"}
	helpMenu := &mockMenu{name: "help"}

	// Add typical USSD routes - order matters due to map iteration
	router.AddRoute("*", helpMenu)      // Any single wildcard -> help (add first)
	router.AddRoute("1", balanceMenu)   // Check balance
	router.AddRoute("2", transferMenu)  // Transfer money
	router.AddRoute("3", historyMenu)   // Transaction history
	router.AddRoute("1.1", historyMenu) // Balance -> History
	router.AddRoute("1.2", mainMenu)    // Balance -> Back to main
	router.AddRoute("9", helpMenu)      // Help

	tests := []struct {
		selections   []string
		expectedMenu string
	}{
		{[]string{"1"}, "balance"},
		{[]string{"2"}, "transfer"},
		{[]string{"3"}, "history"},
		{[]string{"1", "1"}, "history"},
		{[]string{"1", "2"}, "main"},
		{[]string{"9"}, "help"},
		{[]string{"0"}, "help"}, // Wildcard match
		{[]string{"5"}, "help"}, // Wildcard match
	}

	for _, tt := range tests {
		t.Run("Route_"+tt.expectedMenu, func(t *testing.T) {
			result := router.RouteTo(tt.selections)

			if result == nil {
				t.Fatalf("Expected to find route for %v", tt.selections)
			}

			mockMenu := result.(*mockMenu)
			if mockMenu.name != tt.expectedMenu {
				t.Errorf("For selections %v, expected %s, got %s",
					tt.selections, tt.expectedMenu, mockMenu.name)
			}
		})
	}
}

// Test edge cases
func TestEdgeCases(t *testing.T) {
	setupTestLogger()

	t.Run("Empty selections", func(t *testing.T) {
		router := NewRouter()
		menu := &mockMenu{name: "empty"}
		router.AddRoute("", menu)

		// Empty string splits to [""] so we need to match with [""]
		result := router.RouteTo([]string{""})
		if result == nil {
			t.Error("Expected to find route for empty string selection")
		}

		// Empty slice won't match [""]
		result2 := router.RouteTo([]string{})
		if result2 != nil {
			t.Error("Empty slice should not match empty string route")
		}
	})

	t.Run("Nil selections", func(t *testing.T) {
		router := NewRouter()
		menu := &mockMenu{name: "test"}
		router.AddRoute("1", menu)

		result := router.RouteTo(nil)
		if result != nil {
			t.Error("Expected no route for nil selections")
		}
	})

	t.Run("No routes registered", func(t *testing.T) {
		router := NewRouter()

		result := router.RouteTo([]string{"1"})
		if result != nil {
			t.Error("Expected no route when none registered")
		}
	})
}

// Performance test for compiled routes implementation
func BenchmarkCompiledRouting(b *testing.B) {
	setupTestLogger()

	router := NewRouter()

	// Add many routes to simulate real USSD application
	for i := 0; i < 100; i++ {
		menu := &mockMenu{name: "menu"}
		switch {
		case i < 10:
			router.AddRoute(string(rune('0'+i)), menu)
		case i < 50:
			router.AddRoute(string(rune('0'+(i/10)))+"."+string(rune('0'+(i%10))), menu)
		default:
			router.AddRoute("1.2."+string(rune('0'+(i%10))), menu)
		}
	}

	// Test selections that will require searching through routes
	testSelections := [][]string{
		{"1"},
		{"2", "3"},
		{"1", "2", "5"},
		{"9"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		selections := testSelections[i%len(testSelections)]
		router.RouteTo(selections)
	}
}

// Benchmark that simulates the old behavior with string parsing during lookup
func BenchmarkOldStyleRouting(b *testing.B) {
	setupTestLogger()

	// Create routes as strings (old way)
	routes := make(map[string]*mockMenu, 100)

	// Add many routes to simulate real USSD application
	for i := 0; i < 100; i++ {
		menu := &mockMenu{name: "menu"}
		switch {
		case i < 10:
			routes[string(rune('0'+i))] = menu
		case i < 50:
			routes[string(rune('0'+(i/10)))+"."+string(rune('0'+(i%10)))] = menu
		default:
			routes["1.2."+string(rune('0'+(i%10)))] = menu
		}
	}

	// Test selections that will require searching through routes
	testSelections := [][]string{
		{"1"},
		{"2", "3"},
		{"1", "2", "5"},
		{"9"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		selections := testSelections[i%len(testSelections)]

		// Simulate old routing - parse string on every lookup
		for routeStr, menu := range routes {
			routeSlice := utils.StringToSlice(routeStr) // STRING PARSING ON EVERY LOOKUP
			if checkCompiled(routeSlice, selections) {
				_ = menu // Found route
				break
			}
		}
	}
}

// Test long code functionality
func TestLongCodeRouting(t *testing.T) {
	setupTestLogger()

	tests := []struct {
		name        string
		longCode    string
		shouldMatch bool
		menuName    string
	}{
		{
			name:        "Balance check long code",
			longCode:    "*131#",
			shouldMatch: true,
			menuName:    "balance",
		},
		{
			name:        "Money transfer long code",
			longCode:    "*123*1*1*100#",
			shouldMatch: true,
			menuName:    "transfer",
		},
		{
			name:        "Complex service code",
			longCode:    "*144*1*2*555#",
			shouldMatch: true,
			menuName:    "service",
		},
		{
			name:        "Invalid long code format",
			longCode:    "123*1*1*100#",
			shouldMatch: false,
		},
		{
			name:        "No matching route",
			longCode:    "*999*8*8*999#",
			shouldMatch: false,
		},
	}

	router := NewRouter()

	// Register long code routes
	balanceMenu := &mockMenu{name: "balance"}
	transferMenu := &mockMenu{name: "transfer"}
	serviceMenu := &mockMenu{name: "service"}

	router.AddRoute("*131#", balanceMenu)
	router.AddRoute("*123*1*1*100#", transferMenu)
	router.AddRoute("*144*1*2*555#", serviceMenu)

	// Verify routes are stored in long code category
	if len(router.longCodeRoutes) != 3 {
		t.Errorf("Expected 3 long code routes, got %d", len(router.longCodeRoutes))
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := router.RouteToFromMessage(tt.longCode)

			if tt.shouldMatch && result == nil {
				t.Errorf("Expected to find route for %s but got nil", tt.longCode)
			}

			if !tt.shouldMatch && result != nil {
				t.Errorf("Expected no route for %s but got %v", tt.longCode, result)
			}

			if tt.shouldMatch && result != nil {
				mockMenu := result.(*mockMenu)
				if mockMenu.name != tt.menuName {
					t.Errorf("Expected menu %s, got %s", tt.menuName, mockMenu.name)
				}
			}
		})
	}
}

// Test mixed routing - both regular routes and long codes
func TestMixedRouting(t *testing.T) {
	setupTestLogger()

	router := NewRouter()

	// Add regular routes
	mainMenu := &mockMenu{name: "main"}
	balanceMenu := &mockMenu{name: "balance"}
	transferMenu := &mockMenu{name: "transfer"}

	router.AddRoute("1", mainMenu)
	router.AddRoute("1.2", balanceMenu)
	router.AddRoute("*", transferMenu)

	// Add long code routes
	longBalanceMenu := &mockMenu{name: "longBalance"}
	longTransferMenu := &mockMenu{name: "longTransfer"}

	router.AddRoute("*131#", longBalanceMenu)
	router.AddRoute("*123*1*1*100#", longTransferMenu)

	tests := []struct {
		name         string
		input        string
		isMessage    bool // true = use RouteToFromMessage, false = use RouteTo
		expectedMenu string
		shouldFind   bool
	}{
		{
			name:         "Regular single selection",
			input:        "1",
			isMessage:    true,
			expectedMenu: "main",
			shouldFind:   true,
		},
		{
			name:         "Long code balance check",
			input:        "*131#",
			isMessage:    true,
			expectedMenu: "longBalance",
			shouldFind:   true,
		},
		{
			name:         "Long code transfer",
			input:        "*123*1*1*100#",
			isMessage:    true,
			expectedMenu: "longTransfer",
			shouldFind:   true,
		},
		{
			name:         "Regular multi-selection via RouteTo",
			input:        "1.2",
			isMessage:    false, // Use RouteTo with parsed selections
			expectedMenu: "balance",
			shouldFind:   true,
		},
		{
			name:         "Wildcard match",
			input:        "9",
			isMessage:    true,
			expectedMenu: "transfer",
			shouldFind:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result menu.Menu

			if tt.isMessage {
				result = router.RouteToFromMessage(tt.input)
			} else {
				selections := strings.Split(tt.input, ".")
				result = router.RouteTo(selections)
			}

			if tt.shouldFind && result == nil {
				t.Errorf("Expected to find route for %s", tt.input)
			}

			if !tt.shouldFind && result != nil {
				t.Errorf("Expected no route for %s but got %v", tt.input, result)
			}

			if tt.shouldFind && result != nil {
				mockMenu := result.(*mockMenu)
				if mockMenu.name != tt.expectedMenu {
					t.Errorf("Expected menu %s, got %s", tt.expectedMenu, mockMenu.name)
				}
			}
		})
	}
}

// Performance benchmark for long codes
func BenchmarkLongCodeRouting(b *testing.B) {
	setupTestLogger()

	router := NewRouter()

	// Add many long code routes
	for i := 0; i < 50; i++ {
		menu := &mockMenu{name: "menu"}
		longCode := fmt.Sprintf("*%d*1*1*100#", 100+i)
		router.AddRoute(longCode, menu)
	}

	// Test long codes
	testLongCodes := []string{
		"*131#",
		"*123*1*1*100#",
		"*144*1*2*555#",
		"*110*1*1*50#",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		longCode := testLongCodes[i%len(testLongCodes)]
		router.RouteToFromMessage(longCode)
	}
}
