package router

import (
	"github.com/jamesdube/ussd/internal/utils"
	"github.com/jamesdube/ussd/pkg/menu"
)

// RouteType represents different types of routes for optimization
type RouteType int

const (
	ExactRoute    RouteType = iota // "1.2.3" - no wildcards
	WildcardRoute                  // "1.*.3" - contains wildcards
	LongCodeRoute                  // "*123*1*1*100#" - USSD long codes
)

// CompiledRoute represents a pre-parsed route for faster matching
type CompiledRoute struct {
	Pattern     []string  // Pre-parsed route segments: ["1", "*", "3"]
	Menu        menu.Menu // The menu to route to
	Type        RouteType // Route type for optimization
	MinDepth    int       // Minimum required selections
	MaxDepth    int       // Maximum allowed selections
	HasWildcard bool      // Quick wildcard check
}

// Router handles USSD route matching with compiled routes for performance
type Router struct {
	// Compiled routes grouped by type for faster matching
	exactRoutes    []CompiledRoute // Routes with no wildcards (fastest to match)
	wildcardRoutes []CompiledRoute // Routes with wildcards (slower but flexible)
	longCodeRoutes []CompiledRoute // Routes for USSD long codes (e.g., *123*1*1*100#)

	// Legacy field - kept for backward compatibility but not used
	menus []menu.Menu
}

// NewRouter creates a new router with compiled route optimization
func NewRouter() *Router {
	return &Router{
		exactRoutes:    make([]CompiledRoute, 0),
		wildcardRoutes: make([]CompiledRoute, 0),
		longCodeRoutes: make([]CompiledRoute, 0),
		menus:          []menu.Menu{},
	}
}

// AddRoute adds a route with pre-compilation for faster lookup
func (r *Router) AddRoute(route string, menu menu.Menu) {
	// Pre-compile the route pattern
	compiledRoute := compileRoute(route, menu)

	// Group by type for optimized matching
	switch compiledRoute.Type {
	case ExactRoute:
		r.exactRoutes = append(r.exactRoutes, compiledRoute)
	case WildcardRoute:
		r.wildcardRoutes = append(r.wildcardRoutes, compiledRoute)
	case LongCodeRoute:
		r.longCodeRoutes = append(r.longCodeRoutes, compiledRoute)
	}
}

// RouteTo finds a matching route using compiled patterns (no string parsing during lookup)
func (r *Router) RouteTo(selections []string) menu.Menu {
	if selections == nil {
		return nil
	}

	selectionDepth := len(selections)

	// 1. First check exact routes (fastest matching)
	for _, route := range r.exactRoutes {
		// Quick depth check before expensive comparison
		if selectionDepth != route.MaxDepth {
			continue
		}

		if checkCompiled(route.Pattern, selections) {
			utils.Logger.Debug("routing to exact route", "selections", selections)
			return route.Menu
		}
	}

	// 2. Then check long code routes (specific long codes should match before wildcards)
	for _, route := range r.longCodeRoutes {
		// Quick depth check
		if selectionDepth < route.MinDepth || selectionDepth > route.MaxDepth {
			continue
		}

		if checkCompiled(route.Pattern, selections) {
			utils.Logger.Debug("routing to long code route", "selections", selections)
			return route.Menu
		}
	}

	// 3. Finally check wildcard routes (most flexible but lowest priority)
	for _, route := range r.wildcardRoutes {
		// Quick depth check
		if selectionDepth < route.MinDepth || selectionDepth > route.MaxDepth {
			continue
		}

		if checkCompiled(route.Pattern, selections) {
			utils.Logger.Debug("routing to wildcard route", "selections", selections)
			return route.Menu
		}
	}

	return nil
}

// RouteToFromMessage handles both regular selections and long codes
func (r *Router) RouteToFromMessage(message string) menu.Menu {
	// Check if it's a long code first
	if utils.IsLongCode(message) {
		components := utils.ParseLongCode(message)
		if components != nil {
			utils.Logger.Debug("processing long code", "message", message, "components", components)
			return r.RouteTo(components)
		}
	}

	// Not a long code, treat as single selection
	return r.RouteTo([]string{message})
}

// compileRoute pre-processes a route string into optimized structure
func compileRoute(route string, menu menu.Menu) CompiledRoute {
	// Check if this is a long code route pattern (starts with * and ends with #)
	if utils.IsLongCode(route) {
		// Parse the long code into components
		components := utils.ParseLongCode(route)
		return CompiledRoute{
			Pattern:     components,
			Menu:        menu,
			Type:        LongCodeRoute,
			MinDepth:    len(components),
			MaxDepth:    len(components),
			HasWildcard: false,
		}
	}

	// Parse regular route pattern once at registration time
	pattern := utils.StringToSlice(route)

	// Analyze route characteristics
	hasWildcard := false
	minDepth := len(pattern)
	maxDepth := len(pattern)

	for _, segment := range pattern {
		if segment == "*" {
			hasWildcard = true
			break
		}
	}

	// Determine route type
	routeType := ExactRoute
	if hasWildcard {
		routeType = WildcardRoute
	}

	return CompiledRoute{
		Pattern:     pattern,
		Menu:        menu,
		Type:        routeType,
		MinDepth:    minDepth,
		MaxDepth:    maxDepth,
		HasWildcard: hasWildcard,
	}
}

// checkCompiled performs route matching using pre-compiled patterns (no allocations)
func checkCompiled(routePattern []string, selections []string) bool {
	// Length already checked by caller for performance
	if len(routePattern) != len(selections) {
		return false
	}

	// Compare segments directly (no string parsing)
	for i, routeSegment := range routePattern {
		if routeSegment != "*" && routeSegment != selections[i] {
			return false
		}
	}

	return true
}

// Legacy functions maintained for backward compatibility

func check(k []string, v []string) bool {
	return checkCompiled(k, v)
}

func comparator(v1 string, v2 string) bool {
	if v1 == v2 {
		return true
	}

	if v1 == "*" {
		return true
	}

	return false
}
