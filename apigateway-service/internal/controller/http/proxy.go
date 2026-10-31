package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"apigateway-service/pkg/clients"
	sharedhttp "shared/http"
)

// hopByHopHeaders must not be forwarded by a proxy (RFC 7230, section 6.1).
var hopByHopHeaders = map[string]struct{}{
	"Connection":          {},
	"Keep-Alive":          {},
	"Proxy-Authenticate":  {},
	"Proxy-Authorization": {},
	"Te":                  {},
	"Trailer":             {},
	"Transfer-Encoding":   {},
	"Upgrade":             {},
}

func copyResponseHeaders(dst, src http.Header) {
	for key, values := range src {
		if _, skip := hopByHopHeaders[http.CanonicalHeaderKey(key)]; skip {
			continue
		}
		for _, v := range values {
			dst.Add(key, v)
		}
	}
}

func writeBadGateway(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadGateway)
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "upstream service unavailable"})
}

// proxyRequest forwards the incoming request to serviceURL+endpoint and streams
// the upstream response (status, headers and body) back to the client.
//
// The query string of the incoming request is appended automatically unless
// endpoint already contains one (as happens with r.URL.RequestURI()).
func proxyRequest(w http.ResponseWriter, r *http.Request, serviceURL string, cli *clients.HTTPClient, endpoint string) {
	if r.URL.RawQuery != "" && !strings.Contains(endpoint, "?") {
		endpoint += "?" + r.URL.RawQuery
	}
	url := serviceURL + endpoint

	req, err := http.NewRequestWithContext(r.Context(), r.Method, url, r.Body)
	if err != nil {
		sharedhttp.ResponseWithError(w, r, http.StatusInternalServerError, err)
		return
	}

	req.Header = r.Header.Clone()
	req.ContentLength = r.ContentLength

	resp, err := cli.Do(req)
	if err != nil {
		writeBadGateway(w)
		return
	}
	defer resp.Body.Close()

	copyResponseHeaders(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}
