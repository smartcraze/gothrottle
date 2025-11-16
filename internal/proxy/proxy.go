package proxy

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/smartcraze/gothrottle/internal/config"
)

// Handler manages reverse proxy functionality with path-based routing
type Handler struct {
	proxies map[string]*httputil.ReverseProxy // map[path]proxy
	routes  []config.Route                     // ordered list of routes for matching
}

// NewHandler creates a new proxy handler from configuration
func NewHandler(routes []config.Route) (*Handler, error) {
	
	handler := &Handler{
		proxies: make(map[string]*httputil.ReverseProxy),
		routes:  routes,
	}

	// Create a reverse proxy for each route
	for _, route := range routes {
		targetURL, err := url.Parse(route.Target)
		if err != nil {
			return nil, fmt.Errorf("invalid target URL for path %s: %w", route.Path, err)
		}

		// Create reverse proxy with custom director
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		
		// Customize the director to preserve the original path
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
			// Keep the original path (don't strip the prefix)
			// If you want to strip the prefix, uncomment below:
			// req.URL.Path = strings.TrimPrefix(req.URL.Path, route.Path)
		}

		// Custom error handler
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, fmt.Sprintf("Bad Gateway: %v", err), http.StatusBadGateway)
		}

		handler.proxies[route.Path] = proxy
	}

	return handler, nil
}

// Handle processes incoming requests and forwards them to the appropriate upstream
func (h *Handler) Handle(c *gin.Context) {
	requestPath := c.Request.URL.Path

	// Find matching route (longest prefix match)
	var matchedRoute *config.Route
	var matchedProxy *httputil.ReverseProxy

	for i := range h.routes {
		route := &h.routes[i]
		if strings.HasPrefix(requestPath, route.Path) {
			// Use longest prefix match
			if matchedRoute == nil || len(route.Path) > len(matchedRoute.Path) {
				matchedRoute = route
				matchedProxy = h.proxies[route.Path]
			}
		}
	}

	if matchedProxy == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "no route found for path",
			"path":  requestPath,
		})
		return
	}

	// Forward the request to the upstream
	matchedProxy.ServeHTTP(c.Writer, c.Request)
}

// GetRoutes returns the configured routes
func (h *Handler) GetRoutes() []config.Route {
	return h.routes
}
