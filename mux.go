package httpx

import (
	"net/http"

	"github.com/go-chi/chi"
)

// Mux is a simple HTTP route multiplexer that parses a request path,
// records any URL params, and executes an end handler. It implements
// the http.Handler interface and is friendly with the standard library.
//
// Mux is designed to be fast, minimal and offer a powerful API for building
// modular and composable HTTP services with a large set of handlers. It's
// particularly useful for writing large REST API services that break a handler
// into many smaller parts composed of middlewares and end handlers.
type Mux struct {
	chi         *chi.Mux
	middlewares []Middleware
	prefix      string
}

// NewMux returns a newly initialized Mux object
func NewMux() *Mux {
	return &Mux{
		chi:         chi.NewMux(),
		middlewares: []Middleware{},
	}
}

// Use appends a middleware handler to the Mux middleware stack.
func (m *Mux) Use(middlewares ...Middleware) {
	m.middlewares = append(m.middlewares, middlewares...)
}

// With adds inline middlewares for an endpoint handler.
func (m *Mux) With(middlewares ...Middleware) *Mux {
	var mws []Middleware
	mws = make([]Middleware, len(m.middlewares))
	copy(mws, m.middlewares)

	mws = append(mws, middlewares...)

	return &Mux{
		chi:         m.chi,
		middlewares: mws,
		prefix:      m.prefix,
	}
}

// Group creates a new inline-Mux with a fresh middleware stack. It's useful
// for a group of handlers along the same routing path that use an additional
// set of middlewares.
func (m *Mux) Group(fn func(*Mux)) *Mux {
	im := m.With()
	if fn != nil {
		fn(im)
	}
	return im
}

// Route creates a new Mux with a fresh middleware stack and mounts it
// along the `pattern` as a subrouter.
func (m *Mux) Route(pattern string, fn func(*Mux)) *Mux {
	if pattern[len(pattern)-1] != '/' {
		pattern += "/"
	}
	im := m.With()
	im.prefix += pattern
	if fn != nil {
		fn(im)
	}
	return im
}

// Handle adds the route `pattern` that matches any http method to
// execute the `handler` httpx.Handler.
func (m *Mux) Handle(pattern string, handler Handler) {
	m.chi.Handle(m.prefix+pattern, adaptor(NewChain(m.middlewares...).Then(handler)))
}

// HandleFunc adds the route `pattern` that matches any http method to
// execute the `handlerFn` httpx.HandlerFunc.
func (m *Mux) HandleFunc(pattern string, handlerFn HandlerFunc) {
	m.Handle(pattern, handlerFn)
}

// Method adds the route `pattern` that matches `method` http method to
// execute the `handler` httpx.Handler.
func (m *Mux) Method(method, pattern string, h Handler) {
	m.chi.Method(method, m.prefix+pattern, adaptor(NewChain(m.middlewares...).Then(h)))
}

// MethodFunc adds the route `pattern` that matches `method` http method to
// execute the `handlerFn` httpx.HandlerFunc.
func (m *Mux) MethodFunc(method, pattern string, handlerFn HandlerFunc) {
	m.Method(method, pattern, handlerFn)
}

// Connect adds the route `pattern` that matches a CONNECT http method to
// execute the `handlerFn` httpx.HandlerFunc.
func (m *Mux) Connect(pattern string, handlerFn HandlerFunc) {
	m.Method(http.MethodConnect, pattern, handlerFn)
}

// Delete adds the route `pattern` that matches a DELETE http method to
// execute the `handlerFn` httpx.HandlerFunc.
func (m *Mux) Delete(pattern string, handlerFn HandlerFunc) {
	m.Method(http.MethodDelete, pattern, handlerFn)
}

// Get adds the route `pattern` that matches a GET http method to
// execute the `handlerFn` httpx.HandlerFunc.
func (m *Mux) Get(pattern string, handlerFn HandlerFunc) {
	m.Method(http.MethodGet, pattern, handlerFn)
}

// Head adds the route `pattern` that matches a HEAD http method to
// execute the `handlerFn` httpx.HandlerFunc.
func (m *Mux) Head(pattern string, handlerFn HandlerFunc) {
	m.Method(http.MethodHead, pattern, handlerFn)
}

// Options adds the route `pattern` that matches a OPTIONS http method to
// execute the `handlerFn` httpx.HandlerFunc.
func (m *Mux) Options(pattern string, handlerFn HandlerFunc) {
	m.Method(http.MethodOptions, pattern, handlerFn)
}

// Patch adds the route `pattern` that matches a PATCH http method to
// execute the `handlerFn` httpx.HandlerFunc.
func (m *Mux) Patch(pattern string, handlerFn HandlerFunc) {
	m.Method(http.MethodPatch, pattern, handlerFn)
}

// Post adds the route `pattern` that matches a POST http method to
// execute the `handlerFn` httpx.HandlerFunc.
func (m *Mux) Post(pattern string, handlerFn HandlerFunc) {
	m.Method(http.MethodPost, pattern, handlerFn)
}

// Put adds the route `pattern` that matches a PUT http method to
// execute the `handlerFn` httpx.HandlerFunc.
func (m *Mux) Put(pattern string, handlerFn HandlerFunc) {
	m.Method(http.MethodPut, pattern, handlerFn)
}

// Trace adds the route `pattern` that matches a TRACE http method to
// execute the `handlerFn` httpx.HandlerFunc.
func (m *Mux) Trace(pattern string, handlerFn HandlerFunc) {
	m.Method(http.MethodTrace, pattern, handlerFn)
}

// NotFound sets a custom http.HandlerFunc for routing paths that could
// not be found. The default 404 handler is `http.NotFound`.
func (m *Mux) NotFound(handlerFn HandlerFunc) {
	m.chi.NotFound(adaptor(handlerFn))
}

// MethodNotAllowed sets a custom http.HandlerFunc for routing paths where the
// method is unresolved. The default handler returns a 405 with an empty body.
func (m *Mux) MethodNotAllowed(handlerFn HandlerFunc) {
	m.chi.NotFound(adaptor(handlerFn))
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.chi.ServeHTTP(w, r)
}

func adaptor(next Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := next.ServeHTTP(w, r); err != nil {
			if sErr, ok := err.(StatusError); ok {
				http.Error(w, err.Error(), sErr.Status())
			}
			http.Error(w, err.Error(), 500)
		}
	})
}
