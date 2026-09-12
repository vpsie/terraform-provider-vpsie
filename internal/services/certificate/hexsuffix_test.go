package certificate

import "testing"

func TestHasHexSuffix(t *testing.T) {
	cases := []struct {
		candidate string
		name      string
		want      bool
	}{
		{"my-cert-4cb5", "my-cert", true},      // API-appended hex suffix
		{"my-cert-4CB5", "my-cert", true},      // uppercase hex
		{"my-cert-prod", "my-cert", false},     // non-hex suffix must not match
		{"my-cert-", "my-cert", false},         // empty suffix
		{"my-cert", "my-cert", false},          // exact name is not a suffix match
		{"other-4cb5", "my-cert", false},       // different name
		{"my-cert-web-4cb5", "my-cert", false}, // suffix contains non-hex
	}

	for _, c := range cases {
		if got := hasHexSuffix(c.candidate, c.name); got != c.want {
			t.Errorf("hasHexSuffix(%q, %q) = %v, want %v", c.candidate, c.name, got, c.want)
		}
	}
}
