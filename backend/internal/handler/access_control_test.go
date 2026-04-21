package handler

import "testing"

func TestNormalizeAccessToken(t *testing.T) {
	got := normalizeAccessToken("  demo-token \n")
	if got != "demo-token" {
		t.Fatalf("normalizeAccessToken trimmed value = %q", got)
	}
}

func TestHashAccessTokenDeterministic(t *testing.T) {
	a := hashAccessToken("same-token")
	b := hashAccessToken("same-token")
	c := hashAccessToken("other-token")

	if a == "" {
		t.Fatal("hashAccessToken returned empty hash")
	}
	if a != b {
		t.Fatal("hashAccessToken should be deterministic for the same input")
	}
	if a == c {
		t.Fatal("hashAccessToken should differ for different tokens")
	}
}
