package dns_ops

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/Hydroiodic/MiniK8S_SE3356_2024-2025-2/pkg/object"
)

type DynamicProxy struct {
	rules     []object.ProxyRule
	mu        sync.RWMutex
	proxyPool sync.Map
}

// `NewDynamicProxy` creates a new instance of DynamicProxy.
func NewDynamicProxy() *DynamicProxy {
	return &DynamicProxy{
		rules:     make([]object.ProxyRule, 0),
		mu:        sync.RWMutex{},
		proxyPool: sync.Map{},
	}
}

// UpdateRules updates the forwarding rules for the proxy.
func (dp *DynamicProxy) UpdateRules(newRules []object.ProxyRule) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	// Update the rules with the new rules.
	dp.rules = newRules

	// Log the updated rules.
	log.Printf("Proxy rules updated: %d rules loaded\n", len(dp.rules))

	for _, rule := range dp.rules {
		log.Printf(
			"Rule: %s%s -> %s\n",
			rule.Domain,
			rule.PathPrefix,
			rule.Target,
		)
	}
}

// getProxyForTarget gets or creates a ReverseProxy for the given target.
func (dp *DynamicProxy) getProxyForTarget(
	target string,
) *httputil.ReverseProxy {
	// Try to get the proxy from the cache at first.
	if proxy, exists := dp.proxyPool.Load(target); exists {
		return proxy.(*httputil.ReverseProxy)
	}

	// Create a new ReverseProxy for the target.
	targetURL, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Configure the proxy to handle errors and timeouts.
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}
			return dialer.DialContext(ctx, network, addr)
		},
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}

	// Configure the proxy to handle errors.
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		// Log the error and return a 502 Bad Gateway response.
		log.Printf("Proxy error: %v", err)
		w.WriteHeader(http.StatusBadGateway)
		// Write a simple error message to the response.
		_, err = w.Write([]byte("Bad Gateway"))
		if err != nil {
			log.Printf("Error writing response: %v", err)
		}
	}

	// Store the proxy in the cache.
	dp.proxyPool.Store(target, proxy)

	return proxy
}

// `domainMatches` checks if the request host matches the rule domain pattern
func (dp *DynamicProxy) domainMatches(ruleDomain, requestHost string) bool {
	// Normalize both domains to lowercase and remove trailing dots
	normalize := func(domain string) string {
		domain = strings.ToLower(domain)
		return strings.TrimSuffix(domain, ".")
	}

	return normalize(ruleDomain) == normalize(requestHost)
}

// ServeHTTP implements the http.Handler interface for DynamicProxy.
func (dp *DynamicProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Configure a context with a timeout for the request.
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	// Ensure the request context is set to the new context with timeout.
	r = r.WithContext(ctx)

	// Normalize the request host to lowercase and remove port if present.
	host := strings.ToLower(r.Host)
	if strings.Contains(host, ":") {
		var err error
		// Split the host and port if present.
		host, _, err = net.SplitHostPort(host)
		if err != nil {
			http.Error(w, "Invalid host format", http.StatusBadRequest)
			return
		}
	}

	dp.mu.RLock()
	defer dp.mu.RUnlock()

	// Find the first matching rule based on the request path.
	for _, rule := range dp.rules {
		// Check if the request host matches the rule's domain.
		if !dp.domainMatches(rule.Domain, host) {
			continue
		}

		// Check if the request path starts with the rule's path prefix.
		if !strings.HasPrefix(r.URL.Path, rule.PathPrefix) {
			continue
		}

		// Get or create a ReverseProxy for the target.
		proxy := dp.getProxyForTarget(rule.Target)

		// Modify the request URL to point to the target.
		targetURL, _ := url.Parse(rule.Target)
		r.URL.Scheme = targetURL.Scheme
		r.URL.Host = targetURL.Host
		r.Host = targetURL.Host

		// Set the request path to the target path.
		proxy.ServeHTTP(w, r)

		return
	}

	http.Error(w, "No matching rule found", http.StatusNotFound)
}

// `HandleCONNECT` handles HTTP CONNECT requests for tunneling.
func (dp *DynamicProxy) HandleCONNECT(w http.ResponseWriter, r *http.Request) {
	// Configure a context with a timeout for the CONNECT request.
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	// Normalize the request host to lowercase and remove port if present.
	host := strings.ToLower(r.Host)
	if strings.Contains(host, ":") {
		var err error
		// Split the host and port if present.
		host, _, err = net.SplitHostPort(host)
		if err != nil {
			http.Error(w, "Invalid host format", http.StatusBadRequest)
			return
		}
	}

	dp.mu.RLock()
	defer dp.mu.RUnlock()

	// Find the first matching rule based on the request path.
	for _, rule := range dp.rules {
		// Check if the request host matches the rule's domain.
		if !dp.domainMatches(rule.Domain, host) {
			continue
		}

		// Check if the request path starts with the rule's path prefix.
		if !strings.HasPrefix(r.URL.Path, rule.PathPrefix) {
			continue
		}

		// Parse the target URL from the rule.
		targetURL, _ := url.Parse(rule.Target)

		// Use a dialer to connect to the target URL.
		dialer := &net.Dialer{
			Timeout: 5 * time.Second,
		}

		// Try to establish a connection to the target URL.
		destConn, err := dialer.DialContext(ctx, "tcp", targetURL.Host)
		if err != nil {
			log.Printf("CONNECT dial error: %v", err)
			http.Error(w, err.Error(), http.StatusServiceUnavailable)

			return
		}

		// Send a 200 OK response to the client.
		w.WriteHeader(http.StatusOK)

		// Upgrade the connection to a hijacked connection.
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			http.Error(
				w,
				"Hijacking not supported",
				http.StatusInternalServerError,
			)

			return
		}

		clientConn, _, err := hijacker.Hijack()
		if err != nil {
			// Log the error and send a 503 Service Unavailable response.
			log.Printf("Hijack error: %v", err)
			http.Error(w, err.Error(), http.StatusServiceUnavailable)

			// Close the destination connection.
			if err := destConn.Close(); err != nil {
				log.Printf("Error closing destination connection: %v", err)
			}

			return
		}

		// Launch goroutines to transfer data between the client and the destination.
		go transfer(destConn, clientConn)
		go transfer(clientConn, destConn)

		return
	}

	http.Error(w, "No matching rule found", http.StatusNotFound)
}

func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer func() {
		// Close the destination connection.
		if err := destination.Close(); err != nil {
			log.Printf("Error closing destination: %v", err)
		}
		// Close the source connection.
		if err := source.Close(); err != nil {
			log.Printf("Error closing source: %v", err)
		}
	}()

	// Transfer data from source to destination.
	if _, err := io.Copy(destination, source); err != nil {
		log.Printf("Error during data transfer: %v", err)
	}
}
