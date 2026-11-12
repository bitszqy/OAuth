package variant

import (
	"os"
	"strings"
)

const (
	Fixed      = "fixed"
	Vulnerable = "vulnerable"
)

// Current returns the selected lab variant. The fixed implementation is the
// safe default when the environment variable is missing or unrecognized.
func Current() string {
	if strings.EqualFold(strings.TrimSpace(os.Getenv("LAB_VARIANT")), Vulnerable) {
		return Vulnerable
	}
	return Fixed
}

func IsVulnerable() bool {
	return Current() == Vulnerable
}
