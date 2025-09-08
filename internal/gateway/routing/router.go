package routing

import (
	"context"
	"net/http"
	"strings"

	"github.com/enterprise-status/statuspage/internal/config"
	"github.com/enterprise-status/statuspage/internal/gateway/middleware"
	"github.com/enterprise-status/statuspage/internal/gateway/proxy"
	"github.com/enterprise-status/statuspage/pkg/logger"
	"go.uber.org/zap"
)

// Context key type for route parameters
type routeParamKey string

// Router handles HTTP routing for the API Gateway
type Router struct {
	config     *config.Config
	proxy      *proxy.ReverseProxy
	middleware *middleware.Middleware
	routes     map[string]*Route
	groups     map[string]*RouteGroup
}

// Route represents a single route
type Route struct {
	Method      string
	Path        string
	Handler     http.HandlerFunc
	Middlewares []middleware.MiddlewareFunc
}

// RouteGroup represents a group of routes
type RouteGroup struct {
	prefix      string
	middlewares []middleware.MiddlewareFunc
	routes      []*Route
	parent      *Router
}

// NewRouter creates a new router instance
func NewRouter(cfg *config.Config, proxy *proxy.ReverseProxy, middleware *middleware.Middleware) *Router {
	return &Router{
		config:     cfg,
		proxy:      proxy,
		middleware: middleware,
		routes:     make(map[string]*Route),
		groups:     make(map[string]*RouteGroup),
	}
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Find matching route
	route := r.findRoute(req.Method, req.URL.Path)
	if route == nil {
		http.NotFound(w, req)
		return
	}

	// Execute middlewares and handler
	r.executeRoute(w, req, route)
}

// GET adds a GET route
func (r *Router) GET(path string, handler http.HandlerFunc) {
	r.addRoute("GET", path, handler, nil)
}

// POST adds a POST route
func (r *Router) POST(path string, handler http.HandlerFunc) {
	r.addRoute("POST", path, handler, nil)
}

// PUT adds a PUT route
func (r *Router) PUT(path string, handler http.HandlerFunc) {
	r.addRoute("PUT", path, handler, nil)
}

// DELETE adds a DELETE route
func (r *Router) DELETE(path string, handler http.HandlerFunc) {
	r.addRoute("DELETE", path, handler, nil)
}

// OPTIONS adds an OPTIONS route
func (r *Router) OPTIONS(path string, handler http.HandlerFunc) {
	r.addRoute("OPTIONS", path, handler, nil)
}

// Group creates a new route group
func (r *Router) Group(prefix string) *RouteGroup {
	group := &RouteGroup{
		prefix:      prefix,
		middlewares: []middleware.MiddlewareFunc{},
		routes:      []*Route{},
		parent:      r,
	}
	r.groups[prefix] = group
	return group
}

// addRoute adds a route to the router
func (r *Router) addRoute(method, path string, handler http.HandlerFunc, middlewares []middleware.MiddlewareFunc) {
	route := &Route{
		Method:      method,
		Path:        path,
		Handler:     handler,
		Middlewares: middlewares,
	}

	key := method + ":" + path
	r.routes[key] = route

	logger.Log.Debug("Route added",
		zap.String("method", method),
		zap.String("path", path))
}

// findRoute finds a matching route for the given method and path
func (r *Router) findRoute(method, path string) *Route {
	// Check direct routes first
	key := method + ":" + path
	if route, exists := r.routes[key]; exists {
		return route
	}

	// Check parameterized routes
	for _, route := range r.routes {
		if route.Method == method && r.matchPath(route.Path, path) {
			return route
		}
	}

	// Check group routes
	for _, group := range r.groups {
		if route := group.findRoute(method, path); route != nil {
			return route
		}
	}

	return nil
}

// matchPath checks if a path pattern matches the given path
func (r *Router) matchPath(pattern, path string) bool {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return false
	}

	for i, patternPart := range patternParts {
		pathPart := pathParts[i]

		// Check for parameter (starts with :)
		if strings.HasPrefix(patternPart, ":") {
			continue
		}

		// Check for exact match
		if patternPart != pathPart {
			return false
		}
	}

	return true
}

// executeRoute executes a route with its middlewares
func (r *Router) executeRoute(w http.ResponseWriter, req *http.Request, route *Route) {
	// Create context with route parameters
	ctx := r.extractParams(req, route.Path)
	req = req.WithContext(ctx)

	// Execute middlewares
	handler := route.Handler
	for i := len(route.Middlewares) - 1; i >= 0; i-- {
		handler = route.Middlewares[i](handler)
	}

	// Execute the handler
	handler(w, req)
}

// extractParams extracts path parameters and adds them to the context
func (r *Router) extractParams(req *http.Request, pattern string) context.Context {
	ctx := req.Context()

	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(req.URL.Path, "/")

	for i, patternPart := range patternParts {
		if strings.HasPrefix(patternPart, ":") {
			paramName := strings.TrimPrefix(patternPart, ":")
			paramValue := pathParts[i]
			ctx = context.WithValue(ctx, routeParamKey(paramName), paramValue)
		}
	}

	return ctx
}

// GetParam gets a path parameter from the request context
func (r *Router) GetParam(req *http.Request, name string) string {
	if value := req.Context().Value(name); value != nil {
		return value.(string)
	}
	return ""
}

// RouteGroup methods

// GET adds a GET route to the group
func (g *RouteGroup) GET(path string, handler http.HandlerFunc) {
	g.addRoute("GET", path, handler)
}

// POST adds a POST route to the group
func (g *RouteGroup) POST(path string, handler http.HandlerFunc) {
	g.addRoute("POST", path, handler)
}

// PUT adds a PUT route to the group
func (g *RouteGroup) PUT(path string, handler http.HandlerFunc) {
	g.addRoute("PUT", path, handler)
}

// DELETE adds a DELETE route to the group
func (g *RouteGroup) DELETE(path string, handler http.HandlerFunc) {
	g.addRoute("DELETE", path, handler)
}

// OPTIONS adds an OPTIONS route to the group
func (g *RouteGroup) OPTIONS(path string, handler http.HandlerFunc) {
	g.addRoute("OPTIONS", path, handler)
}

// Use adds middleware to the group
func (g *RouteGroup) Use(middleware middleware.MiddlewareFunc) {
	g.middlewares = append(g.middlewares, middleware)
}

// addRoute adds a route to the group
func (g *RouteGroup) addRoute(method, path string, handler http.HandlerFunc) {
	fullPath := g.prefix + path
	route := &Route{
		Method:      method,
		Path:        fullPath,
		Handler:     handler,
		Middlewares: g.middlewares,
	}

	g.routes = append(g.routes, route)

	// Also add to parent router
	key := method + ":" + fullPath
	g.parent.routes[key] = route

	logger.Log.Debug("Group route added",
		zap.String("method", method),
		zap.String("path", fullPath),
		zap.String("group", g.prefix))
}

// findRoute finds a matching route in the group
func (g *RouteGroup) findRoute(method, path string) *Route {
	for _, route := range g.routes {
		if route.Method == method && g.parent.matchPath(route.Path, path) {
			return route
		}
	}
	return nil
}

// Group creates a new route group within this group
func (g *RouteGroup) Group(prefix string) *RouteGroup {
	group := &RouteGroup{
		prefix:      g.prefix + prefix,
		middlewares: g.middlewares,
		routes:      []*Route{},
		parent:      g.parent,
	}
	g.parent.groups[g.prefix+prefix] = group
	return group
}

// RouteMatcher provides advanced route matching capabilities
type RouteMatcher struct {
	routes []*Route
}

// NewRouteMatcher creates a new route matcher
func NewRouteMatcher() *RouteMatcher {
	return &RouteMatcher{
		routes: []*Route{},
	}
}

// AddRoute adds a route to the matcher
func (rm *RouteMatcher) AddRoute(route *Route) {
	rm.routes = append(rm.routes, route)
}

// Match finds the best matching route
func (rm *RouteMatcher) Match(method, path string) *Route {
	var bestMatch *Route
	bestScore := -1

	for _, route := range rm.routes {
		if route.Method != method {
			continue
		}

		score := rm.calculateMatchScore(route.Path, path)
		if score > bestScore {
			bestScore = score
			bestMatch = route
		}
	}

	return bestMatch
}

// calculateMatchScore calculates a match score for a route pattern
func (rm *RouteMatcher) calculateMatchScore(pattern, path string) int {
	patternParts := strings.Split(pattern, "/")
	pathParts := strings.Split(path, "/")

	if len(patternParts) != len(pathParts) {
		return -1
	}

	score := 0
	for i, patternPart := range patternParts {
		pathPart := pathParts[i]

		if strings.HasPrefix(patternPart, ":") {
			// Parameter match
			score += 1
		} else if patternPart == pathPart {
			// Exact match
			score += 2
		} else {
			// No match
			return -1
		}
	}

	return score
}
