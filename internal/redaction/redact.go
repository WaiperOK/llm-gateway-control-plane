package redaction

import "regexp"

var (
	emailRE = regexp.MustCompile(`(?i)\b[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}\b`)
	phoneRE = regexp.MustCompile(`\b(?:\+?\d{1,3}[-.\s]?)?(?:\(?\d{3}\)?[-.\s]?\d{3}[-.\s]?\d{4})\b`)
	ipRE    = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
)

// Result is a scrubbed text output.
type Result struct {
	Text  string
	Count int
}

// Scrub redacts common high-risk fields from logs/audit data.
func Scrub(s string) Result {
	count := 0
	replace := func(re *regexp.Regexp, marker string, src string) string {
		indices := re.FindAllStringIndex(src, -1)
		count += len(indices)
		return re.ReplaceAllString(src, marker)
	}
	out := s
	out = replace(emailRE, "[REDACTED_EMAIL]", out)
	out = replace(phoneRE, "[REDACTED_PHONE]", out)
	out = replace(ipRE, "[REDACTED_IP]", out)
	return Result{Text: out, Count: count}
}
