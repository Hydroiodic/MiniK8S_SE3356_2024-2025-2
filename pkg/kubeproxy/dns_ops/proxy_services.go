package dns_ops

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
)

// Rule defines a forwarding rule for the dynamic proxy.
type Rule struct {
	PathPrefix string
	Target     string // e.g. "http://10.0.0.2:80" or "https://10.0.0.1:443"
}

type DynamicProxy struct {
	rules     []Rule
	mu        sync.RWMutex
	proxyPool map[string]*httputil.ReverseProxy
}

// NewDynamicProxy creates a new instance of DynamicProxy.
func NewDynamicProxy() *DynamicProxy {
	return &DynamicProxy{
		proxyPool: make(map[string]*httputil.ReverseProxy),
	}
}

// UpdateRules updates the forwarding rules for the proxy.
func (dp *DynamicProxy) UpdateRules(newRules []Rule) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	// Update the rules.
	dp.rules = newRules

	// Empty current proxy pool.
	dp.proxyPool = make(map[string]*httputil.ReverseProxy)
}

// getProxyForTarget gets or creates a ReverseProxy for the given target.
func (dp *DynamicProxy) getProxyForTarget(
	target string,
) *httputil.ReverseProxy {
	dp.mu.RLock()
	if proxy, exists := dp.proxyPool[target]; exists {
		dp.mu.RUnlock()
		return proxy
	}
	dp.mu.RUnlock()

	// Create a new ReverseProxy for the target.
	targetURL, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Configure the proxy to skip TLS verification.
	proxy.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		Dial: func(network, addr string) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}

	dp.mu.Lock()
	dp.proxyPool[target] = proxy
	dp.mu.Unlock()

	return proxy
}

// ServeHTTP Implements the http.Handler interface for DynamicProxy.
func (dp *DynamicProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	// Find a matching rule based on the request path.
	for _, rule := range dp.rules {
		if strings.HasPrefix(r.URL.Path, rule.PathPrefix) {
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
	}

	// If no matching rule is found, return a 404 error.
	http.Error(w, "No matching rule found", http.StatusNotFound)
}

// handleCONNECT handles HTTP CONNECT requests for tunneling.
func (dp *DynamicProxy) HandleCONNECT(w http.ResponseWriter, r *http.Request) {
	dp.mu.RLock()
	defer dp.mu.RUnlock()

	// Find a matching rule based on the request path.
	for _, rule := range dp.rules {
		if strings.HasPrefix(r.URL.Path, rule.PathPrefix) {
			targetURL, _ := url.Parse(rule.Target)

			// Create a TCP connection to the target server.
			destConn, err := net.Dial("tcp", targetURL.Host)
			if err != nil {
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
				http.Error(w, err.Error(), http.StatusServiceUnavailable)
				return
			}

			// Start transferring data between the client and the destination.
			go transfer(destConn, clientConn)
			go transfer(clientConn, destConn)

			return
		}
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

func main() {
	// Create a new dynamic proxy instance.
	proxy := NewDynamicProxy()

	// Initialize with some rules.
	initialRules := []Rule{
		{PathPrefix: "/1", Target: "https://10.0.0.1:443"},
		{PathPrefix: "/2", Target: "http://10.0.0.2:80"},
	}
	proxy.UpdateRules(initialRules)

	// Set up the HTTP server with the dynamic proxy.
	server := &http.Server{
		Addr: ":8080",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				proxy.HandleCONNECT(w, r)
			} else {
				proxy.ServeHTTP(w, r)
			}
		}),
	}

	// Start the HTTP server.
	log.Println("Starting proxy server on :8080")
	log.Fatal(server.ListenAndServe())
}
