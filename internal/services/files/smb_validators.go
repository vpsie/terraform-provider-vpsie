package files

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// vpcNameValidator checks how an SMB user names its VPC: dc_identifier with
// vpc_id, or share_identifier alone.
type vpcNameValidator struct{}

func exactlyOneVPCName() resource.ConfigValidator { return vpcNameValidator{} }

func (vpcNameValidator) Description(context.Context) string {
	return "name the VPC with dc_identifier and vpc_id, or with share_identifier"
}

func (v vpcNameValidator) MarkdownDescription(ctx context.Context) string { return v.Description(ctx) }

func (vpcNameValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var dc, share types.String
	var vpc types.Int64
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("dc_identifier"), &dc)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("vpc_id"), &vpc)...)
	resp.Diagnostics.Append(req.Config.GetAttribute(ctx, path.Root("share_identifier"), &share)...)
	if resp.Diagnostics.HasError() || dc.IsUnknown() || vpc.IsUnknown() || share.IsUnknown() {
		return
	}
	byShare := !share.IsNull()
	byVPC := !dc.IsNull() || !vpc.IsNull()
	switch {
	case byShare && byVPC:
		resp.Diagnostics.AddError("Conflicting VPC", "Set share_identifier, or dc_identifier with vpc_id, not both.")
	case !byShare && (dc.IsNull() || vpc.IsNull()):
		resp.Diagnostics.AddError("Missing VPC", "Set dc_identifier together with vpc_id, or set share_identifier.")
	}
}
