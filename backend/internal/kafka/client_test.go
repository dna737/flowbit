package kafka

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfig_TLSEnabled_withCerts(t *testing.T) {
	// TLS is enabled when all certificate files are provided.
	c := Config{
		CertFile: "service.cert",
		KeyFile:  "service.key",
		CAFile:   "ca.pem",
	}
	if !c.TLSEnabled() {
		t.Fatal("expected TLS enabled when cert files are provided")
	}

	// TLS is disabled when any cert file is missing.
	c.CertFile = ""
	if c.TLSEnabled() {
		t.Fatal("expected TLS disabled when CertFile is empty")
	}
}

func TestNewWriter_rejectsNonPEMKey(t *testing.T) {
	dir := t.TempDir()
	certPath := filepath.Join(dir, "c.pem")
	keyPath := filepath.Join(dir, "k.pem")
	caPath := filepath.Join(dir, "ca.pem")
	if err := os.WriteFile(certPath, []byte("unused"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyPath, []byte("not-a-pem-key"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(caPath, []byte("unused"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := NewWriter(Config{
		Brokers:  []string{"localhost:9092"},
		Topic:    "t",
		CertFile: certPath,
		KeyFile:  keyPath,
		CAFile:   caPath,
	})
	if err == nil || !strings.Contains(err.Error(), "not a PEM private key") {
		t.Fatalf("got %v", err)
	}
}
