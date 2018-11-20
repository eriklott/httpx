package httpx

import "net/http"

// A Handler responds to an HTTP request.
type Handler interface {
	ServeHTTP(http.ResponseWriter, *http.Request) error
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as HTTP handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (fn HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return fn(w, r)
}

// RedirectHandler returns a request handler that redirects
// each request it receives to the given url using the given
// status code.
//
// The provided code should be in the 3xx range and is usually
// StatusMovedPermanently, StatusFound or StatusSeeOther.
func RedirectHandler(url string, code int) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		http.Redirect(w, r, url, code)
		return nil
	})
}

// FileServer returns a handler that serves HTTP requests
// with the contents of the file system rooted at root.
//
// To use the operating system's file system implementation,
// use http.Dir:
//
//     http.Handle("/", http.FileServer(http.Dir("/tmp")))
//
// As a special case, the returned file server redirects any request
// ending in "/index.html" to the same path, without the final
// "index.html".
func FileServer(root http.FileSystem) Handler {
	fs := http.FileServer(root)
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		fs.ServeHTTP(w, r)
		return nil
	})
}

// StripPrefix returns a handler that serves HTTP requests
// by removing the given prefix from the request URL's Path
// and invoking the handler h. StripPrefix handles a
// request for a path that doesn't begin with prefix by
// replying with an HTTP 404 not found error.
func StripPrefix(prefix string, h Handler) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		var err error
		http.StripPrefix(prefix, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			err = h.ServeHTTP(w, r)
		})).ServeHTTP(w, r)
		return err
	})
}

func ErrorHandler(code int, message string) Handler {
	return HandlerFunc(func(w http.ResponseWriter, r *http.Request) error {
		return Error(code, message)
	})
}
