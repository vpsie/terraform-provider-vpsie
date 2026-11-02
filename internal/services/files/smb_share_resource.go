package files

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

var (
	_ resource.Resource                   = &smbShareResource{}
	_ resource.ResourceWithConfigure      = &smbShareResource{}
	_ resource.ResourceWithImportState    = &smbShareResource{}
	_ resource.ResourceWithValidateConfig = &smbShareResource{}
)

type smbGrantModel struct {
	Identifier types.String `tfsdk:"identifier"`
	Mode       types.String `tfsdk:"mode"`
}

type smbShareModel struct {
	ShareIdentifier types.String    `tfsdk:"share_identifier"`
	Enabled         types.Bool      `tfsdk:"enabled"`
	Users           []smbGrantModel `tfsdk:"users"`
}

type smbShareResource struct {
	api *shareAPI
}

// NewSmbShareResource creates the vpsie_file_share_smb resource.
func NewSmbShareResource() resource.Resource {
	return &smbShareResource{}
}

func (r *smbShareResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_share_smb"
}

func (r *smbShareResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "SMB access of a file share: whether the share is served over SMB and which SMB users it grants " +
			"(`vpsie_file_smb_user`). The users must belong to the VPC the share is served in. An enabled share needs at " +
			"least one user. The platform reports whether SMB is on but not the granted users, so a change to the users " +
			"made outside Terraform is not detected. Destroying the resource turns SMB off and removes every grant. " +
			"Needs the `fileStorage` licence module.",
		Attributes: map[string]schema.Attribute{
			"share_identifier": schema.StringAttribute{
				MarkdownDescription: "The share identifier (`vpsie_file_share.<name>.identifier`). Changing it replaces the resource.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"enabled": schema.BoolAttribute{
				MarkdownDescription: "Whether the share is served over SMB.",
				Required:            true,
			},
			"users": schema.ListNestedAttribute{
				MarkdownDescription: "The SMB users the share grants: up to 64, each once.",
				Required:            true,
				Validators:          []validator.List{listvalidator.SizeAtMost(64)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							MarkdownDescription: "The SMB user identifier (`vpsie_file_smb_user.<name>.identifier`).",
							Required:            true,
						},
						"mode": schema.StringAttribute{
							MarkdownDescription: "`rw` (the default) or `ro`.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString("rw"),
							Validators:          []validator.String{stringvalidator.OneOf("rw", "ro")},
						},
					},
				},
			},
		},
	}
}

// ValidateConfig refuses what the platform refuses: an enabled share with no
// user, and one user granted twice.
func (r *smbShareResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var cfg smbShareModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}
	seen := map[string]bool{}
	for _, u := range cfg.Users {
		if u.Identifier.IsUnknown() || u.Identifier.IsNull() {
			continue
		}
		if seen[u.Identifier.ValueString()] {
			resp.Diagnostics.AddAttributeError(path.Root("users"), "Duplicate SMB user", fmt.Sprintf("SMB user %s is granted twice.", u.Identifier.ValueString()))
		}
		seen[u.Identifier.ValueString()] = true
	}
	if cfg.Enabled.ValueBool() && !cfg.Enabled.IsUnknown() && cfg.Users != nil && len(cfg.Users) == 0 {
		resp.Diagnostics.AddAttributeError(path.Root("users"), "No SMB user", "An enabled SMB share needs at least one user.")
	}
}

func (r *smbShareResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	api, err := rawapi.FromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Resource Configuration Type", err.Error()+" Please report this issue to the provider developers.")
		return
	}
	r.api = &shareAPI{api: api}
}

func grantsOf(models []smbGrantModel) []smbGrant {
	out := make([]smbGrant, 0, len(models))
	for _, m := range models {
		out = append(out, smbGrant{Identifier: m.Identifier.ValueString(), Mode: m.Mode.ValueString()})
	}
	return out
}

// apply sends the SMB switch and users, then waits until the gateway has the
// change (the share is ready again).
func (r *smbShareResource) apply(ctx context.Context, shareID string, enabled bool, users []smbGrantModel) error {
	if err := r.api.setShareSmb(ctx, shareID, smbShareRequest{Enabled: enabled, Users: grantsOf(users)}); err != nil {
		return err
	}
	_, err := r.api.waitReady(ctx, shareID)
	return err
}

func (r *smbShareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan smbShareModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, plan.ShareIdentifier.ValueString(), plan.Enabled.ValueBool(), plan.Users); err != nil {
		resp.Diagnostics.AddError("Error setting SMB on the file share", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *smbShareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state smbShareModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	s, err := r.api.get(ctx, state.ShareIdentifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading file share", err.Error())
		return
	}
	if s == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	if s.Protocols != nil {
		state.Enabled = types.BoolValue(s.Protocols.SMB)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *smbShareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan smbShareModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.apply(ctx, plan.ShareIdentifier.ValueString(), plan.Enabled.ValueBool(), plan.Users); err != nil {
		resp.Diagnostics.AddError("Error changing SMB on the file share", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *smbShareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state smbShareModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ShareIdentifier.ValueString()
	if err := r.apply(ctx, id, false, nil); err != nil {
		// A share that is already gone has no SMB left to turn off.
		if s, getErr := r.api.get(ctx, id); getErr == nil && s == nil {
			return
		}
		resp.Diagnostics.AddError("Error turning SMB off on the file share", err.Error())
	}
}

func (r *smbShareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("share_identifier"), req, resp)
}
