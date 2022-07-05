package tea

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"

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
// e.g., root is typically the name of the microservice
func (p *Proxy) Direct(root, host string) error {
	root = strings.Trim(root, string(os.PathSeparator))
	hostURL, err := url.ParseRequestURI(host)
	if err != nil {
		return trail.Stacktrace(err)
	}
	p.directors[root] = &httputil.ReverseProxy{
		Director: func(r *http.Request) {
			r.Header.Add("X-Forwarded-Host", r.Host)
			r.Header.Add("X-Forwarded-Proto", r.URL.Scheme)
			r.URL.Host = hostURL.Host
			r.URL.Scheme = hostURL.Scheme
			r.URL.Path = strings.TrimPrefix(r.URL.Path, string(os.PathSeparator))
			r.URL.Path = strings.TrimPrefix(r.URL.Path, root)
		},
	}

	healthURL := *hostURL
	healthURL.Path = path.Join(healthURL.Path, "/health/status")
	p.health.AddDependency(root, healthURL.String())
	return nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, string(os.PathSeparator))
	root := strings.Split(filepath.Dir(path), string(os.PathSeparator))[0]
	director, present := p.directors[root]
	var handler http.Handler
	middlewares := []Middleware{p.cors}
	switch {
	case r.URL.Path == "/health/status":
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Send(w, r, p.health.Status())
		})
	case present:
		handler = director
		middlewares = append(middlewares, p.trace)
		middlewares = append(middlewares, p.middlewares...)
	default:
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
func NewProxy(semver string) *Proxy {
	v, _ := version.NewVersion(semver)
	cv := semver
	if v != nil {
		cv = v.String()
	}

	p := Proxy{
		directors: make(map[string]*httputil.ReverseProxy),
		cors:      NewCORSMiddleware(),
		trace:     trail.NewTraceMiddleware(cv, false),
		health:    health.NewService(cv),
	}

	return &p
}
