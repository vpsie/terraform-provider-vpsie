package baremetal

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// installInputDescription explains the replacement rule in the plan output.
const installInputDescription = "Changing this value orders a new server and releases the current one, " +
	"except when the value is set for the first time on a server adopted with terraform import."

// adoptedByImport reports whether the prior state belongs to a server adopted
// with `terraform import` and not applied since. plan_identifier is required,
// so it is null in state only when Read filled the state of an import: the
// API does not return the plan a server was ordered with.
func adoptedByImport(ctx context.Context, state tfsdk.State) bool {
	if state.Raw.IsNull() {
		return false
	}

	var plan types.String
	if diags := state.GetAttribute(ctx, path.Root("plan_identifier"), &plan); diags.HasError() {
		return false
	}

	return plan.IsNull()
}

// installInputNeedsReplace decides whether a change to an install-time input
// (plan, OS, SSH keys, user data, ...) needs a new server. The API cannot
// change these on a running server, so a change replaces it, with two
// exceptions:
//
//   - an Optional+Computed attribute left out of the configuration plans as
//     unknown until the API fills it; that is not a change;
//   - a server adopted by import has these inputs null in state because the
//     API does not return them; the first configured value is adopted as is.
func installInputNeedsReplace(ctx context.Context, state tfsdk.State, plan tfsdk.Plan, stateValue, planValue, configValue attr.Value) bool {
	if state.Raw.IsNull() || plan.Raw.IsNull() {
		// Create or destroy: nothing to replace.
		return false
	}
	if planValue.IsUnknown() && configValue.IsNull() {
		return false
	}
	if planValue.Equal(stateValue) {
		return false
	}
	if stateValue.IsNull() && adoptedByImport(ctx, state) {
		return false
	}

	return true
}

// installInput is the plan modifier of install-time inputs.
type installInput struct{}

// replaceInstallInput returns the plan modifier of install-time inputs; see
// installInputNeedsReplace.
func replaceInstallInput() installInput {
	return installInput{}
}

var (
	_ planmodifier.String = installInput{}
	_ planmodifier.Int64  = installInput{}
	_ planmodifier.Set    = installInput{}
	_ planmodifier.Object = installInput{}
)

func (installInput) Description(_ context.Context) string {
	return installInputDescription
}

func (m installInput) MarkdownDescription(ctx context.Context) string {
	return m.Description(ctx)
}

func (installInput) PlanModifyString(ctx context.Context, req planmodifier.StringRequest, resp *planmodifier.StringResponse) {
	resp.RequiresReplace = installInputNeedsReplace(ctx, req.State, req.Plan, req.StateValue, req.PlanValue, req.ConfigValue)
}

func (installInput) PlanModifyInt64(ctx context.Context, req planmodifier.Int64Request, resp *planmodifier.Int64Response) {
	resp.RequiresReplace = installInputNeedsReplace(ctx, req.State, req.Plan, req.StateValue, req.PlanValue, req.ConfigValue)
}

func (installInput) PlanModifySet(ctx context.Context, req planmodifier.SetRequest, resp *planmodifier.SetResponse) {
	resp.RequiresReplace = installInputNeedsReplace(ctx, req.State, req.Plan, req.StateValue, req.PlanValue, req.ConfigValue)
}

func (installInput) PlanModifyObject(ctx context.Context, req planmodifier.ObjectRequest, resp *planmodifier.ObjectResponse) {
	resp.RequiresReplace = installInputNeedsReplace(ctx, req.State, req.Plan, req.StateValue, req.PlanValue, req.ConfigValue)
}
