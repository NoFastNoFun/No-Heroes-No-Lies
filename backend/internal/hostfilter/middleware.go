package hostfilter

import (
	"net/http"
	"strings"
)

// Middleware blocks any request whose Host header does not end with the allowed suffix.
// The suffix must start with a dot (e.g. ".example.com").
// For development, if the suffix is "localhost:8080", it allows localhost requests.
func Middleware(allowedSuffix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Special case for development: allow localhost if suffix is set to localhost:8080
			if allowedSuffix == "localhost:8080" && (r.Host == "localhost:8080" || r.Host == "127.0.0.1:8080") {
				next.ServeHTTP(w, r)
				return
			}

			if !strings.HasSuffix(r.Host, allowedSuffix) {
				http.Error(w, "forbidden host", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
