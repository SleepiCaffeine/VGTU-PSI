package api

import "net/http"

const (
	tenantHeader    = "X-Tenant-ID"
	tenantPathParam = "tenantId"
)

// requireTenantContext rejects requests whose path names a tenant other than the caller's.
// The header stands in for the authenticated identity until login exists.
func requireTenantContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pathTenant := r.PathValue(tenantPathParam)
		if pathTenant != "" && pathTenant != r.Header.Get(tenantHeader) {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		next.ServeHTTP(w, r)
	})
}
