package baremetal

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const (
	passwordMinLength = 8
	passwordMaxLength = 128
)

// hostnamePattern matches a DNS hostname of letters, digits, dots and hyphens
// that starts and ends with a letter or digit, as the API's hostname rule does.
var hostnamePattern = regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9.-]{0,251}[A-Za-z0-9])?$`)

// passwordValidator enforces the API's rule for a bare-metal administrator
// password: 8 to 128 printable ASCII characters, without double quotes,
// backslashes or double curly braces. Checking it at plan time keeps an order
// from failing after Terraform has started it.
type passwordValidator struct{}

var _ validator.String = passwordValidator{}

func (passwordValidator) Description(_ context.Context) string {
	return fmt.Sprintf("must be %d to %d printable ASCII characters, without double quotes, backslashes or double curly braces",
		passwordMinLength, passwordMaxLength)
}

func (v passwordValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v passwordValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	if problem := passwordProblem(req.ConfigValue.ValueString()); problem != "" {
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid bare-metal password",
			fmt.Sprintf("The password %s: %s.", v.Description(ctx), problem))
	}
}

// passwordProblem names what makes password unacceptable, or returns "".
func passwordProblem(password string) string {
	if n := len(password); n < passwordMinLength || n > passwordMaxLength {
		return fmt.Sprintf("it has %d characters", n)
	}
	for _, r := range password {
		if r < 0x20 || r > 0x7e {
			return "it contains a character outside printable ASCII"
		}
		if r == '"' || r == '\\' {
			return fmt.Sprintf("it contains %q", r)
		}
	}
	if strings.Contains(password, "{{") || strings.Contains(password, "}}") {
		return "it contains double curly braces"
	}

	return ""
}

// userDataValidator enforces the API's rule for first-boot user data: a
// cloud-config document or a script, at most 64 KB, without NUL bytes.
type userDataValidator struct{}

var _ validator.String = userDataValidator{}

const userDataMaxBytes = 64 * 1024

func (userDataValidator) Description(_ context.Context) string {
	return `must start with "#cloud-config" or "#!", be at most 64 KB and contain no NUL byte`
}

func (v userDataValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v userDataValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	data := req.ConfigValue.ValueString()
	switch {
	case !strings.HasPrefix(data, "#cloud-config") && !strings.HasPrefix(data, "#!"):
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid user data", "User data "+v.Description(ctx)+": it starts with neither.")
	case len(data) > userDataMaxBytes:
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid user data",
			fmt.Sprintf("User data %s: it has %d bytes.", v.Description(ctx), len(data)))
	case strings.ContainsRune(data, 0):
		resp.Diagnostics.AddAttributeError(req.Path, "Invalid user data", "User data "+v.Description(ctx)+": it contains a NUL byte.")
	}
}
