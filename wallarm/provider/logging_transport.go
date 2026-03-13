package wallarm

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// loggingTransport is an http.RoundTripper that logs HTTP request/response
// in SDK-style format via log.Printf, captured by TF_LOG.
type loggingTransport struct {
	transport http.RoundTripper
}

func newLoggingTransport(transport http.RoundTripper) *loggingTransport {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &loggingTransport{transport: transport}
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var reqBody []byte
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
	}
	t.logRequest(req, reqBody)

	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		log.Printf("[DEBUG] ---[ REQUEST ERROR ]---\n%s %s: %s\n---[ END REQUEST ERROR ]---", req.Method, req.URL, err)
		return resp, err
	}

	respBody, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewReader(respBody))
	t.logResponse(resp, respBody)

	return resp, nil
}

func (t *loggingTransport) logRequest(req *http.Request, body []byte) {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("---[ REQUEST ]---\n%s %s %s\nHost: %s\n",
		req.Method, req.URL.RequestURI(), req.Proto, req.URL.Host))
	for key, vals := range req.Header {
		for _, val := range vals {
			if isSensitiveHTTPHeader(key) {
				val = maskHTTPValue(val)
			}
			buf.WriteString(fmt.Sprintf("%s: %s\n", key, val))
		}
	}
	if len(body) > 0 {
		buf.WriteString(fmt.Sprintf("\n%s\n", string(body)))
	}
	buf.WriteString("---[ END REQUEST ]---")
	log.Printf("[DEBUG] %s", buf.String())
}

func (t *loggingTransport) logResponse(resp *http.Response, body []byte) {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("---[ RESPONSE ]---\n%s %s\n", resp.Proto, resp.Status))
	for key, vals := range resp.Header {
		for _, val := range vals {
			buf.WriteString(fmt.Sprintf("%s: %s\n", key, val))
		}
	}
	if len(body) > 0 {
		buf.WriteString(fmt.Sprintf("\n%s\n", string(body)))
	}
	buf.WriteString("---[ END RESPONSE ]---")
	log.Printf("[DEBUG] %s", buf.String())
}

func isSensitiveHTTPHeader(key string) bool {
	lower := strings.ToLower(key)
	return strings.Contains(lower, "token") ||
		strings.Contains(lower, "secret") ||
		strings.Contains(lower, "authorization")
}

func maskHTTPValue(val string) string {
	if len(val) <= 4 {
		return "****"
	}
	return strings.Repeat("*", len(val)-4) + val[len(val)-4:]
}
