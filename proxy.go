package tea

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/pghq/go-tea/health"
)

// Proxy is a multi-host reverse proxy
type Proxy struct {
	directors   map[string]*httputil.ReverseProxy
	base        []Middleware
	middlewares []Middleware
	health      http.Handler
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
		return Stack(err)
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

	return nil
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, string(os.PathSeparator))
	root := strings.Split(filepath.Dir(path), string(os.PathSeparator))[0]
	director, present := p.directors[root]
	var handler http.Handler
	middlewares := p.base
	switch {
	case r.URL.Path == "/health/status":
		handler = p.health
	case present:
		handler = director
		middlewares = append(middlewares, p.middlewares...)
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}

	for _, m := range middlewares {
		handler = m.Handle(handler)
	}

	handler.ServeHTTP(w, r)
}

// NewProxy creates a new multi-host reverse proxy
func NewProxy(semver string) *Proxy {
	v, _ := version.NewVersion(semver)
	p := Proxy{
		directors: make(map[string]*httputil.ReverseProxy),
		base:      []Middleware{Trace(v), CORS()},
		health: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			Send(w, r, health.NewService(semver).Status())
		}),
	}
	return &p
}
