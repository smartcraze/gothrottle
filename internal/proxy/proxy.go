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

/*
Handler manages reverse proxy functionality with path-based routing.
It matches incoming request paths to configured upstream targets using
longest prefix matching and forwards requests accordingly.
*/
type Handler struct {
	proxies map[string]*httputil.ReverseProxy
	routes  []config.Route
}

func NewHandler(routes []config.Route) (*Handler, error) {
	handler := &Handler{
		proxies: make(map[string]*httputil.ReverseProxy),
		routes:  routes,
	}

	for _, route := range routes {
		targetURL, err := url.Parse(route.Target)
		if err != nil {
			return nil, fmt.Errorf("invalid target URL for path %s: %w", route.Path, err)
		}

		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		originalDirector := proxy.Director
		proxy.Director = func(req *http.Request) {
			originalDirector(req)
		}

		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			http.Error(w, fmt.Sprintf("Bad Gateway: %v", err), http.StatusBadGateway)
		}

		handler.proxies[route.Path] = proxy
	}

	return handler, nil
}

/*
Handle processes incoming requests using longest prefix matching to find
the appropriate upstream target, then forwards the request via reverse proxy.
*/
func (h *Handler) Handle(c *gin.Context) {
	requestPath := c.Request.URL.Path
	var matchedRoute *config.Route
	var matchedProxy *httputil.ReverseProxy

	for i := range h.routes {
		route := &h.routes[i]
		if strings.HasPrefix(requestPath, route.Path) {
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

	matchedProxy.ServeHTTP(c.Writer, c.Request)
}

func (h *Handler) GetRoutes() []config.Route {
	return h.routes
}
