package httputil

import (
	"crypto/tls"
	"net/http"
)

// NewSecureTransport returns a clone of http.DefaultTransport
// with TLS minimum version set to TLS 1.2.
func NewSecureTransport() *http.Transport {
	t := http.DefaultTransport.(*http.Transport).Clone()
	if t.TLSClientConfig == nil {
		t.TLSClientConfig = &tls.Config{}
	}
	t.TLSClientConfig.MinVersion = tls.VersionTLS12
	return t
}
