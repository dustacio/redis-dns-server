package main

import (
	"testing"
)

func TestWildCardHostName(t *testing.T) {
	given := "server.domain.local."

	wcName := WildCardHostName(given)
	if wcName != "*.domain.local." {
		t.Error("Expected *.server.domain.local., got ", wcName)
	}
}
