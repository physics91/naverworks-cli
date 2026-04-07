package httputil

import (
	"crypto/tls"
	"testing"
)

func TestNewSecureTransport_MinTLSVersion(t *testing.T) {
	tr := NewSecureTransport()
	if tr.TLSClientConfig == nil {
		t.Fatal("TLSClientConfig is nil")
	}
	if tr.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("expected MinVersion TLS 1.2 (%d), got %d", tls.VersionTLS12, tr.TLSClientConfig.MinVersion)
	}
}

func TestNewSecureTransport_PreservesDefaults(t *testing.T) {
	tr := NewSecureTransport()
	if tr.MaxIdleConns == 0 {
		t.Error("expected non-zero MaxIdleConns from default transport")
	}
	if tr.IdleConnTimeout == 0 {
		t.Error("expected non-zero IdleConnTimeout from default transport")
	}
}
