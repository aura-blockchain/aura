package walletsecurity

import (
	"testing"
)

func TestNewClientNilPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when creating walletsecurity Client with nil auraClient")
		}
	}()
	NewClient(nil)
}
