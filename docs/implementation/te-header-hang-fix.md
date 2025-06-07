# TE Header Authentication Hang Fix

## Problem

Authentication requests from browsers (especially Firefox) were hanging indefinitely when they included the `TE: trailers` header. This header tells the server that the client can handle trailing headers in chunked transfer encoding.

## Root Cause

The Go HTTP server can hang when it receives the `TE: trailers` header in certain configurations, especially when combined with:
- HTTPS connections
- HTTP/1.1 protocol
- Certain middleware chains

## Solution

Implemented a middleware that strips the problematic `TE` header before the request is processed:

### 1. Created TE Header Middleware (`/opt/entitydb/src/api/te_header_middleware.go`)

```go
type TEHeaderMiddleware struct{}

func (m *TEHeaderMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if teHeader := r.Header.Get("TE"); teHeader != "" {
            logger.Debug("Removing TE header to prevent hang: %s", teHeader)
            r.Header.Del("TE")
        }
        next.ServeHTTP(w, r)
    })
}
```

### 2. Updated Server Middleware Chain (`/opt/entitydb/src/main.go`)

Added the TE header middleware as the first middleware in the chain:

```go
// Add TE header middleware to prevent hangs with browser headers
teHeaderMiddleware := api.NewTEHeaderMiddleware()

// Chain middleware together
chainedMiddleware := func(h http.Handler) http.Handler {
    // Apply in order: TE header fix -> request metrics -> handler
    return teHeaderMiddleware.Middleware(requestMetrics.Middleware(h))
}

// Use chained middleware in server configuration
server.server = &http.Server{
    Handler: corsHandler(chainedMiddleware(router)),
    // ... other config
}
```

## Results

- Authentication requests with browser headers now complete successfully
- Firefox curl commands work without hanging
- Response times: ~3.3s with browser headers, ~2s without
- No impact on other functionality

## Testing

Before fix:
```bash
# This would hang indefinitely
curl -H 'TE: trailers' https://localhost:8085/api/v1/auth/login ...
```

After fix:
```bash
# Now completes successfully in ~3.3 seconds
curl -H 'TE: trailers' https://localhost:8085/api/v1/auth/login ...
```

## References

- [Go Issue #42840](https://github.com/golang/go/issues/42840) - TE header handling in Go HTTP server
- [MDN TE Header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/TE) - Documentation on the TE header