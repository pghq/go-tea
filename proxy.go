package tea

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/pghq/go-tea/health"
	"github.com/pghq/go-tea/trail"
)

// Proxy is a multi-host reverse proxy
type Proxy struct {
	directors   map[string]*httputil.ReverseProxy
	middlewares []Middleware
	cors        Middleware
	trace       MiddlewareFunc
	health      *health.Service
}

// Middleware adds a middleware to the proxy
func (p *Proxy) Middleware(middlewares ...Middleware) {
	p.middlewares = append(p.middlewares, middlewares...)
}

// Direct sets a new director for the path
// e.g., pathPrefix is typically the name of the microservice
func (p *Proxy) Direct(pathPrefix, host string) error {
	hostURL, err := url.ParseRequestURI(host)
	if err != nil {
		return trail.Stacktrace(err)
	}

	p.directors[pathPrefix] = &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.Header.Add("X-Forwarded-Host", r.Host)
			r.Header.Add("X-Forwarded-Proto", r.URL.Scheme)
			r.URL.Host = hostURL.Host
			r.URL.Scheme = hostURL.Scheme
		},
	}

	healthURL := *hostURL
	healthURL.Path = path.Join(healthURL.Path, "/health/status")
	p.health.AddDependency(host, healthURL.String())
	return nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var handler http.Handler
	urlPath := strings.TrimPrefix(r.URL.Path, string(os.PathSeparator))
	middlewares := []Middleware{p.cors}
	if r.URL.Path == "/health/status" {
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Send(w, r, p.health.Status())
		})
	} else {
		var sb strings.Builder
		for _, dir := range strings.Split(urlPath, string(os.PathSeparator)) {
			sb.WriteString(dir)
			if director, present := p.directors[sb.String()]; present {
				handler = director
				middlewares = append(middlewares, p.trace)
				middlewares = append(middlewares, p.middlewares...)
				break
			}
			sb.WriteString("/")
		}
	}

	if handler == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	for i := len(middlewares) - 1; i >= 0; i-- {
		m := middlewares[i]
		handler = m.Handle(handler)
	}

	handler.ServeHTTP(w, r)
}

// NewProxy creates a new multi-host reverse proxy
func NewProxy(version string) *Proxy {
	p := Proxy{
		directors: make(map[string]*httputil.ReverseProxy),
		cors:      NewCORSMiddleware(),
		trace:     trail.NewTraceMiddleware(version, false),
		health:    health.NewService(version),
	}

	return &p
}
