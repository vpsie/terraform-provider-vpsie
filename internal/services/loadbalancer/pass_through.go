package loadbalancer

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

const passThroughDeprecation = "pass_through is no longer supported: TLS passthrough never produced a working " +
	"load balancer configuration and the API refuses true. Remove the attribute; set backend_scheme = \"https\" " +
	"to encrypt the traffic to the backends."

// passThroughMustBeFalse refuses pass_through = true at plan time, before the API answers 400.
type passThroughMustBeFalse struct{}

func (passThroughMustBeFalse) Description(_ context.Context) string {
	return "must be false: TLS passthrough is not supported"
}

func (v passThroughMustBeFalse) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (passThroughMustBeFalse) ValidateBool(_ context.Context, req validator.BoolRequest, resp *validator.BoolResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() || !req.ConfigValue.ValueBool() {
		return
	}
	resp.Diagnostics.AddAttributeError(req.Path, "TLS passthrough is not supported", passThroughDeprecation)
}
