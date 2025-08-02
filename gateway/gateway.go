package gateway

import (
	"fmt"
	"net/url"
	"net/http"
	"net/http/httputil"
	"log/slog"
	"strings"
	"context"
)

const unknown string = "unknown"


type Service struct {
	Name string
	URL  *url.URL
	Proxy *httputil.ReverseProxy
	// TODO: add rate limiter bucket
	// TODO: add health checkpoint -- backend changes in purch
}

type Gateway struct {
	services map[string]*Service // maps service name to service
	config GatewayConfig
}

// Creates a new Gateway with services already configured predefined in the config.yml
func NewGateway() *Gateway {
	gateway := &Gateway{
		services: make(map[string]*Service),
		config: *ReadConfig(),
	}

	if err := gateway.configureServices(); err != nil {
		slog.Error("error configuring predefined services for gateway.", "error", err.Error())
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
		req.Header.Del("X-API-Key")
		
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

// TODO: return actually usable errors soon
// errorHandler handles proxy errors
func (g *Gateway) errorHandler(w http.ResponseWriter, r *http.Request, err error) {
	serviceName := r.Context().Value("service")
	slog.Error("[%s] Proxy error: %v", serviceName, err)
	
	// Return appropriate error to client
	if strings.Contains(err.Error(), "connection refused") {
		http.Error(w, "Service temporarily unavailable", http.StatusServiceUnavailable)
	} else {
		http.Error(w, "Internal gateway error", http.StatusInternalServerError)
	}
}

// retrieves serviceName for proxy forwarding
func (g *Gateway) determineService(urlPath string) string {
	splitPath := strings.Split(strings.TrimPrefix(urlPath, "/"), "/")
	potentialService := splitPath[0]

	// verify it's a spun up service
	_, ok := g.services[potentialService]
	if !ok {
		return unknown
	}

	return potentialService
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serviceName := g.determineService(r.URL.Path)
	service, ok := g.services[serviceName]
	if !ok {
		slog.Error(fmt.Sprintf("%s service not configured", serviceName))
		http.Error(w, fmt.Sprintf("%s service not configured", serviceName), http.StatusNotFound)
		return 
	}

	// service info for logging
	ctx := context.WithValue(r.Context(), "service", serviceName)
	r = r.WithContext(ctx)

	// forward request ot the service
	service.Proxy.ServeHTTP(w, r)
}