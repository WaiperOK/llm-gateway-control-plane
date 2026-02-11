package redaction

import "testing"

func TestScrubRedactsEmailPhoneAndIP(t *testing.T) {
	in := "Contact john@example.com or +1 (555) 123-4567 from 10.0.0.2"
	res := Scrub(in)
	if res.Count < 3 {
		t.Fatalf("expected at least 3 redactions, got %d", res.Count)
	}
	if res.Text == in {
		t.Fatal("expected text to be redacted")
	}
}
