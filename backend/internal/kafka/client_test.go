package kafka

import "testing"

func TestConfig_TLSEnabled_defaultSecure(t *testing.T) {
	// Zero value: TLS on (DisableTLS defaults false).
	var c Config
	if !c.TLSEnabled() {
		t.Fatal("expected TLS enabled by default")
	}
	c.DisableTLS = true
	if c.TLSEnabled() {
		t.Fatal("expected TLS disabled when DisableTLS is true")
	}
}
