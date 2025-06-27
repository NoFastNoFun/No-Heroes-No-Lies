package hostfilter

import (
	"net/http"
	"strings"
)

// Middleware blocks any request whose Host header does not end with the allowed suffix.
// The suffix must start with a dot (e.g. ".example.com").
func Middleware(allowedSuffix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasSuffix(r.Host, allowedSuffix) {
				http.Error(w, "forbidden host", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
