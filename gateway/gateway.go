package gateway

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

const unknown string = "unknown"


type Gateway struct {
	services map[string]*Service // maps service name to service
	config GatewayConfig
}

type Service struct {
	Name string
	URL  *url.URL
	Proxy *httputil.ReverseProxy
	// TODO: add rate limiter bucket
	// TODO: add health checkpoint -- requires some backend changes in purch
}

// Creates a new Gateway with services already configured predefined in the config.yml
func NewGateway() *Gateway {
	gateway := &Gateway{
		services: make(map[string]*Service),
		config: *ReadConfig(),
	}

	if err := gateway.configureServices(); err != nil {
		slog.Error("error configuring services defined in config for gateway.", "error", err.Error())
		panic(err)
	}

	return gateway
}

// Configures the gateway with all services defined in the config.yml
func (g *Gateway) configureServices() error {
	for _, serviceConfig := range g.config.Services {
		g.addService(serviceConfig.Name, serviceConfig.Address)
	}
	return nil
}

// Registers a new backend service
func (g *Gateway) addService(name, targetURL string) error {
	target, err := url.Parse(targetURL)
	if err != nil {
		slog.Error("invalid target URL.", "error", err.Error())
		return fmt.Errorf("invalid target URL: %v", err)
	}

	// Create reverse proxy for this service
	proxy := httputil.NewSingleHostReverseProxy(target)
	
	// Customize the proxy behavior
	// proxy.ModifyResponse = g.modifyResponse
	proxy.ErrorHandler = g.errorHandler
	
	// Custom Director to modify requests before forwarding
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req) // Apply default director first
		
		// Add custom headers
		req.Header.Set("X-Gateway", "gateway-purch")
		req.Header.Set("X-Forwarded-Service", name)
		
		// Add user context from authentication middleware
		// if user, ok := req.Context().Value("user").(map[string]interface{}); ok {
		// 	req.Header.Set("X-User-ID", fmt.Sprintf("%v", user["id"]))
		// 	req.Header.Set("X-User-Email", fmt.Sprintf("%v", user["email"]))
		// }
		
		// Remove sensitive headers that shouldn't go to backend
		req.Header.Del("Authorization")
		
		// Log the request
		slog.Debug("forwarding", "service", name, "method", req.Method, "request-url", req.URL.Path, "target-url", target)
	}

	g.services[name] = &Service{
		Name:   name,
		URL:    target,
		Proxy:  proxy,
	}
	
	slog.Info("Registered new service", "service", name, "target-url", targetURL)
	return nil
}

// errorHandler handles proxy errors
func (g *Gateway) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	serviceName := r.Context().Value("service")
	slog.Error("Proxy error", "service", serviceName, "error", err.Error())
	
	// Return appropriate error to client
	// TODO: refactor to make errors more useful if needed
	if strings.Contains(err.Error(), "connection refused") {
		http.Error(w, "service temporarily unavailable", http.StatusServiceUnavailable)
	} else {
		http.Error(w, "internal gateway error", http.StatusInternalServerError)
	}
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serviceName := g.determineService(r.URL.Path)
	service, ok := g.services[serviceName]
	if !ok {
		slog.Error("service not configured", "service", serviceName)
		http.Error(w, fmt.Sprintf("%s service not configured", serviceName), http.StatusNotFound)
		return 
	}

	// service info for logging
	ctx := context.WithValue(r.Context(), "service", serviceName)
	r = r.WithContext(ctx)

	// forward request to the service
	service.Proxy.ServeHTTP(w, r)
}

// retrieves serviceName for request forwarding
func (g *Gateway) determineService(urlPath string) string {
	splitPath := strings.Split(strings.TrimPrefix(urlPath, "/"), "/")
	if len(splitPath) >= 2 && splitPath[0] == "api" {
		return splitPath[1]
	}

	return unknown
}