// Package microauth provides a uniform means of serving HTTP/S for golang
// projects securely. It allows the specification of a certificate (or
// generates one) as well as an auth token which is checked before the request
// is processed.
package microauth

import (
	"crypto/subtle"
	"crypto/tls"
	"errors"
	"net/http"
)

// Auth is a structure containing listener information
type Auth struct {
	child         http.Handler     // child is the http handler passed in
	Header        string           // Header is the authentication token's header name
	Certificate   *tls.Certificate // Certificate is the tls.Certificate to serve requests with
	ExcludedPaths []string         // ExcludedPaths is a list of paths to be excluded from being authenticated
	Token         string           // Token is the security/authentication string to validate by
}

var (
	// DefaultAuth is the default Auth object
	DefaultAuth = &Auth{}
)

func init() {
	DefaultAuth.Header = "X-MICROBOX-TOKEN"
	DefaultAuth.Certificate, _ = Generate("microbox.cloud")
}

// ServeHTTP is to implement the http.Handler interface. Also let clients know
// when I have no matching route listeners
func (me *Auth) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	reqPath := req.URL.Path
	skipOnce := false

	for _, path := range me.ExcludedPaths {
		if path == reqPath {
			skipOnce = true
			break
		}
	}

	// open up for the CORS "secure" pre-flight check (browser doesn't allow devs to set headers in OPTIONS request)
	if req.Method == "OPTIONS" {
		// todo: actually check origin header to better implement CORS
		skipOnce = true
	}

	if !skipOnce {
		auth := ""
		if auth = req.Header.Get(me.Header); auth == "" {
			// check form value (case sensitive) if header not set
			auth = req.FormValue(me.Header)
		}

		if subtle.ConstantTimeCompare([]byte(auth), []byte(me.Token)) == 0 {
			rw.WriteHeader(http.StatusUnauthorized)
			return
		}
	}

	me.child.ServeHTTP(rw, req)
}

// ListenAndServeTLS starts a TLS listener and handles serving https
func (me *Auth) ListenAndServeTLS(addr, token string, h http.Handler, excludedPaths ...string) error {
	if token == "" {
		return errors.New("microauth: token missing")
	}
	config := &tls.Config{
		Certificates: []tls.Certificate{*me.Certificate},
	}

	me.ExcludedPaths = excludedPaths
	me.Token = token

	if h == nil {
		h = http.DefaultServeMux
	}
	me.child = h

	server := &http.Server{Addr: addr, Handler: me, TLSConfig: config}
	return server.ListenAndServeTLS("", "")
}

// ListenAndServe starts a normal tcp listener and handles serving http while
// still validating the auth token.
func (me *Auth) ListenAndServe(addr, token string, h http.Handler, excludedPaths ...string) error {
	if token == "" {
		return errors.New("microauth: token missing")
	}

	me.ExcludedPaths = excludedPaths
	me.Token = token

	if h == nil {
		h = http.DefaultServeMux
	}
	me.child = h

	return http.ListenAndServe(addr, me)
}

// ListenAndServeTLS is a shortcut function which uses the default one
func ListenAndServeTLS(addr, token string, h http.Handler, excludedPaths ...string) error {
	return DefaultAuth.ListenAndServeTLS(addr, token, h, excludedPaths...)
}

// ListenAndServe is a shortcut function which uses the default one
func ListenAndServe(addr, token string, h http.Handler, excludedPaths ...string) error {
	return DefaultAuth.ListenAndServe(addr, token, h, excludedPaths...)
}
