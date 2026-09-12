package domain

import "testing"

func TestNormalizeRecordName(t *testing.T) {
	const dom = "example.com"
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"short name", "www", "www.example.com"},
		{"apex at", "@", "example.com"},
		{"wildcard", "*", "*.example.com"},
		{"already qualified", "www.example.com", "www.example.com"},
		{"qualified different case", "WWW.Example.COM", "WWW.Example.COM"},
		{"equal to domain", "example.com", "example.com"},
		{"nested subdomain", "a.b", "a.b.example.com"},
		{"partial suffix is not a match", "notexample.com", "notexample.com.example.com"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeRecordName(tc.in, dom); got != tc.want {
				t.Errorf("normalizeRecordName(%q, %q) = %q, want %q", tc.in, dom, got, tc.want)
			}
		})
	}
}
